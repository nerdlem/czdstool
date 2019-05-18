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
	"github.com/spf13/viper"
)

var tokenFile string

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Persist authorization ICANN CZDS authorization token for later reuse",
	Long: `This command performs an authentication to the ICANN CZDS using
the credentials taken from the config file. If authentication is successful,
the obtained JWT token is saved into the named file, which is to be provided
as the only command line argument.

The file is not overridden in case of failure.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tokenFile = args[0]
		if verbose {
			fmt.Printf("Would save token in file %s\n", tokenFile)
		}

		if err := s.Login(viper.GetString("auth.username"), viper.GetString("auth.password")); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		if err := s.Store(tokenFile, 0440); err != nil {
			fmt.Fprintf(os.Stderr, "unable to save session to %s: %s\n", tokenFile, err)
			os.Exit(2)
		}

		os.Exit(0)
	},
}
