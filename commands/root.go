/*
Copyright © 2020 Rob Allen <rob@akrabat.com>

Use of this source code is governed by the MIT
license that can be found in the LICENSE file or at
https://akrabat.com/license/mit.
*/

/*
Package cmd implements the commands for the app. In this case, the root command.
*/
package commands

import (
	"fmt"
	"github.com/akrabat/rodeo/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var Version = "dev-build"
var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: Version,
	Use:     "rodeo",
	Short:   "Flickr command line tool",
	Long: `A command line tool to work with Flickr and images.

Rodeo uploads images to Flickr, applying keyword based rules to add the image
to albums and also to delete keywords that you may not want to be published. It
can also resize images for sharing on social media or in messages.
`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/rodeo.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Be quieter when running a command")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in ~/.config/rodeo directory with name "rodeo" (without extension).
		viper.AddConfigPath(internal.ConfigDir())
		viper.SetConfigName("rodeo")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Unable to read config file: ", err)
	}
}
