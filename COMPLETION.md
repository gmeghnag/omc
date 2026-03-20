# Shell Completion Guide

`omc` provides enhanced shell completion support for bash and zsh, making it easier to discover and use resource types without memorizing them.

## Features

The shell completion now supports:

- **Resource type completion** for `omc get` command - suggests resource types like `pod`, `deployment`, `service`, etc.
- **Resource type completion** for `omc describe` command
- **Smart filtering** - completions are filtered based on what you've typed
- **Support for short forms** - works with abbreviations like `po`, `deploy`, `svc`, etc.

## Installation

### Bash

To enable completion for the current session:
```bash
source <(omc completion bash)
```

To enable completion permanently, add to your `~/.bashrc`:
```bash
echo 'source <(omc completion bash)' >> ~/.bashrc
```

Or install to the system-wide completion directory (requires root):
```bash
omc completion bash | sudo tee /etc/bash_completion.d/omc > /dev/null
```

### Zsh

To enable completion for the current session:
```zsh
source <(omc completion zsh)
```

To enable completion permanently, add to your `~/.zshrc`:
```zsh
echo 'source <(omc completion zsh)' >> ~/.zshrc
```

Or for oh-my-zsh users:
```zsh
omc completion zsh > "${fpath[1]}/_omc"
```

After installing, reload your shell or run:
```zsh
compinit
```

## Usage Examples

### Get Command

```bash
# See all available resource types
omc get <TAB>

# Filter by prefix - shows pod, pods, po, poddisruptionbudget, etc.
omc get po<TAB>

# Works with common abbreviations
omc get dep<TAB>      # shows deploy, deployment, deployments, deploymentconfig, deploymentconfigs
omc get svc<TAB>      # shows svc, service, services
omc get cm<TAB>       # shows cm, configmap, configmaps
```

### Describe Command

```bash
# See all describable resource types
omc describe <TAB>

# Filter by prefix
omc describe pod<TAB>  # shows pod, pods, po
omc describe node<TAB> # shows node, nodes, no
```

## Supported Resource Types

The completion system supports all Kubernetes and OpenShift resource types, including:

**Core Resources:**
- pods (po), services (svc), nodes (no), namespaces (ns)
- configmaps (cm), secrets, serviceaccounts (sa)
- persistentvolumes (pv), persistentvolumeclaims (pvc)
- endpoints (ep), events (ev)

**Workload Resources:**
- deployments (deploy), replicasets (rs), statefulsets (sts)
- daemonsets (ds), jobs, cronjobs (cj)
- replicationcontrollers (rc)

**OpenShift Resources:**
- deploymentconfigs (dc), buildconfigs (bc), builds
- routes, projects, imagestreams, imagestreamtags
- templates

**Networking:**
- ingresses (ing), networkpolicies (netpol)
- services (svc), endpoints (ep)

**Storage:**
- storageclasses (sc), persistentvolumes (pv), persistentvolumeclaims (pvc)
- volumeattachments, csidrivers, csinodes

**RBAC:**
- roles, rolebindings, clusterroles, clusterrolebindings
- serviceaccounts (sa)

**And many more...**

## Comparison with kubectl/oc

Unlike `kubectl` and `oc` which provide basic flag completion, `omc` now provides:

1. **Positional argument completion** - suggests resource types at the right position
2. **Context-aware suggestions** - only suggests valid resource types for each command
3. **Full alias support** - all common short forms are included

## Troubleshooting

### Completions not working

1. Make sure you've sourced the completion script:
   ```bash
   source <(omc completion bash)  # for bash
   source <(omc completion zsh)   # for zsh
   ```

2. For zsh, ensure `compinit` has been called after loading the completion

3. Verify completion is loaded:
   ```bash
   complete -p omc  # for bash
   which _omc       # for zsh
   ```

### Completions are slow

The completion system reads from the embedded resource list, so performance should be instant. If you experience slowness, it may be due to other completions in your shell.

## Development

The completion functions are defined in:
- `cmd/get/completion.go` - completion for get command
- `cmd/describe/completion.go` - completion for describe command

To add completion for a new command:
1. Create a `ValidArgsFunction` that returns the list of valid arguments
2. Add the function to the command definition using `ValidArgsFunction: YourCompletionFunc`
3. Return `cobra.ShellCompDirectiveNoFileComp` to prevent file completion
