# `omc`: OpenShift Must-Gather Client

[![GitHub Actions Build Status](https://github.com/gmeghnag/omc/actions/workflows/build.yml/badge.svg)](https://github.com/gmeghnag/omc/actions?query=workflow%3ABuild)
![Downloads](https://img.shields.io/github/downloads/gmeghnag/omc/total)

`omc` tool has been created to allow engineers to inspect resources from a must-gather in the same way as they are retrieved with the `oc` command.

---
## Installation

### Download the latest binary
```
OS=Linux        # or Darwin
curl -sL https://github.com/gmeghnag/omc/releases/latest/download/omc_${OS}_x86_64.tar.gz| tar xzf - omc
chmod +x ./omc
```

### Build from source
```
$ git clone https://github.com/gmeghnag/omc.git
$ cd omc/
$ go install
```

## Upgrade
Starting with `v2.1.0` it's possible to upgrade the tool by running `omc upgrade --to=<version>`

## Usage
Point it to an **extracted** must-gather:
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
$ omc certs inspect -n openshift-dns                                                                                                                                      
configmaps/dns-default[openshift-dns] NOT a ca-bundle
configmaps/kube-root-ca.crt[openshift-dns] NOT a ca-bundle
configmaps/openshift-service-ca.crt[openshift-dns] NOT a ca-bundle
secrets/builder-dockercfg-5vht6[openshift-dns] NOT a tls secret or token secret
secrets/builder-token-pv84n[openshift-dns] - token secret (2023-05-03 09:24:48 +0000 UTC)
    "kube-apiserver-lb-signer" [] issuer="<self>" (2023-05-03 08:59:22 +0000 UTC to 2033-04-30 08:59:22 +0000 UTC)
    "kube-apiserver-localhost-signer" [] issuer="<self>" (2023-05-03 08:59:22 +0000 UTC to 2033-04-30 08:59:22 +0000 UTC)
    "kube-apiserver-service-network-signer" [] issuer="<self>" (2023-05-03 08:59:22 +0000 UTC to 2033-04-30 08:59:22 +0000 UTC)               
<...>
```
or a specific resource kind:
```
$ omc certs inspect -n openshift-kube-apiserver configmap                                                                                                                    
configmaps/aggregator-client-ca[openshift-kube-apiserver] - ca-bundle (2023-05-03 09:20:46 +0000 UTC)
    "openshift-kube-apiserver-operator_aggregator-client-signer@1683173502" [] issuer="<self>" (2023-05-04 04:11:41 +0000 UTC to 2023-06-03 04:11:42 +0000 UTC)
configmaps/bound-sa-token-signing-certs[openshift-kube-apiserver] NOT a ca-bundle
configmaps/cert-regeneration-controller-lock[openshift-kube-apiserver] NOT a ca-bundle
configmaps/check-endpoints-kubeconfig[openshift-kube-apiserver] NOT a ca-bundle
configmaps/client-ca[openshift-kube-apiserver] - ca-bundle (2023-05-03 09:21:59 +0000 UTC)
    "admin-kubeconfig-signer" [] issuer="<self>" (2023-05-03 08:59:21 +0000 UTC to 2033-04-30 08:59:21 +0000 UTC)
    "kube-csr-signer_@1683173756" [] issuer="openshift-kube-controller-manager-operator_csr-signer-signer@1683173496" (2023-05-04 04:15:55 +0000 UTC to 2023-06-03 04:15:56 +0000 UTC)
<...>
```
 
