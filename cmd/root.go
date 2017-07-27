// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "consuldog",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/consuldog.yaml)")
	RootCmd.PersistentFlags().StringP("templateFolder", "t", "/etc/dd-agent/conf.d/auto_conf", "the folder containing our datadog templates")
	RootCmd.PersistentFlags().StringP("datadogFolder", "d", "/etc/dd-agent", "the base datadog config folder (the one containing the datadog.conf file)")
	RootCmd.PersistentFlags().StringP("datadogProcName", "k", "supervisord", "the name of the datadog process we should send reload signals to.  A process with this name, that is running as the same user as consuldog (if one can be found) will be sent a HUP signal")
	RootCmd.PersistentFlags().StringP("prefix", "p", "consuldogConfig:", "the consul tag prefix to look for in consul to know that a service needs monitoring")
	RootCmd.PersistentFlags().StringP("consulAddress", "a", "http://localhost:8500", "the address of the consul server")
	RootCmd.PersistentFlags().Int64P("datadogMinReloadInterval", "m", 10, "the minimum number of seconds between reloads of the DataDog process regardless of how many times the configs are updated in that time.")
	RootCmd.PersistentFlags().StringSliceP("nodeName", "n", []string{}, "the name of the node we want to look at the services of")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)

		// Search config in home directory with name ".consuldog" (without extension).
		viper.AddConfigPath("/etc/")
		viper.SetConfigName("consuldog.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match
	// bind all of our flags so we can access them with viper
	viper.BindPFlags(RootCmd.Flags())

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// we have to set this here since it uses other values
	viper.SetDefault("templateFolder", path.Join(viper.GetString("datadogFolder"), "/conf.d/auto_conf"))
}
