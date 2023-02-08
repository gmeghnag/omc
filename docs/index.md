# `omc`: OpenShift Must-Gather Client :fontawesome-brands-golang:

[![GitHub Actions Build Status](https://github.com/gmeghnag/omc/actions/workflows/build.yml/badge.svg)](https://github.com/gmeghnag/omc/actions?query=workflow%3ABuild)
![Downloads](https://img.shields.io/github/downloads/gmeghnag/omc/total)

`omc` tool has been created to allow engineers to inspect resources from a must-gather in the same way as they are retrieved with the `oc` command.

---

## How to use?
- Collect a [must-gather](https://github.com/openshift/must-gather) via `oc adm must-gather` or `oc adm inspect <args>`:
```
# oc adm inspect ns/default
Gathering data for ns/default...
Wrote inspect data to inspect.local.7055944464130050702.
```
- Point it to an **extracted** must-gather:
```
$ omc use inspect.local.7055944464130050702
```
- Use it like `oc`:
```
omc get clusterversion
omc get pods -o wide -l app=etcd -n openshift-etcd
```
