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
	"github.com/spf13/viper"
	"strings"
)

func init() {
	rootCmd.AddCommand(viewConfigCmd)
}

var viewConfigCmd = &cobra.Command{
	Use:   "viewconfig",
	Short: "View configuration",
	Long:  `View configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		viewConfig()
	},
}

func viewConfig() {
	fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())

	config := GetConfig()

	flickr := config.Flickr
	fmt.Println("\nFlickr settings")
	if flickr.ApiKey == "" {
		fmt.Printf("  API Key is not set\n")
	} else {
		fmt.Printf("  API Key is set\n")
	}
	if flickr.Fullname == "" {
		fmt.Printf("  User is not authenticated\n")
	} else {
		fmt.Printf("  User: %v (username: %v, id:%v)\n", flickr.Fullname, flickr.Username, flickr.UserId)
	}

	resize := config.Resize
	fmt.Println("\nResize settings")
	fmt.Printf("  Method: %v\n", resize.Method)
	fmt.Printf("  Quality: %v\n", resize.Quality)
	fmt.Printf("  Scale: %v\n", resize.Scale)

	fmt.Println("\nUpload rules")
	for n, rule := range config.Rules {

		fmt.Printf("  Rule %v:\n", n+1)
		fmt.Printf("    Conditions:\n")
		includesAll := rule.Condition.IncludesAll
		includesAny := rule.Condition.IncludesAny
		excludesAny := rule.Condition.ExcludesAny
		if len(includesAll) > 0 {
			fmt.Printf("      - must include keyword%v: %v\n", PluralS(includesAll), strings.Join(includesAll, ", "))
		}
		if len(includesAny) > 0 {
			these := "any of these keywords"
			if len(includesAny) == 1 {
				these = "this keyword"
			}
			fmt.Printf("      - must include %v: %v\n", these, strings.Join(includesAny, ", "))
		}
		if len(excludesAny) > 0 {
			fmt.Printf("      - must not include keyword%v: %v\n", PluralS(excludesAny), strings.Join(excludesAny, ", "))
		}

		fmt.Printf("    Action:\n")
		if rule.Action.Delete {
			fmt.Printf("      Delete keyword\n")
		}

		albums := rule.Action.Albums
		if len(albums) > 0 {
			strs := make([]string, len(albums))
			for i, v := range albums {
				strs[i] = v.String()
			}
			fmt.Printf("      Add to album%v: %v\n", PluralS(albums), strings.Join(strs, ", "))
		}


	}
}
