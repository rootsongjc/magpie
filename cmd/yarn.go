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

	"github.com/spf13/cobra"
)

var clustername string
// yarnCmd represents the yarn command
var yarnCmd = &cobra.Command{
	Use:   "yarn",
	Short: "Yarn cluster management tool.",
	Long: "Get cluster cluster status.",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	RootCmd.AddCommand(yarnCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// yarnCmd.PersistentFlags().String("foo", "", "A help for foo")

	 yarnCmd.PersistentFlags().StringVarP(&clustername,"clustername", "c", "", "Yarn cluster name")

}
