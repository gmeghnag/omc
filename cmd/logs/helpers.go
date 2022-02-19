package logs

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"time"
)

var (
	// eol is the end-of-line sign in the log.
	eol = []byte{'\n'}
	// delimiter is the delimiter for timestamp and stream type in log line.
	delimiter = []byte{' '}
	// tagDelimiter is the delimiter for log tags.
	tagDelimiter = []byte(":")
)

const (
	// RFC3339NanoFixed is the fixed width version of time.RFC3339Nano.
	RFC3339NanoFixed = "2006-01-02T15:04:05.000000000Z07:00"
	// RFC3339NanoLenient is the variable width RFC3339 time format for lenient parsing of strings into timestamps.
	RFC3339NanoLenient = "2006-01-02T15:04:05.999999999Z07:00"
	// timeFormatOut is the format for writing timestamps to output.
	timeFormatOut = RFC3339NanoFixed
	// timeFormatIn is the format for parsing timestamps from other logs.
	timeFormatIn = RFC3339NanoLenient
)

// logMessage is the CRI internal log type.
type logMessage struct {
	timestamp time.Time
	stream    string
	log       []byte
}

// parseCRILog parses logs in CRI log format. CRI Log format example:
//   2016-10-06T00:17:09.669794202Z stdout P log content 1
//   2016-10-06T00:17:09.669794203Z stderr F log content 2
func parseCRILog(log []byte, infoLevel bool, warningLevel bool, errorLevel bool) (string, error) {
	var err error
	// Parse timestamp
	idx := bytes.Index(log, delimiter)
	if idx < 0 {
		return "", fmt.Errorf("timestamp is not found")
	}
	//only to check if timestamp is valid
	_, err = time.Parse(timeFormatIn, string(log[:idx]))
	if err != nil {
		return "", fmt.Errorf("unexpected timestamp format %q: %v", timeFormatIn, err)
	}

	// Parse stream type
	log = log[idx+1:]
	idx = bytes.Index(log, delimiter)
	if idx < 0 {
		return "", fmt.Errorf("stream type is not found")
	}
	stream := string(log[:idx])
	if len(stream) == 0 {
		return "", nil
	}
	if string(stream[0]) == "I" && infoLevel {
		return string(log), nil
	}
	if string(stream[0]) == "W" && warningLevel {
		return string(log), nil
	}
	if string(stream[0]) == "E" && errorLevel {
		return string(log), nil
	}

	return "", nil
}

func FilterCatLogs(filePath string, logLevels []string) {
	var infoLevel, warningLevel, errorLevel bool
	for _, i := range logLevels {
		if i == "info" {
			infoLevel = true
		}
		if i == "warning" {
			warningLevel = true
		}
		if i == "error" {
			errorLevel = true
		}
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("error: file " + filePath + " does not exist")
		os.Exit(1)
	}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error: can't open file " + filePath)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		log, err := parseCRILog(scanner.Bytes(), infoLevel, warningLevel, errorLevel)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if log != "" {
			fmt.Println(log)
		}

	}
}
