// Helpful Flickr API operations
package internal

import (
	"errors"
	"fmt"
	"gopkg.in/masci/flickr.v2"
	"gopkg.in/masci/flickr.v2/photosets"
	"strings"
)

func GetFlickrClient() (*flickr.FlickrClient, error) {
	config := GetConfig()

	apiKey := config.Flickr.ApiKey
	apiSecret := config.Flickr.ApiSecret
	oauthToken := config.Flickr.OauthToken
	oauthTokenSecret := config.Flickr.OauthSecret
	if apiKey == "" || apiSecret == "" || oauthToken == "" || oauthTokenSecret == "" {
		fmt.Println("Unable to continue. Please run the 'rodeo authenticate' command first")
		return nil, errors.New("credentials not set")
	}

	client := flickr.NewFlickrClient(apiKey, apiSecret)
	if client == nil {
		return nil, errors.New("unable to connect to Flickr")
	}
	client.OAuthToken = oauthToken
	client.OAuthTokenSecret = oauthTokenSecret

	return client, nil
}

// Get the list of Flickr photosets as a slice
func GetAlbums(client *flickr.FlickrClient, filter string) []photosets.Photoset {
	config := GetConfig()
	userId := config.Flickr.UserId

	response, err := photosets.GetList(client, true, userId, 1)
	if err != nil {
		fmt.Println(err)
		return []photosets.Photoset{}
	}

	photos := response.Photosets

	// Todo: check for pagination
	if photos.Pages > 1 {
		fmt.Println("More albums are available, but getting the second page is not implemented yet")
	}

	// Filter if required
	var albums []photosets.Photoset

	for _, photo := range photos.Items{
		if photo.Id == filter {
			albums = append(albums, photo)
		} else if strings.Contains(strings.ToLower(photo.Title), strings.ToLower(filter)) {
			albums = append(albums, photo)
		}
	}

	return albums
}
