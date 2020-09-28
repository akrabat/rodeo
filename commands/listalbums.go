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
	. "github.com/akrabat/rodeo/internal"
	"github.com/spf13/cobra"
	"gopkg.in/masci/flickr.v2"
	"gopkg.in/masci/flickr.v2/photosets"
	"strings"
)

func init() {
	rootCmd.AddCommand(listAlbumsCmd)

	listAlbumsCmd.Flags().String("filter", "", "Filter the list of albums")
}

var listAlbumsCmd = &cobra.Command{
	Use:   "listalbums",
	Short: "List Flickr albums",
	Long: `List Flickr albums`,
	Run: func(cmd *cobra.Command, args []string) {
		filter, err := cmd.Flags().GetString("filter")
		if err != nil {
			filter = ""
		}

		listAlbums(filter)
	},
}

func listAlbums(filter string) {
	fmt.Print("Flickr albums")
	if filter != "" {
		fmt.Printf(" (filtered by %s)", filter)
	}
	fmt.Print("\n")

	config := GetConfig()
	apiKey := config.Flickr.ApiKey
	apiSecret := config.Flickr.ApiSecret
	oauthToken := config.Flickr.OauthToken
	oauthTokenSecret := config.Flickr.OauthSecret
	userId := config.Flickr.UserId

	if apiKey == "" || apiSecret == "" || oauthToken == "" || oauthTokenSecret == "" || userId == "" {
		fmt.Println("Unable to continue. Please run the 'rodeo authenticate' command first")
	}

	client := flickr.NewFlickrClient(apiKey, apiSecret)
	client.OAuthToken = oauthToken
	client.OAuthTokenSecret = oauthTokenSecret

	response, err := photosets.GetList(client, true, userId, 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	albums := response.Photosets;
	if albums.Pages > 1 {
		fmt.Println("More albums are available, but getting the second page is not implemented yet")
	}

	index := 0
	for _, album := range albums.Items {
		if strings.Contains(strings.ToLower(album.Title), strings.ToLower(filter)) {
			index += 1
			fmt.Printf("%3d: %s (%s)\n", index, album.Title, album.Id)
		}
	}

	fmt.Println("")
}
