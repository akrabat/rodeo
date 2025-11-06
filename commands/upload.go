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
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/akrabat/rodeo/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/masci/flickr.v2"
	"gopkg.in/masci/flickr.v2/photos"
	"gopkg.in/masci/flickr.v2/photosets"
)

const uploadedListBaseFilename = "rodeo-uploaded-files.json"

var verbose bool

func init() {
	rootCmd.AddCommand(uploadCmd)

	// Register command line options
	uploadCmd.Flags().BoolP("force", "f", false, "Force upload of file even if already uploaded")
	uploadCmd.Flags().BoolP("dry-run", "n", false, "Show what would have been uploaded")
	uploadCmd.Flags().BoolP("verbose", "v", false, "Display additional messages during processing")
	uploadCmd.Flags().String("album", "", "Add to specific album, e.g. --album 12345678")
	uploadCmd.Flags().String("create-album", "", "Create a new album and add photo to it, e.g. --create-album 'SVR'")
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

		// Read the value of --force (if it is missing, the value is false)
		forceUpload, err := cmd.Flags().GetBool("force")
		if err != nil {
			forceUpload = false
		}

		// Read the value of --dry-run (if it is missing, the value is false)
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			dryRun = false
		}

		// Read the value of --verbose (if it is missing, the value is false)
		verbose, err = cmd.Flags().GetBool("verbose")
		if err != nil {
			verbose = false
		}

		var albums []Album
		var album Album

		// Read the value of --album (if it is missing, the value is empty)
		albumName, _ := cmd.Flags().GetString("create-album")
		if albumName != "" {
			albums, err = getAlbums(albumName)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}

			if len(albums) == 0 {
				album = Album{Name: albumName}
			} else {
				album = albums[0]
			}
		}

		// Read the value of --album (if it is missing, the value is empty)
		albumId, _ := cmd.Flags().GetString("album")
		if albums == nil && albumId != "" {
			albums, err = getAlbumsOrPromptForNewName(albumId)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}

			if len(albums) == 1 {
				album = albums[0]
			} else {
				album, err = chooseAlbumFromList(albums, albumId)
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}
			}
		}

		config := GetConfig()
		convertCmd := config.Cmd.Convert
		if convertCmd == "" {
			fmt.Println("Error: cmd.convert needs to be configured.")
			fmt.Println("Config file:", viper.ConfigFileUsed(), "\n")
			os.Exit(2)
		}

		var photoIds []string
		for _, filename := range args {
			// If the extension is tiff, then convert to jpeg
			var jpegFilename string
			var err error
			if filepath.Ext(filename) == ".tiff" {
				fmt.Printf("Converting %s to JPEG\n", filepath.Base(filename))
				jpegFilename, err = convertFileToJpeg(filename, convertCmd)
				if err != nil {
					fmt.Printf("Error: Failed to convert %s to JPEG\n", filename)
					continue
				}
				filename = jpegFilename
			}

			// Upload the file to Flickr
			photoId := uploadFile(filename, forceUpload, dryRun, &album)
			if photoId != "" {
				photoIds = append(photoIds, photoId)
			}

			if jpegFilename != "" {
				err = os.Remove(jpegFilename)
				if err != nil {
					fmt.Printf("Error: Failed to delete %s.\n", jpegFilename)
				}
			}
		}

		fmt.Println("All Done")
		fmt.Printf("View: http://www.flickr.com/photos/%s'\n", viper.GetString("flickr.username"))

		if len(photoIds) > 0 {
			fmt.Printf("Edit: http://www.flickr.com/photos/upload/edit/?ids=%s\n", strings.Join(photoIds, ","))
		}
	},
}

func debug(format string, a ...interface{}) {
	if verbose {
		message := fmt.Sprintf(format, a...)
		fmt.Println("DEBUG: " + message)
	}
}

