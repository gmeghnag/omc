package stern

import (
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/cmd/logs"
	"github.com/gmeghnag/omc/vars"

	corev1 "k8s.io/api/core/v1"

	"github.com/spf13/cobra"
)

// Flag binding targets. Cobra binds flags to variable addresses, so these
// stay as package vars. RunE copies them into an Options for Run.
var (
	containerFlag  string
	selectorFlag   string
	excludeFlag    []string
	includeFlag    []string
	highlightFlag  []string
	sinceFlag      time.Duration
	tailFlag       int64
	timestampsFlag string
	colorFlag      string
	outputFlag     string
	noFollowFlag   bool
)

// Stern represents the stern command, an omc counterpart to
// https://github.com/stern/stern for tailing multiple pods/containers
// captured in a must-gather.
var Stern = &cobra.Command{
	Use:          "stern POD_QUERY",
	Short:        "Print the logs of multiple pods and containers matching a query",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		namespaceFlag, _ := cmd.Flags().GetString("namespace")
		if namespaceFlag != "" {
			vars.Namespace = namespaceFlag
		}
		opts := Options{
			Container:    containerFlag,
			Selector:     selectorFlag,
			Exclude:      excludeFlag,
			Include:      includeFlag,
			Highlight:    highlightFlag,
			Since:        sinceFlag,
			Tail:         tailFlag,
			Timestamps:   timestampsFlag,
			Color:        colorFlag,
			Output:       outputFlag,
			NoFollow:     noFollowFlag,
			AllNamespace: vars.AllNamespaceBoolVar,
		}
		return Run(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts, args)
	},
}

func Run(stdout, stderr io.Writer, opts Options, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 'stern POD_QUERY'; a pod-query regular expression is required")
	}
	if opts.Selector != "" || len(opts.Exclude) > 0 || len(opts.Include) > 0 || len(opts.Highlight) > 0 {
		return fmt.Errorf("stern: --selector, --exclude, --include and --highlight are not yet supported")
	}

	podQuery, err := regexp.Compile(args[0])
	if err != nil {
		return fmt.Errorf("failed to compile regular expression from pod-query: %w", err)
	}
	containerQuery, err := regexp.Compile(opts.Container)
	if err != nil {
		return fmt.Errorf("failed to compile regular expression for container query: %w", err)
	}

	if err := normalizeMustGatherRoot(); err != nil {
		return err
	}

	namespaces, err := resolveNamespaces(opts.AllNamespace)
	if err != nil {
		return err
	}

	originalNamespace := vars.Namespace
	defer func() { vars.Namespace = originalNamespace }()

	matched := false
	for _, namespace := range namespaces {
		pods, err := listPods(vars.MustGatherRootPath + "/namespaces/" + namespace)
		if err != nil {
			fmt.Fprintf(stderr, "stern: %s: %v\n", namespace, err)
			continue
		}
		for _, pod := range pods {
			if !podQuery.MatchString(pod.Name) {
				continue
			}
			containerNames := matchingContainers(pod, containerQuery)
			if len(containerNames) == 0 {
				continue
			}
			matched = true
			vars.Namespace = namespace
			for _, containerName := range containerNames {
				pw := newPrefixWriter(stdout, pod.Name, containerName)
				logsOpts := logs.Options{
					RootPath:  vars.MustGatherRootPath,
					Namespace: namespace,
					Tail:      opts.Tail,
				}
				if err := logs.Run(pw, stderr, logsOpts, []string{pod.Name, containerName}); err != nil {
					fmt.Fprintf(stderr, "stern: %s %s: %v\n", pod.Name, containerName, err)
				}
				if err := pw.Flush(); err != nil {
					fmt.Fprintf(stderr, "stern: %s %s: %v\n", pod.Name, containerName, err)
				}
			}
		}
	}

	if !matched {
		return fmt.Errorf("no pods found matching %q", args[0])
	}
	return nil
}

// resolveNamespaces returns the list of namespace directory names to search.
// With allNamespaces it enumerates every namespace captured in the
// must-gather; otherwise it uses the single namespace selected via -n.
func resolveNamespaces(allNamespaces bool) ([]string, error) {
	if !allNamespaces {
		if vars.Namespace == "" {
			return nil, fmt.Errorf("a namespace must be specified with -n, or use -A/--all-namespaces")
		}
		return []string{vars.Namespace}, nil
	}
	namespaces := helpers.GetNamespaces(vars.MustGatherRootPath)
	if len(namespaces) == 0 {
		return nil, fmt.Errorf("no namespaces found under %s", vars.MustGatherRootPath)
	}
	return namespaces, nil
}

// matchingContainers returns the names of the pod's init and regular
// containers whose name matches containerQuery, init containers first.
func matchingContainers(pod corev1.Pod, containerQuery *regexp.Regexp) []string {
	var names []string
	for _, c := range pod.Spec.InitContainers {
		if containerQuery.MatchString(c.Name) {
			names = append(names, c.Name)
		}
	}
	for _, c := range pod.Spec.Containers {
		if containerQuery.MatchString(c.Name) {
			names = append(names, c.Name)
		}
	}
	return names
}

func init() {
	Stern.Flags().Int64Var(&tailFlag, "tail", -1, "The number of lines from the end of the logs to show. Defaults to -1, showing all logs.")
	Stern.Flags().StringVarP(&containerFlag, "container", "c", ".*", "Container name when multiple containers in pod. (regular expression)")
	Stern.PersistentFlags().BoolVarP(&vars.AllNamespaceBoolVar, "all-namespaces", "A", false, "If present, tail across all namespaces.")

	// Most of stern options are not implemented atm
	//Stern.Flags().StringVarP(&selectorFlag, "selector", "l", "", "Selector (label query) to filter on. If present, default to \".*\" for the pod-query.")
	//Stern.Flags().StringArrayVarP(&excludeFlag, "exclude", "e", nil, "Log lines to exclude. (regular expression)")
	//Stern.Flags().StringArrayVarP(&includeFlag, "include", "i", nil, "Log lines to include. (regular expression)")
	//Stern.Flags().StringArrayVarP(&highlightFlag, "highlight", "H", nil, "Log lines to highlight. (regular expression)")
	//Stern.Flags().DurationVarP(&sinceFlag, "since", "s", 0, "Return logs newer than a relative duration like 5s, 2m, or 3h.")
	//Stern.Flags().StringVarP(&timestampsFlag, "timestamps", "t", "", "Print timestamps with the specified format. One of 'default' or 'short'.")
	//Stern.Flags().StringVar(&colorFlag, "color", "auto", "Force set color output. 'auto': colorize if tty attached, 'always': always colorize, 'never': never colorize.")
	//Stern.Flags().StringVarP(&outputFlag, "output", "o", "default", "Specify predefined template. Currently support: [default, raw, json]")
	//Stern.Flags().BoolVar(&noFollowFlag, "no-follow", true, "No-op: must-gather log data is static and is never followed.")
	//Stern.Flags().MarkHidden("no-follow")
}
