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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/akrabat/rodeo/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(resizeCmd)

	// Register command line options
	resizeCmd.Flags().BoolP("quiet", "q", false, "Just print name of resized file on completion")
}

// resizeCmd displays info about the image file
var resizeCmd = &cobra.Command{
	Use:   "resize <files>...",
	Short: "Resize files for use on the web",
	Long:  "Resize files for use on the web",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			fmt.Println("Error: At least one file must be specified.")
			os.Exit(2)
		}

		// Read the value of --quiet (if it is missing, the value is false)
		quiet, err := cmd.Flags().GetBool("quiet")
		if err != nil {
			quiet = false
		}

		for _, filename := range args {
			resize(filename, quiet)
		}
	},
}

// Resize image using ImageMagick's convert
//
// Example:
//
//	convert foo.jpg -scale 2000x2000 -interpolate catrom -quality 75 foo-web.jpeg
func resize(filename string, quiet bool) {

	config := GetConfig()
	convert := config.Cmd.Convert
	scale := config.Resize.Scale
	method := config.Resize.Method
	quality := config.Resize.Quality

	newFilename := filepath.Base(filename)
	newDirectory := filepath.Dir(filename)
	newFilename = strings.TrimSuffix(newFilename, filepath.Ext(filename))
	newFilename += fmt.Sprintf("-web%s", filepath.Ext(filename))

	// add directory
	newFilename = fmt.Sprintf("%s%c%s", newDirectory, os.PathSeparator, newFilename)

	var parameters []string
	parameters = append(parameters, filename)
	parameters = append(parameters, "-scale")
	parameters = append(parameters, scale)
	parameters = append(parameters, "-interpolate")
	parameters = append(parameters, method)
	parameters = append(parameters, "-quality")
	parameters = append(parameters, quality)
	parameters = append(parameters, newFilename)

	if !quiet {
		fmt.Printf("Resizing to %s at %s%% quality\n", scale, quality)
	}
	cmd := exec.Command(convert, parameters...)
	cmd.Dir = filepath.Dir(filename)
	if err := cmd.Run(); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	if quiet {
		fmt.Printf("%v\n", newFilename)
	} else {
		fmt.Printf("    Saved %v\n", newFilename)
	}
}
