package internal

import (
	"fmt"
	"time"
)

func TimeParser(value string) (time.Time, error) {
	layouts := []string{
		time.RFC1123Z, // Mon, 02 Jan 2006 15:04:05 -0700
		time.RFC1123,  // Mon, 02 Jan 2006 15:04:05 MST
		time.RFC3339,  // 2006-01-02T15:04:05Z07:00
		time.RFC822Z,  // 02 Jan 06 15:04 -0700
		time.RFC822,   // 02 Jan 06 15:04 MST

		// Also accept days without a leading zero.
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format: %q", value)
}