func uploadFile(filename string, forceUpload bool, dryRun bool, album *Album) string {
	fmt.Println("Processing " + filename)

	config := GetConfig()

	exiftool := config.Cmd.Exiftool
	if exiftool == "" {
		fmt.Println("Error: cmd.exiftool needs to be configured.")
		fmt.Println("Config file:", viper.ConfigFileUsed(), "\n")
		os.Exit(2)
	}

	// Has this image been uploaded before?
	if uploadedPhotoId := getUploadedPhotoId(filename, config.Upload.StoreUploadListInImageDir); uploadedPhotoId != "" {
		fmt.Print("This image has already been uploaded to Flickr.")
		if forceUpload == true {
			fmt.Println(" Forcing upload.")
		} else {
			fmt.Printf("\nView this photo: http://www.flickr.com/photos/%s/%s\n", config.Flickr.Username, uploadedPhotoId)
			fmt.Println("")
			return ""
		}
	}

	info, err := GetImageInfo(filename, exiftool)
	if err != nil {
		return ""
	}

	// process rules
	var keywordsToRemove []string
	var keywordsToAdd []string
	var albumsToAddTo []Album
	var privacy Permissions
	privacy.SetDefaults()

	if album.Name != "" {
		albumsToAddTo = append(albumsToAddTo, *album)
	}

	if config.Rules != nil {
		for _, rule := range config.Rules {
			debug("Looking at rule '%s'", rule.Name)
			excludesAll := rule.Condition.ExcludesAll
			excludesAny := rule.Condition.ExcludesAny
			includesAll := rule.Condition.IncludesAll
			includesAny := rule.Condition.IncludesAny

			var intersection []string // applicable keywords from the condition

			// If the list of keywords for this image has all of `excludesAll`, then the rule is ignored
			if len(excludesAll) > 0 {
				intersection = Intersection(info.Keywords, excludesAll)
				if len(intersection) == len(excludesAll) {
					// Every `excludesAll` keyword is in info.Keywords, so this rule does not apply
					debug("Excluding due to `excludesAll`")
					continue
				}
				//fmt.Println("`excludesAll` condition does not apply")
			}

			// If the list of keywords for this image has any from `excludesAny`, then the rule is ignored
			if len(excludesAny) > 0 {
				intersection = Intersection(info.Keywords, excludesAny)
				if len(intersection) > 0 {
					// At least one `excludesAny` keyword is in info.Keywords, so this rule does not apply
					debug("Excluding due to `excludesAny`")
					continue
				}
				//fmt.Println("`excludesAny` condition does not apply")
			}

			processRules := false
			if len(includesAll) > 0 {
				//  info.Keywords must contain all keywords in `includesAll`
				intersection = Intersection(info.Keywords, includesAll)
				if len(intersection) != len(includesAll) {
					// All `includesAll` keywords do not exist, so this rule does not apply
					debug("Excluding due to `includesAll`")
					continue
				}
				//fmt.Println("`includesAll` condition is met")
				processRules = true
			} else if len(includesAny) > 0 {
				//  info.Keywords must contain all keywords in `includesAny`
				intersection = Intersection(info.Keywords, includesAny)
				if len(intersection) == 0 {
					// There are no `includesAny` keywords in info.Keywords, so this rule does not apply
					debug("Excluding due to `includesAny`")
					continue
				}
				//fmt.Println("`includesAny` condition is met")
				processRules = true
			}

			if processRules {
				debug("Will process rules")
				debug("Applicable keywords: %s", strings.Join(intersection, ", "))
				if rule.Action.Delete {
					keywordsToRemove = append(keywordsToRemove, intersection...)
				}
				if rule.Action.Privacy != nil {
					privacy = *rule.Action.Privacy
				}
				if len(rule.Action.Albums) > 0 {
					for _, thisAlbum := range rule.Action.Albums {
						albumsToAddTo = append(albumsToAddTo, thisAlbum)
					}
				}
			}
		}
	} else {
		debug("No config found")
	}

	// Set the keywords to be added to the Flickr photo record
	if len(keywordsToRemove) > 0 {
		difference := Difference(info.Keywords, keywordsToRemove)
		keywordsToAdd = difference
	} else {
		keywordsToAdd = info.Keywords
	}

	// output what we are going to do
	if len(keywordsToRemove) > 0 || len(albumsToAddTo) > 0 {
		fmt.Printf("Actions:\n")
		if len(keywordsToRemove) > 0 {
			fmt.Printf("  - keywords to remove: %s\n", strings.Join(keywordsToRemove, ", "))
		}

		fmt.Printf("  - privacy will be set to: Family: %v, Friends: %v, Public: %v\n", privacy.Family, privacy.Friends, privacy.Public)

		if len(albumsToAddTo) > 0 {
			strs := make([]string, len(albumsToAddTo))
			for i, a := range albumsToAddTo {
				strs[i] = a.Name
			}
			fmt.Printf("  - albums to add to: \"%s\"\n", strings.Join(strs, "\", \""))
		}
	}

	title := strings.Trim(info.Title, " ")
	fmt.Printf("  - title will be set to \"%s\"\n", title)
	fmt.Printf("\n")

	// All ready to process now
	if dryRun {
		fmt.Println("Would upload photo to Flickr")
		return ""
	}

	if len(keywordsToRemove) > 0 && exiftool != "" {
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

	// Upload file to Flickr
	fmt.Println("Uploading photo to Flickr")

	client, err := GetFlickrClient()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	if title == "" {
		// no title - use filename (without extension)
		title = filepath.Base(filename)
		title = strings.TrimSuffix(title, filepath.Ext(filename))
	}

	// Upload photo

	// quote keywords for Flickr's tags
	tags := make([]string, len(keywordsToAdd))
	for i, kw := range keywordsToAdd {
		tags[i] = fmt.Sprintf("\"%s\"", kw)
	}

	params := flickr.UploadParams{
		Title:       title,
		Tags:        tags,
		IsFamily:    privacy.Family,
		IsFriend:    privacy.Friends,
		IsPublic:    privacy.Public,
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
	recordUpload(filename, photoId, config.Upload.StoreUploadListInImageDir)
	fmt.Printf("Uploaded photo '%s'\n", title)

	// set date posted to the date that the photo was taken so that it's in the right place
	// in the Flickr photo stream
	setDatePosted := config.Upload.SetDatePosted
	if setDatePosted == true && info.Date != nil {
		datePosted := fmt.Sprintf("%d", info.Date.Unix())
		respSetDate, err := photos.SetDates(client, photoId, datePosted, "")
		if err != nil {
			// noinspection GoNilness
			fmt.Printf("Failed update photo %v's date posted: %v\n%v\n", photoId, err, respSetDate.ErrorMsg())
		}
	}

	if len(albumsToAddTo) > 0 {
		// assign photo to each photoset in the list
		for _, thisAlbum := range albumsToAddTo {
			if thisAlbum.Id == "" {
				// create new photoset on Flickr
				respAdd, err := photosets.Create(client, thisAlbum.Name, "", photoId)
				if err != nil {
					// noinspection GoNilness
					fmt.Println("Failed to create photoset: "+thisAlbum.Name, err, respAdd.ErrorMsg())
				} else {
					if album.Name == thisAlbum.Name {
						album.Id = respAdd.Set.Id
					}
					fmt.Println("Added photo", photoId, "to new set", album.String())
				}
			} else {
				// add to this photoset on Flickr
				respAdd, err := photosets.AddPhoto(client, thisAlbum.Id, photoId)
				if err != nil {
					// noinspection GoNilness
					fmt.Println("Failed adding photo to the set: "+thisAlbum.String(), err, respAdd.ErrorMsg())
				} else {
					fmt.Println("Added photo", photoId, "to set", thisAlbum.String())
				}
			}
		}
	}

	fmt.Printf("View this photo: http://www.flickr.com/photos/%s/%s\n", config.Flickr.Username, photoId)
	fmt.Println("")
	return photoId
}

// Convert file to JPEG using convert
func convertFileToJpeg(filename string, convertCmd string) (string, error) {
	jpegFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".jpeg"

	cmd := exec.Command(convertCmd, filename, jpegFilename)
	_, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s\n", err.(*exec.ExitError).Stderr)
		return "", err
	}

	return jpegFilename, nil
}

func getUploadedListFilename(imageFilename string, storeUploadListInImageDirectory bool) string {
	var directory string

	if storeUploadListInImageDirectory {
		// File is stored in directory where image is and is hidden via a leading `.` on the imageFilename
		directory = filepath.Dir(imageFilename)
		return directory + "/." + uploadedListBaseFilename
	}

	// Storing to the config directory
	return ConfigDir() + "/" + uploadedListBaseFilename
}

// Has this file been uploaded to Flickr?
// Check the `.rodeo-uploaded-files` file that resides in the same directory as `filename`
func getUploadedPhotoId(filename string, storeUploadedListInImageDirectory bool) string {
	uploadedListFilename := getUploadedListFilename(filename, storeUploadedListInImageDirectory)
	filenames := readUploadedListFile(uploadedListFilename)

	// Is imageFilename a key in the map?
	imageFilename := filepath.Base(filename)
	if photoId, ok := filenames[imageFilename]; ok {
		// imageFilename exists, return its associated photoId
		return photoId
	}

	return ""
}

// Record the filename of the image uploaded into the uploaded list
func recordUpload(filename string, photoId string, storeUploadedListInImageDirectory bool) {
	imageFilename := filepath.Base(filename)
	uploadedListFilename := getUploadedListFilename(filename, storeUploadedListInImageDirectory)
	filenames := readUploadedListFile(uploadedListFilename)

	// If the imageFilename is already recorded, then there's nothing to do
	if _, ok := filenames[imageFilename]; ok {
		return
	}

	// Filename not in list, so append to list and save
	filenames[imageFilename] = photoId
	writeUploadedListFile(filenames, uploadedListFilename)
}

// Read the uploaded list from the `uploadedListFilename` and convert to a map from the JSON
func readUploadedListFile(uploadedListFilename string) map[string]string {
	filenames := make(map[string]string)

	// Does the file exist?
	if _, err := os.Stat(uploadedListFilename); err == nil || os.IsExist(err) {
		// File exists - therefore read it
		data, err := ioutil.ReadFile(uploadedListFilename)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return filenames
		}

		err = json.Unmarshal(data, &filenames)
		if err != nil {
			fmt.Println("error:", err)
		}
	}

	return filenames
}

