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
)

func init() {
	rootCmd.AddCommand(listAlbumsCmd)

	listAlbumsCmd.Flags().String("filter", "", "Filter the list of albums")
}

var listAlbumsCmd = &cobra.Command{
	Use:   "listalbums",
	Short: "List Flickr albums",
	Long:  `List Flickr albums`,
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

	flickr, err := GetFlickrClient()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	photosets := GetPhotosets(flickr, filter)

	if len(photosets) == 0 {
		fmt.Println("No albums found")
		return
	}

	for i, album := range photosets {
		fmt.Printf("%3d: %s (%s)\n", i+1, album.Title, album.Id)
	}
}
