package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

type ImageInfo struct {
	Width        uint                   `json:"ExifImageWidth"`
	Height       uint                   `json:"ExifImageHeight"`
	Title        string                 `json:"ObjectName"`
	Description  string                 `json:"Caption-Abstract"`
	Keywords     stringArray            `json:"Keywords"`
	Date         *time.Time
	Make         string                 `json:"Make"`
	Model        string                 `json:"Model"`
	ShutterSpeed string                 `json:"ShutterSpeedValue"`
	Aperture     json.Number            `json:"ApertureValue"`
	ISO          json.Number            `json:"ISO"`
	X            map[string]interface{} `json:"-"`
}

// A stringArray is an array of strings that has been unmarshalled from a JSON
// property that could be either a string or an array of string
type stringArray []string

func (sa *stringArray) UnmarshalJSON(data []byte) error {
	if len(data) > 0 {
		switch data[0] {
		case '[':
			// It's an array of strings and/or numbers, so unmarshal to an array of interfaces and then iterate over
			// the array converting each one to a string.
			var items []interface{}
			var s []string
			if err := json.Unmarshal(data, &items); err != nil {
				return err
			}

			for _, item := range items {
				s = append(s, fmt.Sprintf("%v", item))
			}

			*sa = s
		default:
			// It's a single element that may be a string or a number, unmarshal to an interface and convert to a
			// string.
			var item interface{}
			var s string
			if err := json.Unmarshal(data, &item); err != nil {
				return err
			}
			s = fmt.Sprintf("%v", item)
			*sa = []string{s}
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
	if err := json.Unmarshal(out, &info); err != nil {
		log.Println(err)
	}

	// unmarshall everything into info.X
	if err := json.Unmarshal(out, &info.X); err != nil {
		log.Println(err)
	}

	setImageInfoDate(&info)

	return &info, nil
}

func setImageInfoDate(info *ImageInfo) {
	// Set DateTimeOriginal with offset from OffsetTimeOriginal if it's set and is an offset
	dateTimeOriginal :=  info.X["DateTimeOriginal"].(string)

	// Determine timezone
	tz := time.FixedZone("UTC", 0)
	if maybeOffsetTimeOriginal, ok := info.X["OffsetTimeOriginal"]; ok {
		offsetTimeOriginal := maybeOffsetTimeOriginal.(string)
		tz = getTimeZoneFromOffset(offsetTimeOriginal, tz)
	} else if maybeOffsetTime, ok := info.X["OffsetTime"]; ok {
		offsetTime := maybeOffsetTime.(string)
		tz = getTimeZoneFromOffset(offsetTime, tz)
	}

	date, err := time.ParseInLocation("2006:01:02 15:04:05", dateTimeOriginal, tz)
	if err == nil {
		info.Date = &date
	}
}

// Extract the timezone from an offset of the form "±HH:MM"
func getTimeZoneFromOffset(offset string, tz *time.Location) *time.Location {
	if offset[0] == '+' || offset[0] == '-' {
		// Determine timezone offset
		re := regexp.MustCompile(`[+|-]?\d{2}`)
		parts := re.FindAllString(offset, -1)
		if len(parts) == 2 {
			offsetHours, _ := strconv.Atoi(parts[0])
			offsetMins, _ := strconv.Atoi(parts[1])

			if offsetHours != 0 && offsetMins != 0 {
				// Set the timezone as a numeric index
				// Note: as we don't know the Timezone identifier, leave it blank
				tz = time.FixedZone("", 60*((60*offsetHours)+(30*offsetMins)))
			}
		}
	}

	return tz
}
