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
	"github.com/drewstinnett/letterrestd/letterboxd"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list USERNAME LIST-SLUG",
	Short: "Get information about a given list",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		filmC := make(chan *letterboxd.Film)
		doneC := make(chan error)
		go client.User.StreamListWithChan(ctx, args[0], args[1], filmC, doneC)
		for {
			select {

			case film := <-filmC:
				d, err := yaml.Marshal([]letterboxd.Film{
					*film,
				})
				cobra.CheckErr(err)
				fmt.Println(string(d))
			case err := <-doneC:
				if err != nil {
					log.WithError(err).Error("Error streaming watched")
				} else {
					log.Info("Finished")
					return
				}
			default:
			}
		}
	},
}

func init() {
	scrapeCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
