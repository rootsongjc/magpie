// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"github.com/rootsongjc/magpie/yarn"
	"github.com/spf13/cobra"
)

//nodemanager container config
var (
	YARN_ZK_DIR           string
	YARN_CLUSTER_ID       string
	YARN_RM1_IP           string
	YARN_RM2_IP           string
	YARN_JOBHISTORY_IP    string
	CPU_CORE_NUM          string
	NODEMANAGER_MEMORY_MB string
	network_mode          string
	limit_cpus            int64
	limit_memory_mb       int64
	image                 string
)

var container_name string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new nodemanager for other yarn cluster",
	Long:  "Create a new nodemanager for other yarn cluster",
	Run: func(cmd *cobra.Command, args []string) {
		nodemanager := yarn.Nodemanager_config{
			YARN_ZK_DIR:           YARN_ZK_DIR,
			YARN_CLUSTER_ID:       YARN_CLUSTER_ID,
			YARN_RM1_IP:           YARN_RM1_IP,
			YARN_RM2_IP:           YARN_RM2_IP,
			YARN_JOBHISTORY_IP:    YARN_JOBHISTORY_IP,
			Network_mode:          network_mode,
			CPU_CORE_NUM:          CPU_CORE_NUM,
			NODEMANAGER_MEMORY_MB: NODEMANAGER_MEMORY_MB,
			Limit_cpus:            limit_cpus,
			Limit_memory_mb:       limit_memory_mb,
			Image:                 image,
			Container_name:        container_name,
		}
		if YARN_JOBHISTORY_IP == "" || YARN_RM1_IP == "" || YARN_RM2_IP == "" || YARN_CLUSTER_ID == "" || YARN_ZK_DIR == "" {
			fmt.Println("Lack of required parameters.\nUse -h for help.")
			return
		}
		yarn.Create_new_nodemanager(nodemanager)
	},
}

func init() {
	yarnCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&YARN_ZK_DIR, "zkdir", "", "", "Zookeeper dir ID (Required)")
	createCmd.Flags().StringVarP(&YARN_CLUSTER_ID, "id", "", "", "Yarn cluster ID (Required)")
	createCmd.Flags().StringVarP(&YARN_RM1_IP, "rm1", "", "", "Yarn actrive resourcemanager address (Required)")
	createCmd.Flags().StringVarP(&YARN_RM2_IP, "rm2", "", "", "Yarn standby resourcemanager address (Required)")
	createCmd.Flags().StringVarP(&YARN_JOBHISTORY_IP, "jobhistory", "", "", "Yarn jobhistory address (Required)")
	createCmd.Flags().StringVarP(&network_mode, "network", "", "mynet", "Docker network mode")
	createCmd.Flags().StringVarP(&image, "image", "", "", "Nodemanager docker image")
	createCmd.Flags().StringVarP(&container_name, "name", "", "", "Docker container name")
	createCmd.Flags().StringVarP(&CPU_CORE_NUM, "cpu", "", "4", "Nodemanager CPU")
	createCmd.Flags().StringVarP(&NODEMANAGER_MEMORY_MB, "memory", "", "10240", "Nodemanager memory MB")
	createCmd.Flags().Int64VarP(&limit_cpus, "limit_cpus", "", 5, "Container limit cpu core number")
	createCmd.Flags().Int64VarP(&limit_memory_mb, "limit_memory_mb", "", 12288, "Container limit memory mb")

}
