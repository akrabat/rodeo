/*
Copyright © 2020 Rob Allen <rob@akrabat.com>

Use of this source code is governed by the MIT
license that can be found in the LICENSE file or at
https://akrabat.com/license/mit.
*/

/*
Package cmd implements the commands for the app. In this case, uploading an
image to Flickr.
*/
package commands

import (
	"github.com/akrabat/rodeo/internal"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/masci/flickr.v2"
	"gopkg.in/masci/flickr.v2/photosets"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	rootCmd.AddCommand(uploadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uploadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload <files>...",
	Short: "Upload images to Flickr",
	Long: `Upload images to Flickr

- sets the date uploaded to the creation time of the image so that 
  it appears in the photo stream at the right place.
- sets tags as per exif keywords.
- sets privacy if specific exif-keywords are set.
`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			fmt.Println("Error: At least one file must be specified.")
			os.Exit(2)
		}

		var photoIds []string
		for _, filename := range args {
			photoId := uploadFile(filename)
			if photoId != "" {
				photoIds = append(photoIds, photoId)
			}
		}
		fmt.Println("All Done")

		if len(photoIds) > 0 {
			username := viper.GetString("flickr.username")
			fmt.Printf("View: http://www.flickr.com/photos/%s'\n", username)
		}
	},
}

func uploadFile(filename string) string {
	fmt.Println("Processing " + filename)

	apiKey := viper.GetString("flickr.api_key")
	apiSecret := viper.GetString("flickr.api_secret")
	oauthToken := viper.GetString("flickr.oauth_token")
	oauthTokenSecret := viper.GetString("flickr.oauth_token_secret")
	if apiKey == "" || apiSecret == "" || oauthToken == "" || oauthTokenSecret == "" {
		fmt.Println("Unable to continue. Please run the 'rodeo authenticate' command first")
	}

	exiftool := viper.GetString("cmd.exiftool")
	if exiftool == "" {
		fmt.Println("Error: cmd.exiftool needs to be configured.")
		fmt.Println("Config file:", viper.ConfigFileUsed(), "\n")
		os.Exit(2)
	}

	info, err := internal.GetImageInfo(filename, exiftool)
	if err != nil {
		return ""
	}

	// process keyword settings from config
	var keywordsToRemove []string
	var keywordsToAdd []string
	var albumsToAddTo []string

	keywordSettings := viper.GetStringMap("keywords")
	if len(keywordSettings) > 0 {
		// Iterate over the keywords on this photo. for each keyword, look through the
		// list of keywords in the config and if it matches, then process the options for
		// that keyword:
		// 	  - delete: if `true`, then do not include this keyword as a Flickr tag
		//    - album_id: if set, then add this photo to that album
		//    - permissions: if set, then set permissions on this photo
		for _, keyword := range info.Keywords {
			addKeyword := true
			settings, ok := keywordSettings[keyword].(map[string]interface{})
			if ok {
				for key, val := range settings {
					if key == "delete" && val.(bool) == true {
						keywordsToRemove = append(keywordsToRemove, keyword)
						addKeyword = false
					}
					if key == "album_id" {
						albumsToAddTo = append(albumsToAddTo, val.(string))
					}
				}
			}
			if addKeyword {
				keywordsToAdd = append(keywordsToAdd, keyword)
			}
		}

		if len(keywordsToRemove) > 0 {
			// If exiftool is available, remove keywords from original file
			exiftool := viper.GetString("cmd.exiftool")
			if exiftool != "" {
				// Format of command: exiftool -overwrite_original -keywords-=one -keywords-=two FILENAME
				var parameters []string
				parameters = append(parameters, "-overwrite_original")
				for _, k := range keywordsToRemove {
					parameters = append(parameters, fmt.Sprintf("-keywords-=%s", k))
					parameters = append(parameters, fmt.Sprintf("-subject-=%s", k))
				}
				parameters = append(parameters, filename)
				//fmt.Println("Removing keywords from photo")
				cmd := exec.Command(exiftool, parameters...)
				cmd.Dir = filepath.Dir(filename)
				if err := cmd.Run(); err != nil {
					fmt.Println("Error: ", err)
				}
			}
		}
	} else {
		keywordsToAdd = info.Keywords
	}

	// Upload file to Flickr
	fmt.Println("Uploading photo to Flickr")
	client := flickr.NewFlickrClient(apiKey, apiSecret)
	client.OAuthToken = oauthToken
	client.OAuthTokenSecret = oauthTokenSecret

	title := strings.Trim(info.Title, " ")
	if title == "" {
		// no title - use filename (without extension)
		title = filepath.Base(filename)
		title = strings.TrimSuffix(title, filepath.Ext(filename))
	}

	// Upload photo
	params := flickr.UploadParams{
		Title:       title,
		Tags:        keywordsToAdd,
		IsPublic:    true,
		IsFamily:    true,
		IsFriend:    true,
		ContentType: 1, // photo
		Hidden:      1, // not hidden
		SafetyLevel: 1, // safe
	}
	if info.Description != "" {
		params.Description = info.Description
	}

	response, err := flickr.UploadFile(client, filename, &params)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	photoId := response.ID
	fmt.Printf("Uploaded photo '%s'\n", title)

	if len(albumsToAddTo) > 0 {
		// assign photo to each photoset in the list
		for _, albumId := range albumsToAddTo {
			respAdd, err := photosets.AddPhoto(client, albumId, photoId)
			if err != nil {
				//noinspection GoNilness
				fmt.Println("Failed adding photo to the set:"+albumId, err, respAdd.ErrorMsg())
			} else {
				fmt.Println("Added photo", photoId, "to set", albumId)
			}
		}
	}

	fmt.Println("")
	return photoId
}
