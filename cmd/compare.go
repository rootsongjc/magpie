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
	"github.com/rootsongjc/magpie/tool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var yarncluster string

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare the yarn clusters with docker cluster.",
	Long:  "Compare the yarn nodemanagers with docker cluster containers.",
	Run: func(cmd *cobra.Command, args []string) {
		if yarncluster == "" {
			cluster_names := viper.GetStringSlice("clusters.cluster_name")
			for i := range cluster_names {
				tool.Compare_yarn_docker_cluster(cluster_names[i])
			}
		} else {
			tool.Compare_yarn_docker_cluster(yarncluster)
		}
	},
}

func init() {
	toolCmd.AddCommand(compareCmd)

	compareCmd.Flags().StringVarP(&yarncluster, "clustername", "c", "", "Yarn cluster name")

}
