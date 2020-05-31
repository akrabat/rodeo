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
)

func init() {
	rootCmd.AddCommand(listAlbumsCmd)
}

var listAlbumsCmd = &cobra.Command{
	Use:   "listalbums",
	Short: "List Flickr albums",
	Long: `List Flickr albums`,
	Run: func(cmd *cobra.Command, args []string) {
		listAlbums()
	},
}

func listAlbums() {
	fmt.Println("Listing Flickr albums")

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

	photosets := response.Photosets;
	if photosets.Pages > 1 {
		fmt.Println("More photosets are available, but getting the second page is not implemented yet")
	}

	for num, photoset := range photosets.Items {
		fmt.Printf("%d: %s (%s)\n", num+1, photoset.Title, photoset.Id)
	}

	fmt.Println("")
}
