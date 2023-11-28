# `omc`: OpenShift Must-Gather Client

[![GitHub Actions Build Status](https://github.com/gmeghnag/omc/actions/workflows/build.yml/badge.svg)](https://github.com/gmeghnag/omc/actions?query=workflow%3ABuild)
![Downloads](https://img.shields.io/github/downloads/gmeghnag/omc/total)

`omc` tool has been created to allow engineers to inspect resources from a must-gather in the same way as they are retrieved with the `oc` command.

---
## Installation

### Linux / OS X
```
# cd to a directory that is in your $PATH

curl -sL https://github.com/gmeghnag/omc/releases/latest/download/omc_$(uname -o)_$(uname -m).tar.gz | tar xzf - omc && chmod +x ./omc

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
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|         ENDPOINT          |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| https://10.0.202.187:2379 | 9f38784f0a8ae43  | 3.4.14  | 147 MB  | false     | false      |        24 |    5682273 |            5682273 |        |
| https://10.0.132.49:2379  | 83b81478d4b02409 | 3.4.14  | 148 MB  | false     | false      |        24 |    5682423 |            5682423 |        |
| https://10.0.178.163:2379 | dd17c7ce8efc0349 | 3.4.14  | 147 MB  | true      | false      |        24 |    5682537 |            5682537 |        |
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```
- Retrive the prometheus alerts in `firing` or `pending` state:
```
$ omc alert rule -s firing,pending -o wide
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
