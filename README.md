# `omc`: OpenShift Must-Gather Client

[![GitHub Actions Test Status](https://github.com/gmeghnag/omc/actions/workflows/test.yml/badge.svg)](https://github.com/gmeghnag/omc/actions?query=workflow%3ATest) [![GitHub Actions Build Status](https://github.com/gmeghnag/omc/actions/workflows/build.yml/badge.svg)](https://github.com/gmeghnag/omc/actions?query=workflow%3ABuild) ![Go version](https://img.shields.io/github/go-mod/go-version/gmeghnag/omc)
![Downloads](https://img.shields.io/github/downloads/gmeghnag/omc/total)




`omc` tool has been created to allow engineers to inspect resources from a must-gather in the same way as they are retrieved with the `oc` command.

---
## Installation

### Linux / OS X
```
# cd to a directory that is in your $PATH

curl -sL "https://github.com/gmeghnag/omc/releases/latest/download/omc_$(uname)_$(uname -m).tar.gz" | tar xzf - omc && chmod +x ./omc

omc -h
```
**Note:** OS X may block the downloaded omc binary until it is approved via `System Settings` -> `Privacy & Security`.

### Other Operating systems
1. View the available downloads from the latest releases page
1. Chose and download the Asset that best suits your operating system
1. Un zip/tar the binary and move it to a directory location that is in your executable path.

### Build from source
```
$ git clone https://github.com/gmeghnag/omc.git
$ cd omc/
$ go install
```

## Upgrade
Starting with `v2.1.0` it's possible to upgrade the tool by running `omc upgrade --to=<version>`

## Usage
Point it to a must-gather. This can be a local extracted must-gather, a local tarball, or a remote tarball:
```
$ omc use </path/to/must-gather/>
```
Use it like `oc`:
```
$ omc get clusterversion
$ omc get pods -o wide -l app=etcd -n openshift-etcd
```

### Parallel Usage with Multiple Config Files

By default, omc stores configuration in `~/.omc/omc.json`. To work on multiple cases in parallel, you can use different config files:

#### Using Environment Variable (Recommended)
```bash
# Terminal 1 - Working on case 12345
export OMCCONFIG=~/.omc/case-12345.json
omc use /cases/12345/must-gather
omc get nodes

# Terminal 2 - Working on case 67890 (in parallel)
export OMCCONFIG=~/.omc/case-67890.json
omc use /cases/67890/must-gather
omc get pods -A
```

#### Using Command-Line Flag
```bash
# One-off command with custom config
omc --omcconfig=~/.omc/test.json use /path/to/must-gather
```

#### How It Works
- **Config files**: Each case has its own config file (e.g., `~/.omc/case-12345.json`)
- **CRDs**: Shared across all configs (stored in `~/.omc/customresourcedefinitions/`)
- **Pull secrets**: Shared across all configs (stored in `~/.omc/pull-secret.txt`)

This design allows parallel analysis of different must-gathers without race conditions, while sharing common resources like CRDs and container registry credentials.

**Configuration Precedence:**
1. `--omcconfig` flag (highest priority)
2. `OMCCONFIG` environment variable
3. `~/.omc/omc.json` (default)

### Examples
- Retrieving master nodes by label:
```
$ omc get node -l node-role.kubernetes.io/master= -o name
node/ip-10-0-132-49.eu-central-1.compute.internal
node/ip-10-0-178-163.eu-central-1.compute.internal
node/ip-10-0-202-187.eu-central-1.compute.internal
```
- Retrieving etcd pod name from node name:
```
$ omc get pod -l app=etcd -o jsonpath="{.items[?(@.spec.nodeName=='ip-10-0-132-49.eu-central-1.compute.internal')].metadata.name}"
etcd-ip-10-0-132-49.eu-central-1.compute.internal
```
- Check the ETCD status:
```
$ omc etcd status
+----------------------------+------------------+---------+----------------+----------+-----------+------------+-----------+------------+--------------------+--------+
|          ENDPOINT          |        ID        | VERSION | DB SIZE/IN USE | NOT USED | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+----------------------------+------------------+---------+----------------+----------+-----------+------------+-----------+------------+--------------------+--------+
| https://10.44.134.165:2379 | 1763488a02d62c90 | 3.5.9   | 133 MB/90 MB   | 33%      | true      | false      |         7 |    2111896 |            2111896 |        |
| https://10.44.135.227:2379 | 96e0b13f9c1287ea | 3.5.9   | 123 MB/90 MB   | 27%      | false     | false      |         7 |    2111896 |            2111896 |        |
| https://10.44.135.186:2379 | bbdf013955819908 | 3.5.9   | 125 MB/90 MB   | 28%      | false     | false      |         7 |    2111896 |            2111896 |        |
+----------------------------+------------------+---------+----------------+----------+-----------+------------+-----------+------------+--------------------+--------+
```
- Retrive the prometheus alerts in `firing` or `pending` state:
```
$ omc prom rules -s firing,pending -o wide
GROUP                        RULE                                 STATE     AGE   ALERTS   ACTIVE SINCE
cluster-version              UpdateAvailable                      firing    11s   1        27 Jan 22 14:32 UTC
logging_fluentd.alerts       FluentdQueueLengthIncreasing         pending   27s   1        29 Jan 22 11:48 UTC
general.rules                Watchdog                             firing    11s   1        25 Jan 22 08:50 UTC
openshift-kubernetes.rules   AlertmanagerReceiversNotConfigured   firing    5s    1        25 Jan 22 08:51 UTC
```
- Retreive details of any certificate contained in configmaps/secrets/certificatesigningrequests:
```
$ omc certs inspect
NAME                       KIND        AGE   CERTTYPE    SUBJECT                                             NOTBEFORE                       NOTAFTER
kube-root-ca.crt           ConfigMap   47h   ca-bundle   CN=kube-apiserver-lb-signer,OU=openshift            2023-05-03 08:59:22 +0000 UTC 　2033-04-30 08:59:22 +0000 UTC
kube-root-ca.crt           ConfigMap   47h   ca-bundle   CN=kube-apiserver-localhost-signer,OU=openshift     2023-05-03 08:59:22 +0000 UTC 　2033-04-30 08:59:22 +0000 UTC
kube-root-ca.crt           ConfigMap   47h   ca-bundle   CN=*.apps.example.com                               2023-05-03 09:20:57 +0000 UTC 　2025-05-02 09:20:58 +0000 UTC
kube-root-ca.crt           ConfigMap   47h   ca-bundle   CN=ingress-operator@1683105658                      2023-05-03 09:20:57 +0000 UTC 　2025-05-02 09:20:58 +0000 UTC
openshift-service-ca.crt   ConfigMap   47h   ca-bundle   CN=openshift-service-serving-signer@1683105630      2023-05-03 09:20:29 +0000 UTC 　2025-07-01 09:20:30 +0000 UTC
builder-token-9f5cx        Secret      47h   ca-bundle   CN=kube-apiserver-lb-signer,OU=openshift            2023-05-03 08:59:22 +0000 UTC 　2033-04-30 08:59:22 +0000 UTC
builder-token-9f5cx        Secret      47h   ca-bundle   CN=*.apps.example.com                               2023-05-03 09:20:57 +0000 UTC 　2025-05-02 09:20:58 +0000 UTC
builder-token-9f5cx        Secret      47h   ca-bundle   CN=ingress-operator@1683105658                      2023-05-03 09:20:57 +0000 UTC 　2025-05-02 09:20:58 +0000 UTC
<...>
```
- Retreive HAProxy backends (of any namespace) from the ingresscontroller (HAProxy) config in the must-gather:
```
$ omc haproxy backends
NAMESPACE       NAME                            INGRESSCONTROLLER    SERVICES                        PORT            TERMINATION
testdata        rails-postgresql-example        default              rails-postgresql-example        web(8080)       http
other-testdata  hello-node-secure               default              hello-node                      8080            edge/Redirect
```

### Working with Custom Resource Definitions (CRDs)

omc discovers CRD definitions automatically if they are present in the must-gather under `cluster-scoped-resources/apiextensions.k8s.io/customresourcedefinitions/`. If the must-gather does not include them, you can place them in `~/.omc/customresourcedefinitions/`:
```
$ BASE=~/.omc/customresourcedefinitions
$ mkdir -p $BASE
$ curl -sL https://raw.githubusercontent.com/NVIDIA/gpu-operator/main/config/crd/bases/nvidia.com_clusterpolicies.yaml -o $BASE/clusterpolicies.nvidia.com.yaml
$ curl -sL https://raw.githubusercontent.com/NVIDIA/k8s-nim-operator/main/config/crd/bases/apps.nvidia.com_nimservices.yaml -o $BASE/nimservices.apps.nvidia.com.yaml
$ curl -sL https://raw.githubusercontent.com/NVIDIA/k8s-nim-operator/main/config/crd/bases/apps.nvidia.com_nimcaches.yaml -o $BASE/nimcaches.apps.nvidia.com.yaml
```

Example with CRDs not included in omc by default:
```
$ omc get clusterpolicy
NAME             STATUS   AGE
cluster-policy   ready    52d

$ omc get nimservice -A
NAMESPACE   NAME                                     STATUS     AGE
llms        meta-llama-3-3-70b-instruct-fp8-public   NotReady   49d

$ omc get nimcache -A
NAMESPACE   NAME                              STATUS   PVC                                   AGE
llms        meta-llama-3-3-70b-instruct-fp8   Ready    meta-llama-3-3-70b-instruct-fp8-pvc   50d
```
- Summarize **PodNetworkConnectivityCheck** resources from `pod_network_connectivity_check/podnetworkconnectivitychecks.yaml`:
```
$ omc network connectivity
NAME                                                        TARGET                    REACHABLE   LAST_FAILURE_REASON   LAST_FAILURE_MESSAGE
network-check-source-apiserver                              https://10.0.0.1:6443     True
network-check-source-to-load-balancer-api-internal          10.10.10.1:8080           True        TCPConnectError       network-check-source-to-load-balancer-api-internal: failed to establish a TCP connection ...
```
Use `omc network connectivity --wide` (or `-w`) for full failure text and the Reachable condition’s last transition time; use `--unhealthy-only` to list only checks where Reachable is not True. Filter by namespace with the root flag, for example `omc network connectivity -n openshift-network-diagnostics`.
- Tail the logs of multiple pods and containers at once (a must-gather counterpart of [stern](https://github.com/stern/stern)). `POD_QUERY` is a regular expression matched against pod names; scope the search to one namespace with `-n` or to every namespace with `-A`. Each output line is prefixed with `<pod> <container>`:
```
$ omc stern -n openshift-etcd etcd
etcd-master-0 etcd {"level":"info","ts":"...","msg":"serving client traffic securely"}
etcd-master-0 etcdctl {"level":"info","ts":"...","msg":"compacted"}
etcd-master-1 etcd {"level":"info","ts":"...","msg":"serving client traffic securely"}
```
Use `-c` to select containers by regular expression (defaults to `.*`, all containers) and `--tail N` to limit the number of lines shown per container. Only `--selector`, `--exclude`, `--include` and `--highlight` are not yet supported.
