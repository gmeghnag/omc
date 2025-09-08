package getsource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
)

var GetManifest = &cobra.Command{
	Use:   "manifest",
	Short: "Retrieve OpenShift container image manifest.",
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
		manifest := getManifest(registry, repository, token, manifestDigest)
		fmt.Println(string(manifest))
	},
}

func getManifest(registry string, repository string, token string, manifestDigest string) string {

	// URL for the blob
	url := fmt.Sprintf("https://%s/v2/%s/blobs/%s", registry, repository, manifestDigest)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{
		// Follow redirects like `-L` in curl
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
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

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response body: %v\n", err)
		os.Exit(1)
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshal response body: %v\n", err)
		os.Exit(1)
	}

	// Encode back to JSON (pretty printed)
	output, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal response body: %v\n", err)
		os.Exit(1)
	}

	return string(output)
}

func init() {
	GetManifest.PersistentFlags().StringVarP(&vars.Container, "container", "c", "", "Return the manifest for the specified container.")
	GetManifest.PersistentFlags().StringVar(&Image, "image", "", "Return the manifest for the specified image.")
	GetManifest.PersistentFlags().StringVar(&AuthFile, "authfile", "", "Path of the authentication file.")
}
