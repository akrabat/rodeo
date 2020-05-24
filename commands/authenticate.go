/*
Copyright © 2020 Rob Allen <rob@akrabat.com>

Use of this source code is governed by the MIT
license that can be found in the LICENSE file or at
https://akrabat.com/license/mit.
*/

/*
Package cmd implements the commands for the app. In this case, authenticating with Flickr.
*/
package commands

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/masci/flickr.v2"
	"os"
	"strings"
)

func init() {
	rootCmd.AddCommand(authenticateCmd)
}

// authenticateCmd represents the authenticate command
var authenticateCmd = &cobra.Command{
	Use:   "authenticate",
	Short: "Authenticate with Flickr",
	Long: "Authenticate with Flickr",
	Run: func(cmd *cobra.Command, args []string) {
		authenticate()
	},
}

func authenticate() {

	fmt.Println("Set up Flickr configuration")

	buf := bufio.NewReader(os.Stdin)

	apiKey := viper.GetString("flickr.api_key")
	fmt.Println("Please provide your Flickr API Key. (Default: " + apiKey + ")")
	fmt.Print("> ")
	input, err := buf.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	input = strings.TrimSuffix(input, "\n")
	if input == "" {
		input = apiKey
	}
	apiKey = input

	if apiKey == "" {
		fmt.Println("Error must provide an API Key")
		return
	}

	viper.Set("flickr.api_key", input)
	if err := viper.WriteConfig(); err != nil {
		fmt.Println("Error: Unable to write config", err)
	}

	fmt.Printf("Saved your Flickr API Key\n\n")


	apiSecret := viper.GetString("flickr.api_secret")
	if apiSecret != "" {
		fmt.Println("Please provide your Flickr API Secret (Leave blank to use current secret)")
	} else {
		fmt.Println("Please provide your Flickr API Secret")
	}
	fmt.Print("> ")
	input, err = buf.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	input = strings.TrimSuffix(input, "\n")
	if input == "" {
		input = apiSecret
	}
	apiSecret = input

	if apiSecret == "" {
		fmt.Println("Error must provide an API Secret")
		return
	}

	viper.Set("flickr.api_secret", input)
	if err := viper.WriteConfig(); err != nil {
		fmt.Println("Error: Unable to write config", err)
	}

	fmt.Printf("Saved your Flickr API Secret\n\n")


	// Authenticate
	client := flickr.NewFlickrClient(apiKey, apiSecret)

	requestToken, _ := flickr.GetRequestToken(client)
	url, _ := flickr.GetAuthorizeUrl(client, requestToken)
	url = strings.Replace(url, "delete", "write", -1)

	fmt.Println("Please authenticate with Flickr in your browser at:")
	fmt.Println("   " + url)
	fmt.Println("")
	fmt.Println("Enter your confirmation token here and press <return>")

	fmt.Print("> ")
	confirmationCode, err := buf.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	confirmationCode = strings.TrimSuffix(confirmationCode, "\n")

	accessTok, err := flickr.GetAccessToken(client, requestToken, confirmationCode)
	if err != nil {
		fmt.Println(err)
		return
	}

	viper.Set("flickr.oauth_token", accessTok.OAuthToken)
	viper.Set("flickr.oauth_token_secret", accessTok.OAuthTokenSecret)
	viper.Set("flickr.full_name", accessTok.Fullname)
	viper.Set("flickr.username", accessTok.Username)
	viper.Set("flickr.user_nsid", accessTok.UserNsid)
	if err := viper.WriteConfig(); err != nil {
		fmt.Println("Error: Unable to write config", err)
	}

	fmt.Println("")
	fmt.Println("Access token saved for " + accessTok.Fullname + ".")
}
