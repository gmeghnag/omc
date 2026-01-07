package use

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	pathlib "path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
)

const (
	fileTypeTar     string = "tar"
	fileTypeTarGzip string = "tar.gz"
	fileTypeXZ      string = "tar.xz"
	fileTypeZip     string = "zip"
)

func humanizeBytes(bytes int64) string {
	var human string
	if float64(bytes) < math.Pow(2, 10) {
		human = fmt.Sprintf("%.0f B", float64(bytes))
	} else if float64(bytes) < math.Pow(2, 20) {
		human = fmt.Sprintf("%.1f K", float64(bytes)/math.Pow(2, 10))
	} else {
		human = fmt.Sprintf("%.1f M", float64(bytes)/math.Pow(2, 20))
	}
	return human
}

type WriteCounter struct {
	length     string
	downloaded int64
	lastShown  time.Time
}

func NewWriteCounter(total int64) *WriteCounter {
	length := ""
	if total != -1 {
		length = humanizeBytes(total)
	} else {
		length = "?"
	}
	counter := &WriteCounter{
		length:     length,
		downloaded: 0,
		lastShown:  time.Now(),
	}
	return counter
}

func (counter *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	counter.downloaded += int64(n)
	counter.ShowProgress()
	return n, nil
}

func (counter *WriteCounter) Downloaded() string {
	return humanizeBytes(counter.downloaded)
}

func (counter *WriteCounter) ShowProgress() {
	// rate limit
	throttleDuration, _ := time.ParseDuration("100ms")
	if time.Since(counter.lastShown).Nanoseconds() < throttleDuration.Nanoseconds() {
		return
	}

	fmt.Printf("\r%s", strings.Repeat(" ", 78))
	fmt.Printf("\rDownloading... %s / %s", counter.Downloaded(), counter.length)

	counter.lastShown = time.Now()
}

func GetHeaderFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot open "+path+": "+err.Error())
		return "", err
	}
	defer file.Close()

	buff := make([]byte, 512)

	_, err = file.Read(buff)

	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading file header: "+err.Error())
		return "", err
	}
	filetype := http.DetectContentType(buff)

	return filetype, nil
}

func isTarFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot open "+path+": "+err.Error())
		return false, err
	}
	defer file.Close()
	tarReader := tar.NewReader(file)
	_, err = tarReader.Next()
	if err != nil {
		return false, fmt.Errorf("unable to read tarbal file: %w", err)
	}

	return true, nil
}

func isZip(path string) (bool, error) {
	header, err := GetHeaderFile(path)
	if err == nil {
		return header == "application/zip", nil
	}
	return false, err
}

func isGzip(path string) (bool, error) {
	header, err := GetHeaderFile(path)
	if err == nil {
		return header == "application/x-gzip", nil
	}
	return false, err
}

func isXZ(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	_, err = xz.NewReader(file)
	if err != nil {
		return false, err
	}
	return true, nil
}

func IsCompressedFile(path string) (bool, string, error) {
	result, err := isGzip(path)
	if err != nil {
		return false, "", err
	} else if result {
		return result, fileTypeTarGzip, nil
	}

	result, err = isZip(path)
	if err != nil {
		return false, "", err
	} else if result {
		return result, fileTypeZip, nil
	}

	result, err = isXZ(path)
	if err != nil {
		return false, "", err
	} else if result {
		return result, fileTypeXZ, nil
	}

	result, err = isTarFile(path)
	if err != nil {
		return false, "", err
	}

	return result, fileTypeTar, nil
}

func IsRemoteFile(path string) bool {
	parsedURL, err := url.Parse(path)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

func DownloadFile(path string) (string, error) {
	tmpdir, err := os.MkdirTemp("", "omc-*")
	if err != nil {
		return "", err
	}

	resp, err := http.Get(path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Use a sensible filename
	var filename string
	// First, try to extract filename from headers
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			filename = params["filename"]
		}
	}
	// If that fails, resort to parsing the path
	if filename == "" {
		if parsedURL, err := url.Parse(path); err == nil {
			filename = pathlib.Base(parsedURL.Path)
		}
	}

	outpath := filepath.Join(tmpdir, filename)
	fmt.Println("Downloading file " + path + " in " + outpath)

	out, err := os.Create(outpath)
	if err != nil {
		return "", err
	}

	counter := NewWriteCounter(resp.ContentLength)
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return "", err
	}

	out.Close()
	fmt.Println()

	return out.Name(), nil
}

