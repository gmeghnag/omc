# `omc`: OpenShift Must-Gather Client

[![GitHub Actions Test Status](https://github.com/gmeghnag/omc/actions/workflows/test.yml/badge.svg)](https://github.com/gmeghnag/omc/actions?query=workflow%3ATest) [![GitHub Actions Build Status](https://github.com/gmeghnag/omc/actions/workflows/build.yml/badge.svg)](https://github.com/gmeghnag/omc/actions?query=workflow%3ABuild) ![Go version](https://img.shields.io/github/go-mod/go-version/gmeghnag/omc)
![Downloads](https://img.shields.io/github/downloads/gmeghnag/omc/total)

`omc` inspects an OpenShift **must-gather** offline, letting you explore its resources with the same commands you'd use against a live cluster with `oc` — no API server required.

## What it can do

- **`get`** — read any resource like `oc get`: tables, `-o wide|yaml|json|jsonpath|custom-columns`, label selectors (`-l`), `-A`, `--sort-by`. CRDs in the must-gather are discovered automatically.
- **`describe`** — `oc describe`-style detail for pods and nodes.
- **`logs`** — container logs for a pod; **`stern`** tails many pods/containers at once by regex.
- **`events`** — events sorted by time.
- **`etcd`** — etcd member status and health.
- **`prometheus`** (`prom`) — alerts and rules captured in the must-gather.
- **`certs`** — inspect certificates found in configmaps, secrets and CSRs.
- **`haproxy`** — router backends from the ingress-controller config.
- **`network connectivity`** — summarize `PodNetworkConnectivityCheck` resources.
- **node-logs, machineconfig, ovn, ceph, insights, get-source** and more OpenShift-specific helpers.
- **Multiple must-gathers at once** — query several captures as one, newest wins (see below).
- **Parallel sessions** — isolate independent analyses with `OMCCONFIG`.

---

## Installation

### Linux / macOS
```
# cd to a directory that is in your $PATH
curl -sL "https://github.com/gmeghnag/omc/releases/latest/download/omc_$(uname)_$(uname -m).tar.gz" | tar xzf - omc && chmod +x ./omc
omc -h
```
**Note:** macOS may block the downloaded binary until it is approved via `System Settings` → `Privacy & Security`.

### Other operating systems
Download the asset matching your OS from the [latest release](https://github.com/gmeghnag/omc/releases/latest), unpack it, and move the `omc` binary somewhere in your `$PATH`.

### Build from source
```
git clone https://github.com/gmeghnag/omc.git
cd omc/ && go install
```

### Upgrade
Since `v2.1.0` you can self-upgrade: `omc upgrade --to=<version>`.

---

## Quick start

Point `omc` at a must-gather — an extracted directory, a local tarball, or a remote tarball URL:
```
omc use </path/to/must-gather/>
```
Then use it like `oc`:
```
omc get clusterversion
omc get pods -o wide -l app=etcd -n openshift-etcd
omc describe node ip-10-0-132-49.eu-central-1.compute.internal
```

---

## Multiple must-gathers at once

Point `omc use` at a directory that contains **several** must-gathers and it discovers them all and groups them into one context (a single must-gather still works exactly as before):
```
omc use /cases/12345/           # holds must-gather-1/, must-gather-2/, ...
omc get pods -A
```

The captures are ordered by collection time (their `timestamp` file), **most recent first**, and that order is the rule for everything:

> **Resources from the most recent must-gather win.** Each resource is identified by its `metadata.uid`.

- **`get`** queries every must-gather, **unions** the results and **deduplicates by `uid`**: when the same resource appears in more than one capture, only the copy from the **most recent** one is shown; resources that exist in just one capture are all kept.
- **`describe`** and **`logs`** read a resource from the **same** capture `get` would show — one consistent source of truth, never mixing data from different captures. For `logs`, if that pod has no logs in its winning capture, none are shown (no fallback to an older one).
- **`events`** are merged across all captures, deduplicated by `uid` (keeping the freshest copy), and shown on a single time-sorted timeline.

`omc use` prints the grouped captures; the one marked `*` is the most recent and backs `describe`/`logs`.

---

## Parallel sessions (`OMCCONFIG`)

By default omc keeps its state in `~/.omc/omc.json`. To analyze unrelated cases side by side without clashing, give each session its own config file:
```bash
# terminal 1
export OMCCONFIG=~/.omc/case-12345.json
omc use /cases/12345/must-gather && omc get nodes

# terminal 2 (in parallel)
export OMCCONFIG=~/.omc/case-67890.json
omc use /cases/67890/must-gather && omc get pods -A
```
A one-off override is also available with the `--omcconfig` flag. CRDs (`~/.omc/customresourcedefinitions/`) and pull secrets are shared across configs, so only the per-case context is isolated.

Precedence: `--omcconfig` flag → `OMCCONFIG` env var → `~/.omc/omc.json`.

---

## Custom Resource Definitions (CRDs)

omc discovers CRDs bundled in the must-gather under `cluster-scoped-resources/apiextensions.k8s.io/customresourcedefinitions/`. For CRDs not present there, drop their definitions into `~/.omc/customresourcedefinitions/` and omc will render those resources too:
```
BASE=~/.omc/customresourcedefinitions && mkdir -p $BASE
curl -sL https://raw.githubusercontent.com/NVIDIA/gpu-operator/main/config/crd/bases/nvidia.com_clusterpolicies.yaml -o $BASE/clusterpolicies.nvidia.com.yaml

omc get clusterpolicy
# NAME             STATUS   AGE
# cluster-policy   ready    52d
```

---

## More examples

Resources behave like `oc`, including selectors and jsonpath:
```
omc get node -l node-role.kubernetes.io/master= -o name
omc get pod -l app=etcd -o jsonpath="{.items[?(@.spec.nodeName=='ip-10-0-132-49...')].metadata.name}"
```

etcd member status:
```
omc etcd status
+----------------------------+------------------+---------+----------------+-----------+...
|          ENDPOINT          |        ID        | VERSION | DB SIZE/IN USE | IS LEADER |...
| https://10.44.134.165:2379 | 1763488a02d62c90 | 3.5.9   | 133 MB/90 MB   | true      |...
```

Prometheus alerts/rules that are `firing` or `pending`:
```
omc prom rules -s firing,pending -o wide
GROUP             RULE              STATE    AGE   ALERTS   ACTIVE SINCE
cluster-version   UpdateAvailable   firing   11s   1        27 Jan 22 14:32 UTC
general.rules     Watchdog          firing   11s   1        25 Jan 22 08:50 UTC
```

Certificates inside configmaps/secrets/CSRs:
```
omc certs inspect
NAME               KIND        AGE   CERTTYPE    SUBJECT                                    NOTAFTER
kube-root-ca.crt   ConfigMap   47h   ca-bundle   CN=kube-apiserver-lb-signer,OU=openshift   2033-04-30 08:59:22 +0000 UTC
```

HAProxy router backends from the ingress config:
```
omc haproxy backends
NAMESPACE   NAME                       INGRESSCONTROLLER   SERVICES                   PORT        TERMINATION
testdata    rails-postgresql-example   default             rails-postgresql-example   web(8080)   http
```

`PodNetworkConnectivityCheck` summary (`--wide` for full failure text, `--unhealthy-only` to list only failures):
```
omc network connectivity
NAME                             TARGET                  REACHABLE   LAST_FAILURE_REASON
network-check-source-apiserver   https://10.0.0.1:6443   True
```

Tail logs of many pods/containers at once — a must-gather counterpart of [stern](https://github.com/stern/stern). `POD_QUERY` is a regex on pod names; scope with `-n` or `-A`, pick containers with `-c` (default `.*`), limit lines with `--tail`:
```
omc stern -n openshift-etcd etcd
etcd-master-0 etcd    {"level":"info","msg":"serving client traffic securely"}
etcd-master-1 etcd    {"level":"info","msg":"serving client traffic securely"}
```
