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
	"github.com/rootsongjc/magpie/docker"
	"github.com/spf13/cobra"
)

var hostname string
var containerid string
var list string

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete containers on the host",
	Long:  "Delete all the existed containers on the host.",
	Run: func(cmd *cobra.Command, args []string) {
		if hostname != "" {
			docker.Delete_containers_on_host(hostname)
		}
		if containerid != "" {
			docker.Delete_container(containerid, nil)
		}
		if list != "" {
			docker.Delete_container_file_list(list)
		}
	},
}

func init() {
	dockerCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&hostname, "hostname", "n", "", "Hostname of the containers existed.")
	deleteCmd.Flags().StringVarP(&containerid, "containerid", "c", "", "Docker container ID.")
	deleteCmd.Flags().StringVarP(&list, "file", "f", "", "Docker containers list file, each line of a containerID.")

}