func CopyFile(path string, destinationfile string) error {
	source, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error opening file "+path+": "+err.Error())
		return err
	}
	defer source.Close()
	dest, err := os.Create(destinationfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating file "+destinationfile+": "+err.Error())
		return err
	}
	defer dest.Close()
	_, err = io.Copy(dest, source)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error copying file "+path+" to "+destinationfile+": "+err.Error())
	}
	return err
}

func DecompressFile(path string, outpath string, fileType string) (string, error) {
	fmt.Println("Decompressing file " + path + " in " + outpath)
	var err error
	var mgRootDir string = ""

	switch fileType {
	case fileTypeTar:
		mgRootDir, err = ExtractTar(path, outpath)
	case fileTypeTarGzip:
		mgRootDir, err = ExtractTarGz(path, outpath)
	case fileTypeXZ:
		mgRootDir, err = extractTarXZ(path, outpath)
	case fileTypeZip:
		mgRootDir, err = ExtractZip(path, outpath)
	default:
		return "", fmt.Errorf("unable to decompress file: unknown file type %s", fileType)
	}

	return mgRootDir, err
}

func ExtractTarStream(st io.Reader, destinationdir string) (string, error) {
	firstDirectory := false
	var mgRootDir string = ""
	tarReader := tar.NewReader(st)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, "cannot extract tar: "+err.Error())
			return "", err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if !firstDirectory {
				firstDirectory = true
				mgRootDir = destinationdir + "/" + header.Name
			}
			directory := filepath.Join(destinationdir, header.Name)
			if _, err := os.Stat(directory); os.IsNotExist(err) {
				if err := os.Mkdir(directory, 0755); err != nil {
					fmt.Fprintln(os.Stderr, "mkdir failed extracting tar: "+err.Error())
					return "", err
				}
			}
		case tar.TypeReg:
			// Root dir is not part of the archive
			if mgRootDir == "" {
				mgRootDir = filepath.Join(destinationdir, filepath.Dir(header.Name))
				firstDirectory = true
				err := os.MkdirAll(mgRootDir, os.ModePerm)
				if err != nil && !os.IsExist(err) {
					return "", err
				}
			}
			outpath := filepath.Join(destinationdir, header.Name)
			if _, err := os.Stat(outpath); !os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "create file failed extracting tar: file already exists")
			}
			outFile, err := os.Create(outpath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "create file failed extracting tar: "+err.Error())
				return "", err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				fmt.Fprintln(os.Stderr, "copy file failed extracting tar: "+err.Error())
				return "", err
			}
			outFile.Close()
		default:
			fmt.Fprintf(os.Stderr, "unknown type(%s) in %s: "+err.Error(), header.Typeflag, header.Name)
			return "", err
		}
	}
	return mgRootDir, nil
}

func ExtractTar(tarfile string, destinationdir string) (string, error) {
	tarStream, err := os.Open(tarfile)
	var mgRootDir string
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot open "+tarfile+": "+err.Error())
		return "", err
	}
	defer tarStream.Close()

	var fileReader io.ReadCloser = tarStream
	mgRootDir, err = ExtractTarStream(fileReader, destinationdir)

	return mgRootDir, err
}

func ExtractZip(zipfile string, destinationdir string) (string, error) {

	firstDirectory := false
	var mgRootDir string = ""
	archive, err := zip.OpenReader(zipfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot uncompress zip "+zipfile+": "+err.Error())
		return "", err
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(destinationdir, f.Name)

		// Root dir is not part of the archive
		if !f.FileInfo().IsDir() && mgRootDir == "" {
			mgRootDir = filepath.Dir(filePath)
			firstDirectory = true
			err := os.MkdirAll(mgRootDir, os.ModePerm)
			if err != nil && !os.IsExist(err) {
				return "", err
			}
		}

		if f.FileInfo().IsDir() {
			if !firstDirectory {
				firstDirectory = true
				mgRootDir = filePath
			}
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error: cannot create directory "+filePath+": "+err.Error())
				return "", err
			}
		} else {
			dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				fmt.Fprintln(os.Stderr, "error: cannot create file "+filePath+": "+err.Error())
				return "", err
			}
			defer dstFile.Close()

			fileInArchive, err := f.Open()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error: cannot open file "+f.Name+": "+err.Error())
				return "", err
			}
			defer fileInArchive.Close()

			if _, err := io.Copy(dstFile, fileInArchive); err != nil {
				fmt.Fprintln(os.Stderr, "error: cannot copy file to "+dstFile.Name()+": "+err.Error())
				return "", err
			}
		}
	}

	return mgRootDir, err
}

