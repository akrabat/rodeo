package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type ImageInfo struct {
	Width        uint        `json:"ExifImageWidth"`
	Height       uint        `json:"ExifImageHeight"`
	Title        string      `json:"ObjectName"`
	Description  string      `json:"Caption-Abstract"`
	Keywords     stringArray `json:"Keywords"`
	DateCreated  string      `json:"DateTimeOriginal"`
	Make         string      `json:"Make"`
	Model        string      `json:"Model"`
	ShutterSpeed string      `json:"ShutterSpeedValue"`
	Aperture     json.Number `json:"ApertureValue"`
	ISO          json.Number `json:"ISO"`
}

// A stringArray is an array of strings that has been unmarshalled from a JSON
// property that could be either a string or an array of string
type stringArray []string

func (sa *stringArray) UnmarshalJSON(data []byte) error {
	if len(data) > 0 {
		switch data[0] {
		case '"':
			var s string
			if err := json.Unmarshal(data, &s); err != nil {
				return err
			}
			*sa = stringArray([]string{s})
		case '[':
			var s []string
			if err := json.Unmarshal(data, &s); err != nil {
				return err
			}
			*sa = stringArray(s)
		}
	}
	return nil
}

// Read metadata (Exif/IPTC) from image using exiftool
func GetImageInfo(filename string, exiftool string) (*ImageInfo, error) {

	cmd := exec.Command(exiftool, "-j", "-exif:*", "-iptc:*", filename)

	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}

	// Exiftool always returns an array but as we only passed in one filename we know that there's only one element.
	// Remove the surrounding `[` and `]` from the JSON string to convert to a single object.
	out = out[1 : len(out)-2]

	info := ImageInfo{}
	err = json.Unmarshal(out, &info)
	if err != nil {
		log.Println(err)
	}

	return &info, nil
}