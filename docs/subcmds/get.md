# `omc get <args> [<flags>]`
```bash
# Get commands with basic output
omc get services                          # List all services in the namespace
omc get pods --all-namespaces             # List all pods in all namespaces
omc get pods -o wide                      # List all pods in the current namespace, with more details
omc get deployment my-dep                 # List a particular deployment
omc get pods                              # List all pods in the namespace
omc get pod my-pod -o yaml                # Get a pod's YAML
```

| Output format             | Description                                                                                               | 
|---------------------------|-----------------------------------------------------------------------------------------------------------|
| `-o=json`                 | Output a JSON formatted API object                                                                        |
| `-o=jsonpath=<template>`  | Print the fields defined in a jsonpath expression                                                         |
| `-o=name`                 | Print only the resource name and nothing else                                                             | 
| `-o=wide`                 | Output in the plain-text format with any additional information, and for pods, the node name is included  | 
| `-o=yaml`                 | Output a YAML formatted API object                                                                        | 
| `-o=custom-columns`       | Allows a user to customise the fields that are output and their corresponding header names                | 