func ExtractTarGz(gzipfile string, destinationdir string) (string, error) {
	gzipStream, err := os.Open(gzipfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot open "+gzipfile+": "+err.Error())
		return "", err
	}
	defer gzipStream.Close()
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot uncompress gzip "+gzipfile+": "+err.Error())
		return "", err
	}
	return ExtractTarStream(uncompressedStream, destinationdir)
}

func extractTarXZ(xzFile string, destinationdir string) (string, error) {
	stream, err := os.Open(xzFile)
	if err != nil {
		return "", fmt.Errorf("error: cannot open %q: %w", xzFile, err)
	}
	defer stream.Close()

	xzReader, err := xz.NewReader(stream)
	if err != nil {
		return "", fmt.Errorf("error: cannot uncompress xz file %q: %w", xzFile, err)
	}
	return ExtractTarStream(xzReader, destinationdir)
}

const (
	prowHost = "prow.ci.openshift.org"
)

// IsProwURL checks if the given URL is a Prow Job URL
func IsProwURL(path string) bool {
	parsedURL, err := url.Parse(path)
	if err != nil {
		return false
	}
	return parsedURL.Host == prowHost && strings.HasPrefix(parsedURL.Path, "/view/gs/")
}

// GetArtifactsURLFromProw fetches the Prow job page and extracts the Artifacts link
func GetArtifactsURLFromProw(prowURL string) (string, error) {
	fmt.Printf("Fetching Prow job page: %s\n", prowURL)

	resp, err := http.Get(prowURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Prow page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d for Prow page", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Prow page body: %w", err)
	}

	// Look for the Artifacts link in the page
	// The link typically contains "gcsweb" and points to the artifacts directory
	artifactsURL := findArtifactsLink(string(body), prowURL)
	if artifactsURL == "" {
		return "", fmt.Errorf("could not find Artifacts link in Prow page")
	}

	return artifactsURL, nil
}

// findArtifactsLink searches the HTML content for the Artifacts link
func findArtifactsLink(content string, baseURL string) string {
	// Parse base URL for resolving relative links
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	// Look for links containing "gcsweb" or links with text "Artifacts"
	// Common patterns:
	// - <a href="https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/...">Artifacts</a>
	// - Links with "PR Artifacts" text
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Check if line contains artifacts-related keywords
		lineLower := strings.ToLower(line)
		if !strings.Contains(lineLower, "artifact") && !strings.Contains(lineLower, "gcsweb") {
			continue
		}

		// Extract href from the line
		idx := strings.Index(line, `href="`)
		if idx == -1 {
			continue
		}
		start := idx + 6
		end := strings.Index(line[start:], `"`)
		if end == -1 {
			continue
		}
		href := line[start : start+end]

		// Check if this is a gcsweb link
		if strings.Contains(href, "gcsweb") {
			// Return the URL, ensuring it ends with /
			if !strings.HasSuffix(href, "/") {
				href += "/"
			}
			return href
		}

		// Handle relative URLs that might be artifacts links
		if strings.Contains(lineLower, "artifact") {
			refURL, err := url.Parse(href)
			if err != nil {
				continue
			}
			absURL := base.ResolveReference(refURL).String()
			if strings.Contains(absURL, "gcsweb") {
				if !strings.HasSuffix(absURL, "/") {
					absURL += "/"
				}
				return absURL
			}
		}
	}

	return ""
}

