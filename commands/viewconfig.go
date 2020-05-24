/*
Copyright © 2020 Rob Allen <rob@akrabat.com>

Use of this source code is governed by the MIT
license that can be found in the LICENSE file or at
https://akrabat.com/license/mit.
*/

/*
Package cmd implements the commands for the app. In this case, listing Flickr albums.
*/
package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(viewConfigCmd)
}

var viewConfigCmd = &cobra.Command{
	Use:   "viewconfig",
	Short: "View configuration",
	Long: `View configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		viewConfig()
	},
}

func viewConfig() {
	fmt.Println("Rodeo Configuration\n")
	fmt.Println("Using config file:", viper.ConfigFileUsed(), "\n")


	apiKey := viper.GetString("flickr.api_key")
	userFullName := viper.GetString("flickr.full_name")
	username := viper.GetString("flickr.username")
	userId := viper.GetString("flickr.user_nsid")

	fmt.Println("Flickr settings")
	fmt.Printf("  API Key: %v\n", apiKey)
	fmt.Printf("  User: %v (%v, %v)\n", userFullName, username, userId)

	keywords := viper.GetStringMap("keywords")
	fmt.Println("\nKeyword settings")
	for keyword, settings := range keywords {
		fmt.Printf("  %v: \n", keyword)

		settings, _ := settings.(map[string]interface{})
		for key, val := range settings {
			fmt.Printf("    %s: %v\n", key, val)
		}
	}

	fmt.Println("")
}
