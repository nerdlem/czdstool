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
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/nerdlem/czdstool/czds"
)

var details, zonesOnly bool
var infoJobs int
var wg sync.WaitGroup

func lookupAndOutput(i interface{}) {
	n := i.(string)
	if zonesOnly {
		fmt.Println(czds.TLDFromURL(n))
		wg.Done()
		return
	}

	if details {
		if verbose {
			fmt.Fprintf(os.Stderr, "fetching details of %s\n", n)
		}
		if det, err := s.Details(n); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", n, err)
		} else {
			fmt.Println(det)
		}
		wg.Done()
		return
	}

	fmt.Println(n)
	wg.Done()
	return
}

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List available TLD zone URLs in the ICANN CZDS",
	Long: `Use the ICANN CZDS REST API to fetch and return a list of
available zone file URLs for downloading.`,
	Run: func(cmd *cobra.Command, args []string) {

		viper.SetDefault("api.info_workers", "4")

		authenticate(s)

		var list *[]string
		var err error

		ch := launchParallelWorkers(lookupAndOutput, viper.GetInt("api.info_workers"))

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
			ch <- n
		}

		wg.Wait()
		os.Exit(0)
	},
}
