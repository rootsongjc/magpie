// Copyright © 2017 Jimmy Song
//
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
	"github.com/rootsongjc/magpie/docker"
	"github.com/rootsongjc/magpie/tool"
	"github.com/rootsongjc/magpie/yarn"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var distribute bool
var view bool
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get yarn cluster status",
	Long:  "Get the resource usage, node status of yarn cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		cluster_names := viper.GetStringSlice("clusters.cluster_name")
		if distribute == false && view == false {
			if clustername == "" {
				yarn.Get_yarn_status(cluster_names)
			} else {
				names := []string{clustername}
				yarn.Get_yarn_status(names)
			}
			if clustername == "" {
				docker.Get_docker_status(cluster_names)
			} else {
				names := []string{clustername}
				docker.Get_docker_status(names)
			}
			fmt.Println("============NODEMANAGER AND DOCKER CONTAINERS COMPARATION===========")
			fmt.Println("CLUSTER\tNODEMANAGER\tCONTAINER\tRESULT")
			if clustername == "" {
				for _, c := range cluster_names {
					tool.Compare_yarn_docker_cluster(c)
				}
			} else {
				tool.Compare_yarn_docker_cluster(clustername)
			}
		}
		if distribute == true {
			yarn.Yarn_distribution(clustername)
		}

		if view == true {
			if clustername == "" {
				fmt.Println("You must specify the clusteranem with -c")
			} else {
				yarn.Yarn_view(clustername)
			}
		}

	},
}

func init() {
	yarnCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVarP(&clustername, "clustername", "c", "", "Yarn cluster name")
	statusCmd.Flags().BoolVarP(&distribute, "distribute", "d", false, "Yarn nodemanager distribution")
	statusCmd.Flags().BoolVarP(&view, "view", "v", false, "Show the yarn cluster containers view")

}
