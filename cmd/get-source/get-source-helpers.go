package getsource

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gmeghnag/omc/vars"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

const (
	imageManifest = "application/vnd.oci.image.manifest.v1+json"
	imageIndex    = "application/vnd.oci.image.index.v1+json"
	amd64         = "amd64"
	linux         = "linux"
)

// Structs to parse JSON responses
type PullSecret struct {
	Auths map[string]struct {
		Auth string `json:"auth"`
	} `json:"auths"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func getRegistryAccessToken(registry string, repository string, authfile string) string {
	var data []byte
	var err error
	if authfile != "" {
		data, err = os.ReadFile(authfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading authentication file: %v\n", err)
			os.Exit(1)
		}
	} else {
		homeDir, _ := os.UserHomeDir()
		pullSecretPath := filepath.Join(homeDir, ".omc", "pull-secret.txt")
		data, err = os.ReadFile(pullSecretPath)
		if err != nil {
			pullSecretPath := filepath.Join(homeDir, ".omc", "pull-secret.json")
			data, err = os.ReadFile(pullSecretPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading authentication file: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// Parse JSON
	var secret PullSecret
	if err := json.Unmarshal(data, &secret); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	authBase64, ok := secret.Auths[registry]
	if !ok {
		fmt.Fprintf(os.Stderr, "error: \"%s\" auth not found in pull secret", registry)
		os.Exit(1)
	}

	// Decode base64 auth
	authDecoded, err := base64.StdEncoding.DecodeString(authBase64.Auth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding auth: %v\n", err)
		os.Exit(1)
	}

	registryAuthMethod, err := detectRegistryAuth(registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting auth method for registry \"%s\": %v\n", registry, err)
		os.Exit(1)
	}
	var url string
	if registryAuthMethod == "keycloak" {
		url = fmt.Sprintf("https://%s/auth/realms/rhcc/protocol/redhat-docker-v2/auth?scope=repository:%s:pull&service=docker-registry", registry, repository)
	} else {
		url = fmt.Sprintf("https://%s/v2/auth?service=%s&scope=repository:%s:pull", registry, registry, repository)
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(authDecoded))

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

	// Parse token response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding token response: %v\n", err)
		os.Exit(1)
	}

	return tokenResp.Token
}

// DetectRegistryAuth fa una richiesta a https://<registry>/v2/
// e ritorna "classic", "keycloak" o "none" in base al tipo di autenticazione.
func detectRegistryAuth(registry string) (string, error) {
	url := fmt.Sprintf("https://%s/v2/", registry)

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	authHeader := resp.Header.Get("Www-Authenticate")
	if authHeader == "" {
		return "none", nil
	}

	// Normalizziamo lowercase per sicurezza
	h := strings.ToLower(authHeader)

	switch {
	case strings.Contains(h, "/v2/auth"):
		return "classic", nil
	case strings.Contains(h, "/auth/realms/"):
		return "keycloak", nil
	default:
		// Se c'è l'header ma non matcha i pattern noti
		return "unknown", nil
	}
}

type Manifest struct {
	MediaType string `json:"mediaType"`
	Config    struct {
		Digest string `json:"digest"`
	} `json:"config,omitempty"`
	Manifests []ArchManifest `json:"manifests,omitempty"`
}

type ArchManifest struct {
	MediaType string   `json:"mediaType"`
	Digest    string   `json:"digest"`
	Size      int      `json:"size"`
	Platform  Platform `json:"platform"`
}

type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Variant      string `json:"variant,omitempty"` // optional
}

func getRegistryRepositoryImageDigest(image string) (string, string, string) {

	// Regex per catturare registry, repository e digest
	re := regexp.MustCompile(`^([^/]+)/(.+)@(.+)$`)
	matches := re.FindStringSubmatch(image)

	if len(matches) != 4 {
		fmt.Fprintf(os.Stderr, "Image format not valid: %s\n", image)
		os.Exit(1)
	}

	registry := matches[1]
	repository := matches[2]
	digest := matches[3]

	return registry, repository, digest
}

func getManifestDigest(registry string, repository string, token string, openshiftReleaseImageDigest string) string {
	// Get token from environment variable

	// URL for the manifest
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, repository, openshiftReleaseImageDigest)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Bearer "+token)

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

	// Parse the JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response body: %v\n", err)
		os.Exit(1)
	}

	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	if manifest.Config.Digest != "" {
		return manifest.Config.Digest
	} else if len(manifest.Manifests) != 0 {
		for _, arch := range manifest.Manifests {
			if arch.Platform.Architecture == amd64 && arch.Platform.OS == linux {
				return getManifestDigest(registry, repository, token, arch.Digest)
			}
		}
	}
	fmt.Fprintf(os.Stderr, "Error getting linux/amd64 image manifest from: %v\n", manifest)
	os.Exit(1)
	return ""
}

type Blob struct {
	Config struct {
		Labels map[string]string `json:"Labels"`
	} `json:"config"`
}

type CommitInfo struct {
	CommitUrl  string `json:"commitUrl"`
	Repository string `json:"repository"`
	Username   string `json:"username"`
	CommitId   string `json:"commitId"`
}

func getCommitUrl(registry string, repository string, token string, manifestDigest string) string {

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

	// Parse JSON
	var blob Blob
	if err := json.Unmarshal(body, &blob); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Extract the specific label
	commitURL, ok := blob.Config.Labels["io.openshift.build.commit.url"]
	if !ok {
		fmt.Fprintf(os.Stderr, "Label 'io.openshift.build.commit.url' not found in image manifest\n")
		os.Exit(1)
	}
	return commitURL
}

func parseCommitUrl(commitUrl string) (string, string, string) {
	parsedUrl, err := url.Parse(commitUrl)
	if err != nil {
		panic(err)
	}
	parts := strings.Split(strings.Trim(parsedUrl.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "commit" {
		fmt.Fprintf(os.Stderr, "Invalid commit URL format: %s\n", parsedUrl.Path)
		os.Exit(1)
	}

	username := parts[0]
	repository := parts[1]
	commit := parts[3]
	return username, repository, commit
}

// GitTree represents the root JSON structure from GitHub API
type GitTree struct {
	Tree []GitTreeEntry `json:"tree"`
}

// GitTreeEntry represents each entry in the Git tree
type GitTreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
	URL  string `json:"url"`
}

func getRedHatReleaseImage(podName string, namespaceName string, containerName string) string {
	var redHatReleaseImage string
	var containersNames []string
	if namespaceName == "" {
		namespaceName = vars.Namespace
	}
	namespacePath := fmt.Sprintf("%s/namespaces/%s", vars.MustGatherRootPath, namespaceName)
	podPath := fmt.Sprintf("%s/pods/%s/%s.yaml", namespacePath, podName, podName)
	podYaml, err := os.ReadFile(podPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: pod \"%s\" not found in namespace \"%s\".\n", podName, namespaceName)
		os.Exit(1)
	}
	var Pod v1.Pod
	if err := yaml.Unmarshal([]byte(podYaml), &Pod); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to unmarshal file %s/pods/%s/%s.yaml\n", namespacePath, podName, podName)
		os.Exit(1)
	}

	if len(Pod.Spec.Containers) == 1 && containerName == "" {
		redHatReleaseImage = Pod.Spec.Containers[0].Image
	} else {
		var containerSlice []v1.Container
		containerSlice = append(containerSlice, Pod.Spec.Containers...)
		containerSlice = append(containerSlice, Pod.Spec.InitContainers...)
		for _, c := range containerSlice {
			if containerName == c.Name {
				redHatReleaseImage = c.Image
				break
			}
			containersNames = append(containersNames, c.Name)
		}
	}
	if redHatReleaseImage == "" {
		if containerName != "" {
			fmt.Fprintf(os.Stderr, "error: container \"%s\" not found in pod \"%s\".\n", containerName, podName)
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "error: a container name must be specified for pod \"%s\", choose one of: %v\n", podName, containersNames)
			os.Exit(1)
		}
	}
	return redHatReleaseImage
}

func parseArgs(args []string) string {
	podName := args[0]
	if len(args) == 1 {
		s := strings.Split(args[0], "/")
		if len(s) == 2 {
			if strings.ToLower(s[0]) == "po" || strings.ToLower(s[0]) == "pod" || strings.ToLower(s[0]) == "pods" {
				podName = strings.ToLower(s[1])
				if podName == "" {
					fmt.Fprintf(os.Stderr, "arguments in POD/POD_NAME form must have a single resource and name\n")
					os.Exit(1)
				}
			}
		} else {
			podName = s[0]
			return podName
		}
	}
	return podName
}

//func main () {
//	token := getQuayAccessToken()
//	manifestDigest := getManifestDigest(token)
//	commitUrl := getCommitUrl(token, manifestDigest)
//	username, repository, commit:= parseCommitUrl(commitUrl)
//	searchFileInGitHubRepository(username, repository, commit, "operator.go", "")
//}
