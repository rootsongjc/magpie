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
	"strings"
)

var nodemanager string
var nodefile string

// decomisCmd represents the decomis command
var decomisCmd = &cobra.Command{
	Use:   "decomis",
	Short: "Decommising nodemanagers.",
	Long:  "Decommising nodemanagers of yarn clusters. ",
	Run: func(cmd *cobra.Command, args []string) {
		if nodemanager != "" {
			nms := strings.Split(nodemanager, ",")
			yarn.Decommis_nodemanagers(nms)
		}
		if nodefile != "" {
			yarn.Decommis_nodemanagers_through_file(nodefile)
		}
	},
}

func init() {
	yarnCmd.AddCommand(decomisCmd)

	decomisCmd.Flags().StringVarP(&nodefile, "nodefile", "f", "", "Each nodemanger a line in the file")
	decomisCmd.Flags().StringVarP(&nodemanager, "nodemanager", "n", "", "Nodemanager hostname，splited by comma")
}
