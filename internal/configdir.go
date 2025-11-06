package internal

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os"
)

func ConfigDir() string {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return home + "/.config/rodeo"
}
