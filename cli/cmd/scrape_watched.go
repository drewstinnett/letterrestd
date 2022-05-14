/*
Copyright © 2022 Drew Stinnett <drew@drewlink.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/apex/log"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
)

// watchedCmd represents the watched command
var watchedCmd = &cobra.Command{
	Use:   "watched USERNAME",
	Short: "Get a users watched film history",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// getExternalIds, err := cmd.Flags().GetBool("get-external-ids")
		// cobra.CheckErr(err)
		ctx := context.Background()
		watched, _, err := client.User.ListWatched(&ctx, args[0])
		cobra.CheckErr(err)
		d, err := yaml.Marshal(watched)
		cobra.CheckErr(err)
		fmt.Println(string(d))
		log.WithFields(log.Fields{
			"count": len(watched),
		}).Info("Watched movies")
	},
}

func init() {
	scrapeCmd.AddCommand(watchedCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// watchedCmd.PersistentFlags().String("foo", "", "A help for foo")
	// watchedCmd.PersistentFlags().Bool("get-external-ids", false, "Get external IDs for each film")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// watchedCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
