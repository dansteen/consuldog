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
	"log"

	"github.com/dansteen/consuldog/communicator"
	"github.com/dansteen/consuldog/datadog"
	"github.com/dansteen/consuldog/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch consul and act on new services",

	Long: `Watch services in consul`,
	Run:  watch,
}

func init() {
	RootCmd.AddCommand(watchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// watchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// watchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func watch(cmd *cobra.Command, args []string) {
	client := communicator.NewConsulClient(viper.GetString("consulAddress"))
	newServices := make(chan services.NodeServices, 5)
	stop := make(chan bool)
	// we need to gather our services by node
	allServices := services.NewServices()

	// set up some chans and run our reloader so we can reload datadog when needed
	triggerReload := make(chan bool)
	go datadog.Reloader(triggerReload, stop)

	// run a thread for each onode we are monitoring
	for _, node := range viper.GetStringSlice("nodeName") {
		go client.MonitorNode(node, newServices, stop)
	}
	// listen for new services
	for {
		select {
		case nodeServices := <-newServices:
			// first clear existing services for this node from our service record
			allServices.ClearNode(nodeServices.Node)
			// then add in the new services for this node
			log.Println(nodeServices.Node)
			for _, service := range nodeServices.Services {
				log.Printf("%+v:%+v\n", service.ConfigTemplate, service.DatadogType)
				log.Printf("%+v:%+v\n\n", service.Address, service.Port)
				allServices.Add(service)
			}
			datadog.WriteConfig(allServices)
			triggerReload <- true
		}
	}
}
