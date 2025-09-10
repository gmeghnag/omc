package getsource

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
)

var OpenInBrowser bool
var ExcludePrefix string

var GetCode = &cobra.Command{
	Use:   "code",
	Short: "Retrieve OpenShift container image code.",
	Run: func(cmd *cobra.Command, args []string) {
		if (len(args) == 0 && Image == "") || len(args) > 2 {
			fmt.Fprintln(os.Stderr, "error: expected 'omc source manifest (POD_NAME | POD/POD_NAME) [-n NAMESPACE] [-c CONTAINER]'.")
			fmt.Fprintln(os.Stderr, "POD or POD/POD_NAME is a required argument for the \"source manifest\" command (if --image is not provided)")
			fmt.Fprintln(os.Stderr, "See 'omc source manifest -h' for help and examples")
			os.Exit(1)
		}
		var registry, repository, imageDigest string
		if Image != "" {
			registry, repository, imageDigest = getRegistryRepositoryImageDigest(Image)
		} else {
			podName := parseArgs(args)
			namespaceName, _ := cmd.Flags().GetString("namespace")
			Image = getRedHatReleaseImage(podName, namespaceName, vars.Container)
			registry, repository, imageDigest = getRegistryRepositoryImageDigest(Image)
		}
		token := getRegistryAccessToken(registry, repository, AuthFile)
		manifestDigest := getManifestDigest(registry, repository, token, imageDigest)
		commitUrl := getCommitUrl(registry, repository, token, manifestDigest)
		var filesFound []string
		username, repository, commit := parseCommitUrl(commitUrl)
		if FileName != "" {
			filesFound = searchFileInGitHubRepository(username, repository, commit, FileName, ExcludePrefix)
		} else {
			filesFound = append(filesFound, fmt.Sprintf("https://github.com/%s/%s/tree/%s", username, repository, commit))
		}
		if len(filesFound) == 0 {
			fmt.Fprintf(os.Stderr, "No match found for filename \"%s\" in repository \"https://github.com/%s/%s/blob/%s\"\n", FileName, username, repository, commit)
			os.Exit(1)
		}
		if OpenInBrowser {
			if len(filesFound) > 1 {
				p := tea.NewProgram(initialModel(filesFound))
				if _, err := p.Run(); err != nil {
					fmt.Printf("Error running program: %v\n", err)
					os.Exit(1)
				}
			} else {
				openBrowser(filesFound[0])
			}
		} else {
			for _, file := range filesFound {
				fmt.Println(file)
			}
		}
	},
}

func init() {
	GetCode.PersistentFlags().StringVarP(&vars.Container, "container", "c", "", "Return the code for the specified container.")
	GetCode.PersistentFlags().StringVar(&Image, "image", "", "Return the code for the specified image.")
	GetCode.PersistentFlags().StringVar(&AuthFile, "authfile", "", "Path of the authentication file.")
	GetCode.PersistentFlags().BoolVar(&OpenInBrowser, "open", false, "Open the codebase file in your browser.")
	GetCode.PersistentFlags().StringVarP(&FileName, "filename", "f", "", "File (or file:<line_number>) to search in the repository.")
	GetCode.PersistentFlags().StringVar(&ExcludePrefix, "exclude", "", "Exclude the prefix from the file search.")
}

func searchFileInGitHubRepository(username string, repository string, commit string, filename string, exclude string) []string {
	// Make HTTP GET request
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", username, repository, commit)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	// GitHub requires a User-Agent header
	req.Header.Set("User-Agent", "omc")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP request error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Unexpected status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response body: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var treeResp GitTree
	if err := json.Unmarshal(body, &treeResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}
	var lineNumber string
	parts := strings.SplitN(filename, ":", 2)
	if len(parts) == 2 {
		filename = parts[0]
		lineNumber = parts[1]
	}
	var matches []string
	var githubRawsMatches []string
	for _, item := range treeResp.Tree {
		if exclude != "" && strings.HasPrefix(item.Path, strings.TrimLeft(exclude, "/")) {
			continue
		}
		if strings.HasSuffix(item.Path, "/"+filename) {
			matches = append(matches, fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s", username, repository, commit, item.Path))
		}
	}
	if lineNumber != "" {
		lineNumberInt, err := strconv.Atoi(lineNumber)
		if err != nil {
			fmt.Printf("error converting \"%s\" to int\n", lineNumber)
			os.Exit(1)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		apiToRaw := make(map[string]string)
		for _, rawUrl := range matches {
			api := rawUrl
			rawUrl = strings.Replace(rawUrl, "github.com", "raw.githubusercontent.com", 1)
			rawUrl = strings.Replace(rawUrl, "/blob", "", 1)
			githubRawsMatches = append(githubRawsMatches, rawUrl)
			apiToRaw[rawUrl] = api
		}
		filtered := isLineNumberInFile(ctx, githubRawsMatches, 1, lineNumberInt)
		var filteredMatches []string
		for _, rawMatch := range filtered {
			filteredMatches = append(filteredMatches, apiToRaw[rawMatch]+"#L"+lineNumber)
		}
		return filteredMatches
	}
	return matches
}

type model struct {
	cursor   int
	choices  []string
	selected string
	done     bool
}

func initialModel(stringList []string) model {
	return model{
		choices: stringList,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		// Quit with q or ctrl+c
		case "q", "ctrl+c":
			return m, tea.Quit

		// Move cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// Move cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// Select item with enter
		case "enter":
			m.selected = m.choices[m.cursor]
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	// If user is done, show result
	if m.done {
		openBrowser(m.selected)
		return m.selected
	}

	// Otherwise, show menu
	s := "Select the file to open in the browser:\n\n"

	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\n↑/↓ to move, enter to select, q to quit.\n"
	return s
}

// openBrowser tries to open the URL in the default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// "start" is a built-in command, so we need cmd /c
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		// macOS uses "open"
		cmd = exec.Command("open", url)
	default: // "linux", "freebsd", etc.
		// Linux/BSD usually have xdg-open available
		cmd = exec.Command("xdg-open", url)
	}

	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open browser: %v\n", err)
		os.Exit(1)
	}
}

// checkLines returns (url, true) if the file has AT LEAST lineNumber lines
func checkLines(ctx context.Context, url string, lineNumber int) (string, bool) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("fetch error for %s: %v\n", url, err)
		return "", false
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	count := 0
	for scanner.Scan() {
		count++
		if count >= lineNumber {
			return url, true // ile has enough lines
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("scan error for %s: %v\n", url, err)
		return "", false
	}

	return "", false // file ended before reaching lineNumber
}

// FilterURLsWithFewerThan200Lines returns a slice of URLs that have <200 lines
func isLineNumberInFile(ctx context.Context, urls []string, maxWorkers int, lineNumber int) []string {
	var wg sync.WaitGroup
	jobs := make(chan string)
	results := make(chan string, len(urls))

	// Worker pool
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range jobs {
				if res, ok := checkLines(ctx, url, lineNumber); ok {
					results <- res
				}
			}
		}()
	}

	// Feed jobs
	go func() {
		for _, url := range urls {
			jobs <- url
		}
		close(jobs)
	}()

	// Close results after all workers finish
	go func() {
		wg.Wait()
		close(results)
	}()
	// Collect results
	var filtered []string
	for res := range results {
		filtered = append(filtered, res)
	}
	return filtered
}
