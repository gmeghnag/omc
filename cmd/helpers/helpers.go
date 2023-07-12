package helpers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gmeghnag/omc/types"
	"github.com/gmeghnag/omc/vars"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/yaml"
)

// TYPES
type Contexts []types.Context

// CONSTS
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// VARS
var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// FUNCS
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandString(length int) string {
	return StringWithCharset(length, charset)
}

func PrintTable(headers []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("   ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)
	table.Render()
}

func FormatDiffTime(diff time.Duration) string {
	if diff.Hours() > 48 {
		if diff.Hours() > 200000 {
			return "Unknown"
		}
		return strconv.Itoa(int(diff.Hours()/24)) + "d"
	}
	if diff.Hours() < 48 && diff.Hours() > 10 {
		var h float64
		h = diff.Minutes() / 60
		return strconv.Itoa(int(h)) + "h"
	}
	if diff.Minutes() > 60 {
		var hours float64
		hours = diff.Minutes() / 60
		remainMinutes := int(diff.Minutes()) % 60
		if remainMinutes > 0 {
			return strconv.Itoa(int(hours)) + "h" + strconv.Itoa(remainMinutes) + "m"
		}
		return strconv.Itoa(int(hours)) + "h"

	}
	if diff.Seconds() > 60 {
		var minutes float64
		minutes = diff.Seconds() / 60
		remainSeconds := int(diff.Seconds()) % 60
		if remainSeconds > 0 && diff.Minutes() < 4 {
			return strconv.Itoa(int(minutes)) + "m" + strconv.Itoa(remainSeconds) + "s"
		}
		return strconv.Itoa(int(minutes)) + "m"

	}
	return strconv.Itoa(int(diff.Seconds())) + "s"
}

func ExecuteJsonPath(data interface{}, jsonPathTemplate string) {
	buf := new(bytes.Buffer)
	jPath := jsonpath.New("out")
	jPath.AllowMissingKeys(false)
	jPath.EnableJSONOutput(false)
	err := jPath.Parse(jsonPathTemplate)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: error parsing jsonpath "+jsonPathTemplate+", "+err.Error())
		os.Exit(1)
	}
	jPath.Execute(buf, data)
	fmt.Print(buf)
}

func CreateConfigFile(cfgFilePath string) {
	config := types.Config{}
	file, _ := json.MarshalIndent(config, "", " ")
	err := ioutil.WriteFile(cfgFilePath, file, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func GetData(data [][]string, allNamespacesFlag bool, showLabels bool, labels string, outputFlag string, column int32, _list []string) [][]string {
	var toAppend []string
	if allNamespacesFlag == true {
		if outputFlag == "" {
			toAppend = _list[0:column] // -A
		}
		if outputFlag == "wide" {
			toAppend = _list // -A -o wide
		}
	} else {
		if outputFlag == "" {
			toAppend = _list[1:column]
		}
		if outputFlag == "wide" {
			toAppend = _list[1:] // -o wide
		}
	}

	if showLabels {
		toAppend = append(toAppend, labels)
	}
	data = append(data, toAppend)
	return data
}

func ExtractLabels(_labels map[string]string) string {
	labels := ""
	for k, v := range _labels {
		labels += k + "=" + v + ","
	}
	if labels == "" {
		labels = "<none>"
	} else {
		labels = strings.TrimRight(labels, ",")
	}
	return labels
}

func ExtractLabel(_labels map[string]string, _label string) string {
	label := ""
	for k, v := range _labels {
		if k == _label {
			return v
		}
	}
	return label
}

// doing this because of a bug who append three characthers to the first node yaml file
func ReadYaml(YamlPath string) []byte {
	var __file []byte
	_file, err := os.Open(YamlPath)
	if err != nil {
		log.Fatal(err)
	}
	defer _file.Close()

	scanner := bufio.NewScanner(_file)
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		if len(line) != 4 {
			__file = append(__file, []byte(line)...)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return __file
}

func GetAge(resourcefilePath string, resourceCreationTimeStamp v1.Time) string {
	ResourceFile, _ := os.Stat(resourcefilePath)
	t2 := ResourceFile.ModTime()
	diffTime := t2.Sub(resourceCreationTimeStamp.Time).String()
	d, _ := time.ParseDuration(diffTime)
	return FormatDiffTime(d)

}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func PrintOutput(resource interface{}, columns int16, outputFlag string, resourceName string, allNamespacesFlag bool, showLabels bool, _headers []string, data [][]string, jsonPathTemplate string) bool {
	var headers []string
	if outputFlag == "" {
		if allNamespacesFlag == true {
			headers = _headers[0:columns]
		} else {
			headers = _headers[1:columns]
		}
		if showLabels {
			headers = append(headers, "labels")
		}
		PrintTable(headers, data)
		return false
	}
	if outputFlag == "wide" {
		if allNamespacesFlag == true {
			headers = _headers
		} else {
			headers = _headers[1:]
		}
		if showLabels {
			headers = append(headers, "labels")
		}
		PrintTable(headers, data)
		return false
	}

	// TODO: de-slice single-item slice into element

	if outputFlag == "yaml" {
		y, _ := yaml.Marshal(resource)
		fmt.Println(string(y))
	}
	if outputFlag == "json" {
		j, _ := json.MarshalIndent(resource, "", "  ")
		fmt.Println(string(j))
	}
	if strings.HasPrefix(outputFlag, "jsonpath=") {
		ExecuteJsonPath(resource, jsonPathTemplate)
	}
	return false
}

func Cat(filePath string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "error: file "+filePath+" does not exist")
		os.Exit(1)
	}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: can't open file "+filePath)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fmt.Println(scanner.Text())

	}
}

