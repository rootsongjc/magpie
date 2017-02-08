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
	"github.com/rootsongjc/magpie/yarn"
	"github.com/spf13/cobra"
)

var offhost string

// offlineCmd represents the offline command
var offlineCmd = &cobra.Command{
	Use:   "offline",
	Short: "Offline the nodemanagers container of the host",
	Long:  "Decomissing the nodemanagers and then delete the docker contianers.",
	Run: func(cmd *cobra.Command, args []string) {
		if offhost != "" {
			yarn.Offline_host(offhost)
		}
	},
}

func init() {
	yarnCmd.AddCommand(offlineCmd)

	offlineCmd.Flags().StringVarP(&offhost, "hostname", "o", "", "Offline hostname")

}