// parseGCSWebLinks parses the HTML content from a gcsweb page and extracts directory and file links
// It only returns links that are children of the baseURL (going downwards only)
func parseGCSWebLinks(body []byte, current string) (dirs []string, files []string) {
	content := string(body)

	// Parse base URL to construct absolute URLs
	currentURL, err := url.Parse(current)
	if err != nil {
		return nil, nil
	}

	hostname := currentURL.Hostname()
	scheme := currentURL.Scheme

	// Normalize current path for comparison (ensure it ends with /)
	currentPath := currentURL.Path
	if !strings.HasSuffix(currentPath, "/") {
		currentPath += "/"
	}

	// Find all href attributes in anchor tags
	// The gcsweb HTML has links like: <a href="artifacts/">artifacts/</a>
	// We use a simple approach to find href="..." patterns
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Look for href patterns
		idx := strings.Index(line, `href="`)
		if idx == -1 {
			continue
		}
		start := idx + 6 // len(`href="`)
		end := strings.Index(line[start:], `"`)
		if end == -1 {
			continue
		}
		href := line[start : start+end]

		// Skip parent directory links, empty hrefs, and any path with ..
		if href == "" || href == ".." || href == "../" || strings.Contains(href, "..") {
			continue
		}
		// Skip absolute URLs to other hosts
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			continue
		}

		// Add a leading slash if it's not there
		if !strings.HasPrefix(href, "/") {
			href = "/" + href
		}

		// Construct absolute URL - href already has leading slash
		absURLStr := scheme + "://" + hostname + href
		refURL, err := url.Parse(absURLStr)
		if err != nil {
			continue
		}

		// Only include URLs that are children of the current URL (going downwards)
		if !strings.HasPrefix(refURL.Path, currentPath) {
			continue
		}

		// Categorize as directory or file
		if strings.HasSuffix(href, "/") {
			dirs = append(dirs, refURL.String())
		} else {
			files = append(files, refURL.String())
		}
	}

	return dirs, files
}

// isMustGatherFile checks if a filename matches must-gather patterns
func isMustGatherFile(filename string) bool {
	base := pathlib.Base(filename)
	// Match must-gather.tar, must-gather.tar.gz, must-gather.tar.xz, etc.
	return strings.HasPrefix(base, "must-gather") && strings.Contains(base, ".tar")
}

// FindMustGatherInProw uses breadth-first search to find the first must-gather.tar file
// It stops as soon as it finds one and returns it
func FindMustGatherInProw(gcswebURL string) (string, error) {
	fmt.Printf("Searching for must-gather files starting from %s\n", gcswebURL)

	// Queue for BFS
	queue := []string{gcswebURL}

	for len(queue) > 0 {
		// Dequeue
		currentURL := queue[0]
		queue = queue[1:]

		fmt.Printf(".")

		// Fetch the page
		resp, err := http.Get(currentURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to fetch %s: %v\n", currentURL, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			fmt.Fprintf(os.Stderr, "warning: unexpected status code %d for %s\n", resp.StatusCode, currentURL)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to read response body for %s: %v\n", currentURL, err)
			continue
		}

		dirs, files := parseGCSWebLinks(body, currentURL)

		// Check files for must-gather - return immediately if found
		for _, file := range files {
			if isMustGatherFile(file) {
				fmt.Println()
				fmt.Printf("Found must-gather: %s\n", file)
				return file, nil
			}
		}

		// Add directories to the queue for further exploration
		queue = append(queue, dirs...)
	}

	return "", fmt.Errorf("no must-gather file found")
}

func extractClientVersion(mustGatherLogsFilePath string) string {
	filePath := mustGatherLogsFilePath
	clientVersion := ""
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Initialize a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Variable to store the matching line
	var clientVersionLine string

	// Counter for the first 20 lines
	lineCount := 0

	// Read the file line by line
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Check if the line starts with "ClientVersion: "
		if strings.HasPrefix(line, "ClientVersion: ") {
			clientVersionLine = line
			break // Exit the loop as we found the line
		}

		// Stop checking after 20 lines as it should be at line 4
		if lineCount >= 20 {
			break
		}
	}

	// Handle potential scanning error
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return ""
	}

	// Check if we found the line and print the result
	if clientVersionLine != "" {
		parts := strings.Split(clientVersionLine, ":")
		if len(parts) == 2 {
			// Trim spaces and get the version part
			clientVersion = strings.TrimSpace(parts[1])
			return clientVersion
		}
	}
	return ""
}
