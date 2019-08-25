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

	"github.com/spf13/cobra"
)

// requestCmd represents the request command
var requestCmd = &cobra.Command{
	Use:   "request",
	Short: "Request access to TLD zones via the ICANN CZDS",
	Long: `Use the ICANN CZDS REST API to request access to one or more zone files.
TLDs can be provided as command line arguments. Alternatively, use flags to
perform on appropriate TLDs based on status.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		authenticate(s)

		if err := s.RequestAccess(args); err != nil {
			fmt.Fprintf(os.Stderr, "request returned an error: %s\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	},
}
