/*
Copyright © 2023 Bram Verschueren <bverschueren@redhat.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package logs

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// logReader reads (current/previous/rotated) logs from a tree in a base directory:
// $ tree path/to/<resource>/*/*/*/logs/
// ├── current.log
// ├── previous.insecure.log
// ├── previous.log
// └── rotated
//     ├── 0.log.xxx.gz
//     ├── 0.log.yyy.gz
//     └── 0.log.zzz

const (
	currentLogFile          = "current.log"
	previousLogFile         = "previous.log"
	previousInsecureLogFile = "previous.insecure.log"
	rotatedLogDir           = "rotated"
)

type LogReader struct {
	dirname string
	files   *[]string
	filter  logLineFilter
}

// Create a LogReader which holds a reader to either a plain (bufio.Reader) or gzipped (gzip.Reader) logfile.
func NewLogReader(dirname string) *LogReader {
	l := new(LogReader)
	l.dirname = dirname
	l.files = &[]string{currentLogFile}

	return l
}

func (l *LogReader) WithFilter(llf logLineFilter) {
	l.filter = llf
}

func (l *LogReader) FromPrevious() {
	l.files = &[]string{previousLogFile, previousInsecureLogFile}
}

func (l *LogReader) FromRotated() {
	l.files = rotatedFiles(l.dirname)
}

// Print the current reader (filtered) to a provided writer.
// If unfilter, write to provided writer (w).
// If filtered read from reader line-by-line and apply the filter.
func (l *LogReader) Read(w io.Writer) {
	for _, filename := range *l.files {
		reader, err := open(l.dirname + "/" + filename)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer reader.Close()
		if err != nil {
			fmt.Println(err)
		}
		if l.filter == nil {
			// without filter, copy entire content to the provided writer
			if _, err := io.Copy(w, reader); err != nil {
				log.Fatalf("fatal: %v", err)
			}
		} else {
			// with filter, read line by line and apply the filter
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				log := l.applyFilter(scanner.Bytes())
				if len(log) > 0 {
					fmt.Fprintln(w, string(log))
				}
			}
		}
	}
}

func (l *LogReader) applyFilter(raw []byte) []byte {
	if l.filter != nil {
		log, err := l.filter.filterLogLine(raw)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
		return log
	}
	return raw
}

// open the logfile and return either a *os.File or *gzip.Reader
func open(filename string) (io.ReadCloser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}
	var reader io.ReadCloser
	reader, err = gzip.NewReader(file)
	if err != nil {
		// after trying to read in a gzip.Reader, reset the offset to start
		file.Seek(0, io.SeekStart)
		reader = file
	}
	//	defer file.Close()
	return reader, nil
}

// read rotated dir and return relative filenames for plain and gzipped logfiles
func rotatedFiles(rotatedDir string) *[]string {
	var files []string
	err := filepath.WalkDir(rotatedDir+"/"+rotatedLogDir,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type().IsRegular() {
				files = append(files, filepath.Join(rotatedLogDir, d.Name()))
			}
			return nil
		})
	if err != nil {
		// if rotated dir does not exist, return empty slice
		if os.IsNotExist(err) {
			return &files
		}
		log.Println(err)
	}
	return &files
}
