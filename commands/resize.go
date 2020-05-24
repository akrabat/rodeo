/*
Copyright © 2020 Rob Allen <rob@akrabat.com>

Use of this source code is governed by the MIT
license that can be found in the LICENSE file or at
https://akrabat.com/license/mit.
*/

/*
Package cmd implements the commands for the app. In this case, resizing an image for
use on the web
*/
package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	rootCmd.AddCommand(resizeCmd)
}

// resizeCmd displays info about the image file
var resizeCmd = &cobra.Command{
	Use:   "resize <files>...",
	Short: "Resize files for use on the web",
	Long: "Resize files for use on the web",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			fmt.Println("Error: At least one file must be specified.")
			os.Exit(2)
		}

		for _, filename := range args {
			resize(filename)
		}
	},
}

func getSetting(name string, defaultValue string) string {
	value := viper.GetString(name)
	if value == "" {
		value = defaultValue
		viper.Set(name, value)
		viper.WriteConfig()
	}
	return value
}

func resize(filename string) {

	//convert RKA-20200527-073835-IMG_7029.jpeg -scale 2000x2000 -interpolate catrom -quality 75  RKA-20200527-073524-IMG_7028-web.jpeg

	convert := getSetting("cmd.convert", "/usr/local/bin/convert")
	scale := getSetting("resize.scale", "2000x2000")
	method := getSetting("resize.method", "catrom")
	quality := getSetting("resize.quality", "75")

	newFilename := filepath.Base(filename)
	newFilename = strings.TrimSuffix(newFilename, filepath.Ext(filename))
	newFilename += fmt.Sprintf("-web%s", filepath.Ext(filename))

	var parameters []string
	parameters = append(parameters, filename)
	parameters = append(parameters, "-scale")
	parameters = append(parameters, scale)
	parameters = append(parameters, "-interpolate")
	parameters = append(parameters, method)
	parameters = append(parameters, "-quality")
	parameters = append(parameters, quality)
	parameters = append(parameters, newFilename)

	fmt.Printf("Resizing to %s at %s%% quality\n", scale, quality)
	cmd := exec.Command(convert, parameters...)
	cmd.Dir = filepath.Dir(filename)
	if err := cmd.Run(); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	fmt.Printf("    Saved %v\n", newFilename)
}