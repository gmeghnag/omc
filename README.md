## OMC
---

Inspired by [omg tool](https://github.com/kxr/o-must-gather), with `omc` we want to be able to inspect a must-gather in the same way as we inspect a cluster with the oc command.

The `omc` tool does not simply parse the yaml file, it uses the official Kubernetes and OpenShift golang types to decode yaml file to their respective objects.

---
### Supported resources and flags

To date, the `omc` tool supports the following resources:

- ClusterOperators
- ClusterVersion
- Deployments
- Nodes
- Pods
- ReplicaSets
- Services
- Routes


Supported Flags (for get command):
- -A, --all-namespaces
- -n, --namespace
- -o, --output [ json | yaml | wide | jsonpath=... ]
- --show-labels

### Usage
Point it to an extracted must-gather:
```
$ omc use </path/to/must-gather/>
```
Use it like oc:
```
$ omc get clusterVersion
$ omc get clusterOperators
$ omc project openshift-ingress
$ omc get pods -o wide
```