func GetJsonTemplate(outputStringVar string) string {
	jsonPathTemplate := ""
	if strings.HasPrefix(outputStringVar, "jsonpath=") {
		s := outputStringVar[9:]
		if len(s) < 1 {
			fmt.Fprintln(os.Stderr, "error: template format specified but no template given")
			os.Exit(1)
		}
		jsonPathTemplate = s
	}
	return jsonPathTemplate
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func MatchLabels(labels string, selector string) bool {
	isMatching := true
	if selector == "" {
		return isMatching
	}
	selectorArray := strings.Split(selector, ",")
	labelsArray := strings.Split(labels, ",")

	for _, s := range selectorArray {
		if !strings.Contains(s, "!=") && !strings.Contains(s, "=") && !strings.Contains(s, "==") {
			s = "app=" + s
		}
		if strings.Contains(s, "!=") {
			if StringInSlice(strings.ReplaceAll(s, "!=", "="), labelsArray) {
				isMatching = false
				break
			}
		} else if strings.Contains(s, "==") {
			if !StringInSlice(strings.ReplaceAll(s, "==", "="), labelsArray) {
				isMatching = false
				break
			}
		} else if strings.Contains(s, "=") {
			if !StringInSlice(s, labelsArray) {
				isMatching = false
				break
			}
		}
	}
	return isMatching
}

func MatchLabelsFromMap(labels map[string]string, selector string) (bool, error) {
	if selector == "" {
		return true, nil
	}
	selectorArray := strings.Split(selector, ",")

	for _, s := range selectorArray {
		if strings.Contains(s, "!=") {
			split := strings.Split(s, "!=")
			if len(split) != 2 {
				return false, fmt.Errorf("invalid labels input")
			}
			key := split[0]
			val := split[1]
			value, _ := labels[key]
			if val == value {
				return false, nil
			}
		} else if strings.Contains(s, "==") {
			split := strings.Split(s, "==")
			if len(split) != 2 {
				return false, fmt.Errorf("invalid labels input")
			}
			key := split[0]
			val := split[1]
			value, isPresent := labels[key]
			if val != value || !isPresent {
				return false, nil
			}
		} else if strings.Contains(s, "=") {
			split := strings.Split(s, "=")
			if len(split) != 2 {
				return false, fmt.Errorf("invalid labels input")
			}
			key := split[0]
			val := split[1]
			value, isPresent := labels[key]
			if val != value || !isPresent {
				return false, nil
			}
		} else if !strings.Contains(s, "!=") && !strings.Contains(s, "=") && !strings.Contains(s, "==") {
			s = "app=" + s
			split := strings.Split(s, "=")
			if len(split) != 2 {
				return false, fmt.Errorf("invalid labels input")
			}
			key := split[0]
			val := split[1]
			value, _ := labels[key]
			if val != value {
				return false, nil
			}
		}
	}
	return true, nil
}

func TranslateTimestamp(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	ResourceFile, _ := os.Stat(vars.MustGatherRootPath + "/namespaces")
	t2 := ResourceFile.ModTime()
	return ShortHumanDuration(t2.Sub(timestamp.Time))
}
func ShortHumanDuration(d time.Duration) string {
	// Allow deviation no more than 2 seconds(excluded) to tolerate machine time
	// inconsistence, it can be considered as almost now.
	if seconds := int(d.Seconds()); seconds < -1 {
		return fmt.Sprintf("<invalid>")
	} else if seconds < 0 {
		return fmt.Sprintf("0s")
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}

func GetFromJsonPath(data interface{}, jsonPathTemplate string) string {
	buf := new(bytes.Buffer)
	jPath := jsonpath.New("out")
	jPath.AllowMissingKeys(false)
	jPath.EnableJSONOutput(false)
	err := jPath.Parse(jsonPathTemplate)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: error parsing jsonpath "+jsonPathTemplate+", "+err.Error())
		os.Exit(1)
	}
	jPath.Execute(buf, data)
	return buf.String()
}