// Write the uploaded list to the `uploadedListFilename` in JSON format
func writeUploadedListFile(filenames map[string]string, uploadedListFilename string) {
	// Convert to JSON
	data, err := json.MarshalIndent(filenames, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}

	// Write to disk
	err = ioutil.WriteFile(uploadedListFilename, data, 0664)
	if err != nil {
		fmt.Printf("Error: Unable to write %s: %v", filepath.Base(uploadedListFilename), err)
		return
	}
}

// Retrieve album from Flickr's API so that we have the full information about it
func getAlbums(albumId string) ([]Album, error) {
	var albums []Album
	if albumId == "" {
		// No album to look for! So we return an empty album
		albums = append(albums, Album{})
		return albums, nil
	}

	client, err := GetFlickrClient()
	if err != nil {
		fmt.Println(err)
		return []Album{}, err
	}

	photosets := GetPhotosets(client, albumId)
	if len(photosets) == 0 {
		// no photsets found, so return an empty album
		return []Album{}, nil
	}

	// At least one photoset found
	for _, photo := range photosets {
		album := Album{
			Id:   photo.Id,
			Name: photo.Title,
		}
		albums = append(albums, album)
	}
	return albums, nil
}

// Retrieve album from Flickr's API so that we have the full information about it
func getAlbumsOrPromptForNewName(albumId string) ([]Album, error) {
	albums, err := getAlbums(albumId)
	if err != nil {
		return albums, err
	}

	if len(albums) == 0 {
		// No photosets found. Prompt to see if we want to create one
		fmt.Printf("No albums found for %s\n", albumId)
		album, err := promptForNewAlbum(albumId)
		if err != nil {
			return []Album{}, err
		}

		albums = append(albums, *album)
	}

	return albums, nil
}

