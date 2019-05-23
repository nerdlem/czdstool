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

	"github.com/nerdlem/czdstool/czds"
	"github.com/spf13/viper"
)

// Worker represents the signature of the functions we'll use to perform
// parallel work.
type Worker func(interface{})

// Existing represents a TLD zone file as perhaps present in the filesystem
type Existing struct {
	// The URL from the ICANN CZDS REST API for the TLD file
	URL string
	// FileName is the absolute path of the file to be downloaded
	FileName string
	// Length is the size of the actual zone file to download
	Length int64
}

func launchParallelWorkers(w Worker, n int) chan interface{} {

	c := make(chan interface{}, 100)

	for i := 1; i <= n; i++ {
		go func() {
			for s := range c {
				w(s)
			}
		}()
	}

	return c
}

func fileFromURL(u string) string {
	tld := czds.TLDFromURL(u)
	return fmt.Sprintf("%s/%s.zone.gz", viper.GetString("download.destination_dir"), tld)
}

func authenticate(sess *czds.Sess) {
	if tokenFile != "" {
		if verbose {
			fmt.Fprintf(os.Stderr, "Reading authorization token from %s\n", tokenFile)
		}

		if err := sess.Fetch(tokenFile); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else {
		if verbose {
			fmt.Fprintf(os.Stderr, "Requesting auth token using credentials\n")
		}

		if err := sess.Login(viper.GetString("auth.username"), viper.GetString("auth.password")); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Request is authorized\n")
	}

}
