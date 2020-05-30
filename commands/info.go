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
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func init() {
	rootCmd.AddCommand(infoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// infoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// infoCmd displays info about the image file
var infoCmd = &cobra.Command{
	Use:   "info <files>...",
	Short: "View information on these files",
	Long: `View information on these files
`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			fmt.Println("Error: At least one file must be specified.")
			os.Exit(2)
		}

		exiftool := viper.GetString("cmd.exiftool")
		if exiftool == "" {
			fmt.Println("Error: cmd.exiftool needs to be configured.")
			fmt.Println("Config file:", viper.ConfigFileUsed(), "\n")
			os.Exit(2)
		}

		for _, filename := range args {
			fileInfo(filename, exiftool)
			fmt.Printf("\n")
		}
	},
}

func fileInfo(filename string, exiftool string) {
	fmt.Printf("%v:\n", filepath.Base(filename))

	info, err := internal.GetImageInfo(filename, exiftool)
	if err != nil {
		return
	}

	fmt.Printf("  Title:       %v\n", info.Title)
	fmt.Printf("  Description: %v\n", info.Description)

	sort.Sort(sort.StringSlice(info.Keywords[:]))
	fmt.Printf("  Keywords:    %v\n", strings.Join(info.Keywords[:], ", "))

	fmt.Printf("  Dimensions:  width:%v, height:%v\n", info.Width, info.Height)
	fmt.Printf("  Camera:      %v %v\n", info.Make, info.Model)
	fmt.Printf("  Exposure:    %vs, f/%v, ISO%v\n", info.ShutterSpeed, info.Aperture, info.ISO)
}