// Ask the user for an album to create
func promptForNewAlbum(albumId string) (*Album, error) {
	// Ask user for name of new album
	fmt.Printf("Enter name for new album (press enter for '%s'): \n", albumId)
	reader := bufio.NewReader(os.Stdin)
	chosenAlbumName, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	chosenAlbumName = strings.TrimSuffix(chosenAlbumName, "\n")

	if chosenAlbumName == "" {
		chosenAlbumName = albumId
	}

	chosenAlbum := Album{Name: chosenAlbumName}
	return &chosenAlbum, nil
}

// Display the available albums and allow the user to select one
func chooseAlbumFromList(albums []Album, searchTerm string) (Album, error) {
	fmt.Printf("Available albums that match \"%s\":\n", searchTerm)
	for i, album := range albums {
		fmt.Printf("%3d: %s (%s)\n", i+1, album.Name, album.Id)
	}

	// Ask for user input
	fmt.Println("Select album: ")
	reader := bufio.NewReader(os.Stdin)
	chosenAlbum, err := reader.ReadString('\n')
	if err != nil {
		return Album{}, err
	}
	chosenAlbum = strings.TrimSuffix(chosenAlbum, "\n")

	// Convert to integer
	chosenAlbumIndex, err := strconv.Atoi(chosenAlbum)
	if err != nil {
		return Album{}, err
	}

	return albums[chosenAlbumIndex-1], nil
}
