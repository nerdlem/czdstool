package cmd

// Copyright © 2018 Luis E. Muñoz <github@lem.click>
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/nerdlem/czdstool/czds"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var destinationDir string
var forceFetch, keepAnyway bool
var fetchJobs int

var fc chan interface{}

func lookupAndCheck(i interface{}) {
	u := i.(string)

	if verbose {
		fmt.Fprintf(os.Stderr, "looking up data for %s\n", u)
	}

	det, err := s.Details(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", u, err)
		wg.Done()
		return
	}

	e := Existing{
		URL:       u,
		FileName:  fileFromURL(u),
		APILength: int64(det.ContentLength),
	}

	stat, err := os.Stat(e.FileName)
	if err == nil || err == os.ErrNotExist {
		if err == nil {
			if !det.LastModified.After(stat.ModTime()) {
				if verbose {
					fmt.Fprintf(os.Stderr, "skip %s - our %s is up to date\n", u, e.FileName)
					wg.Done()
					return
				}
			}

			if !keepAnyway {
				e.Length = stat.Size()
			} else {
				e.Length = -1
			}
		}
	} else {
		if verbose {
			fmt.Fprintf(os.Stderr, "stat() %s: %s\n", e.FileName, err)
		}
	}
	// We should do a wg.Add(1) and a wg.Done(). Those would cancel each
	// other, so we perform none.
	fc <- e
}

func fetchZone(i interface{}) {
	var copied int64
	var err error
	var reader *io.ReadCloser
	var fh *os.File

	e := i.(Existing)
	defer wg.Done()

	if verbose {
		start := time.Now()
		defer func(e *Existing, s time.Time) {
			fmt.Fprintf(os.Stderr, "fetching of %s for %d bytes took %s\n",
				e.FileName, copied, time.Now().Sub(s).String())
		}(&e, start)
	}

	tmpFile := fmt.Sprintf("%s.tmp", e.FileName)

	if verbose {
		fmt.Fprintf(os.Stderr, "will fetch %s via %s for %d bytes\n", e.FileName, e.URL, e.APILength)
	}

	reader, err = s.Download(e.URL)
	defer (*reader).Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	fh, err = os.Create(tmpFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file create %s: %s\n", e.FileName, err)
		return
	}

	copied, err = io.Copy(fh, io.Reader(*reader))

	if err != nil {
		fmt.Fprintf(os.Stderr, "writing zone file from %s to %s (%d bytes out of %d): %s\n",
			e.URL, tmpFile, copied, e.APILength, err)
		os.Remove(tmpFile)
		return
	}

	if e.Length > 0 && copied != e.APILength {
		fmt.Fprintf(os.Stderr, "API reported length of %s as %d but wrote %d bytes\n",
			e.URL, e.Length, copied)
		os.Remove(tmpFile)
		return
	}

	err = os.Rename(tmpFile, e.FileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "moving %s to %s: %s\n", tmpFile, e.FileName, err)
		os.Remove(tmpFile)
	}

	return
}

// fetchCmd represents the ls command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download TLD zone files using the ICANN CZDS REST API",
	Long: `Use the ICANN CZDS REST API to fetch the TLD zone files
provided via the ICANN CZDS REST API.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("api.info_workers", "2")

		authenticate(s)

		var list *[]string
		var err error
		var started time.Time

		if verbose {
			started = time.Now()
			fmt.Fprintf(os.Stderr, "beginning fetch process\n")
		}

		sc := launchParallelWorkers(lookupAndCheck, viper.GetInt("api.info_workers"))
		fc = launchParallelWorkers(fetchZone, viper.GetInt("api.fetch_workers"))

		if len(args) == 0 {

			if verbose {
				fmt.Fprintf(os.Stderr, "fetching list of TLD URLs\n")
			}

			list, err = s.List()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}

			if list == nil {
				fmt.Fprintf(os.Stderr, "nil list returned\n")
				os.Exit(2)
			}
		} else {
			if verbose {
				fmt.Fprintf(os.Stderr, "processing list of TLDs provided\n")
			}

			l := make([]string, len(args))
			for i, t := range args {
				l[i] = czds.URLFromTLD(t)
			}

			list = &l
		}

		for _, n := range *list {
			wg.Add(1)
			sc <- n
		}

		wg.Wait()

		if verbose {
			fmt.Fprintf(os.Stderr, "fetch process took %s\n", time.Now().Sub(started).String())
		}

		os.Exit(0)
	},
}
