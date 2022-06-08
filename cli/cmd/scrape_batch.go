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

// batchCmd represents the batch command
var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Do a batch scrape to get films from different places",
	Run: func(cmd *cobra.Command, args []string) {
		userWatched, err := cmd.Flags().GetStringArray("user-watched")
		cobra.CheckErr(err)
		filmOpts := &letterboxd.FilmBatchOpts{
			Watched: userWatched,
		}
		ctx := context.Background()
		films, pagination, err := client.Film.StreamBatch(ctx, filmOpts)
		cobra.CheckErr(err)
		log.Infof("Paginage: %+v", pagination)
		log.Debug("ABOUT TO JUMP IN TO FOR LOOP")
		for i := 0; i < pagination.TotalItems; i++ {
			film := <-films
			d, err := yaml.Marshal(film)
			cobra.CheckErr(err)
			fmt.Println(string(d))
		}
	},
}

func init() {
	scrapeCmd.AddCommand(batchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchCmd.PersistentFlags().String("foo", "", "A help for foo")
	batchCmd.PersistentFlags().StringArray("user-watched", []string{}, "Watched films for a given user")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
