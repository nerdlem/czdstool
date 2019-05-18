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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"lem.click/czdstool/czds"
)

var cfgFile string
var verbose bool
var s = czds.NewSession()

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "czdstool",
	Short: "Download ICANN CZDS zone files",
	Long: `Utility program to use the ICANN CZDS REST API to download authorized
TLD zone files.`,
	Run: func(cmd *cobra.Command, args []string) {
		authenticate(s)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/czds.toml)")
	RootCmd.Flags().StringVarP(&tokenFile, "auth-file", "A", "", "auth file previously created with save")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output format")

	RootCmd.AddCommand(saveCmd)

	RootCmd.AddCommand(lsCmd)
	// tokenFile from save
	lsCmd.Flags().StringVarP(&tokenFile, "auth-file", "A", "", "auth file previously created with save")
	lsCmd.Flags().BoolVarP(&zonesOnly, "zones-only", "z", false, "list TLD rather than zone URL")
	lsCmd.Flags().BoolVarP(&details, "details", "l", false, "list zones and details")
	lsCmd.Flags().IntVarP(&infoJobs, "info-workers", "n", 4, "number of parallel jobs")

	viper.BindPFlag("api.info_workers", lsCmd.Flags().Lookup("info-workers"))

	RootCmd.AddCommand(fetchCmd)
	// tokenFile from save
	fetchCmd.Flags().StringVarP(&tokenFile, "auth-file", "A", "", "auth file previously created with save")
	fetchCmd.Flags().IntVarP(&infoJobs, "info-workers", "n", 2, "number of info parallel jobs")
	fetchCmd.Flags().IntVarP(&fetchJobs, "fetch-workers", "N", 4, "number of fetch parallel jobs")
	fetchCmd.Flags().BoolVarP(&forceFetch, "force", "F", false, "fetch regardless of file modification")
	fetchCmd.Flags().StringVarP(&destinationDir, "destination-dir", "D", "./", "destination for fetched TLD files")
	fetchCmd.Flags().BoolVarP(&keepAnyway, "keep-anyway", "K", false, "keeps downloaded zone if size mismatch vs API")

	viper.BindPFlag("api.info_workers", fetchCmd.Flags().Lookup("info-workers"))
	viper.BindPFlag("api.fetch_workers", fetchCmd.Flags().Lookup("fetch-workers"))
	viper.BindPFlag("download.destination_dir", fetchCmd.Flags().Lookup("destination-dir"))

	viper.SetDefault("api.fetch_workers", "4")
	viper.SetDefault("download.destination_dir", "./")

	cobra.MarkFlagFilename(RootCmd.Flags(), "auth-file")
	cobra.MarkFlagFilename(lsCmd.Flags(), "auth-file")
	cobra.MarkFlagFilename(fetchCmd.Flags(), "destination-dir")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc")
		viper.AddConfigPath(".")
		viper.SetConfigName("czds")
	}

	viper.SetEnvPrefix("czds")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()

	// Try to be smart about errors
	if err != nil {
		_, fnfe := err.(viper.ConfigFileNotFoundError)
		if cfgFile != "" || !fnfe {
			panic(fmt.Errorf("invalid config file %s: %s", viper.ConfigFileUsed(), err))
		}
	}
}
