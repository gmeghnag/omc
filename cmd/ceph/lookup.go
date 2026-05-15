package ceph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func BuildFilename(prefix string, args []string) string {
	if len(args) == 0 {
		return prefix
	}
	return prefix + "_" + strings.Join(args, "_")
}

func ParseArgs(args []string) (commandArgs []string, format string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--output" || arg == "--format" || arg == "-o" {
			if i+1 < len(args) {
				format = args[i+1]
				i++
				continue
			}
		}
		if strings.HasPrefix(arg, "--output=") {
			format = strings.TrimPrefix(arg, "--output=")
			continue
		}
		if strings.HasPrefix(arg, "--format=") {
			format = strings.TrimPrefix(arg, "--format=")
			continue
		}
		if strings.HasPrefix(arg, "-o=") {
			format = strings.TrimPrefix(arg, "-o=")
			continue
		}
		commandArgs = append(commandArgs, arg)
	}
	return
}

func LookupAndPrint(mustGatherRoot string, prefix string, args []string) {
	commandArgs, format := ParseArgs(args)
	filename := BuildFilename(prefix, commandArgs)

	var dir, fullFilename string
	if format == "json" || format == "json-pretty" {
		dir = filepath.Join(mustGatherRoot, "ceph", "must_gather_commands_json_output")
		fullFilename = filename + "_--format_json-pretty"
	} else {
		dir = filepath.Join(mustGatherRoot, "ceph", "must_gather_commands")
		fullFilename = filename
	}

	filePath := filepath.Join(dir, fullFilename)
	data, err := os.ReadFile(filePath)
	if err == nil {
		fmt.Print(string(data))
		return
	}

	// Try alternate filename for "ceph config show <param>" -> "config_<param>"
	if prefix == "ceph" && len(commandArgs) >= 3 && commandArgs[0] == "config" && commandArgs[1] == "show" {
		altFilename := "config_" + strings.Join(commandArgs[2:], "_")
		altPath := filepath.Join(dir, altFilename)
		data, err = os.ReadFile(altPath)
		if err == nil {
			fmt.Print(string(data))
			return
		}
	}

	SuggestCommands(mustGatherRoot, prefix, commandArgs)
}

func SuggestCommands(mustGatherRoot string, prefix string, args []string) {
	cmdDir := filepath.Join(mustGatherRoot, "ceph", "must_gather_commands")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: no Ceph data found in this must-gather.")
		fmt.Fprintln(os.Stderr, "This must-gather may not be from an ODF/OCS cluster.")
		os.Exit(1)
	}

	searchPrefix := prefix
	if len(args) > 0 {
		searchPrefix = BuildFilename(prefix, args)
	}

	var suggestions []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, searchPrefix) {
			suggestions = append(suggestions, FilenameToCommand(name))
		}
	}
	sort.Strings(suggestions)

	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "Error: command output not found for: %s\n", prefix+" "+strings.Join(args, " "))
	} else {
		fmt.Fprintf(os.Stderr, "Error: no arguments provided.\n")
	}

	if len(suggestions) > 0 {
		fmt.Fprintln(os.Stderr, "\nAvailable commands in this must-gather:")
		for _, s := range suggestions {
			fmt.Fprintf(os.Stderr, "  omc %s\n", s)
		}
	}
	os.Exit(1)
}

func FilenameToCommand(filename string) string {
	name := strings.TrimSuffix(filename, "_--format_json-pretty")
	return strings.ReplaceAll(name, "_", " ")
}
