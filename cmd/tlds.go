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

	"encoding/json"

	"github.com/spf13/cobra"
)

var tJSON bool

// tldsCmd represents the tlds command
var tldsCmd = &cobra.Command{
	Use:   "tlds",
	Short: "List TLD data from ICANN CZDS",
	Long: `Use the ICANN CZDS REST API to fetch and return a list of
tlds along with various details.`,
	Run: func(cmd *cobra.Command, args []string) {
		authenticate(s)
		m, err := s.GetTLDsMetadata()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot fetch TLD metadata: %s", err)
			os.Exit(1)
		}

		for n, meta := range m {
			if tJSON {
				if bytes, err := json.Marshal(meta); err != nil {
					fmt.Fprintf(os.Stderr, "%s: %s\n", n, err)
				} else {
					fmt.Printf("%s\n", string(bytes))
				}
				continue
			}

			fmt.Printf("%s\n", meta.String())
		}

		os.Exit(0)
	},
}
