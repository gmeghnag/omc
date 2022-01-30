# `omc`: OpenShift Must-Gather Client

[![GitHub Actions Build Status](https://github.com/gmeghnag/omc/actions/workflows/build.yml/badge.svg)](https://github.com/gmeghnag/omc/actions?query=workflow%3ABuild)

Inspired by [omg tool](https://github.com/kxr/o-must-gather), with `omc` you can inspect a must-gather in the same way as we inspect a cluster with the oc command.

The `omc` tool does not simply parse yaml files, it uses the official Kubernetes and OpenShift golang types to decode yaml files to their respective OpenShift resources.

---
### Usage
Point it to an extracted must-gather:
```
$ omc use </path/to/must-gather/>
```
Use it like oc:
```
$ omc get clusterversion
$ omc get pods -o wide -l app=etcd -n openshift-etcd
```

### Example
```
$ omc get node -l node-role.kubernetes.io/master= -o name   
node/ip-10-0-132-49.eu-central-1.compute.internal
node/ip-10-0-178-163.eu-central-1.compute.internal
node/ip-10-0-202-187.eu-central-1.compute.internal

$ omc get pod -l app=etcd -o jsonpath="{.items[?(@.spec.nodeName=='ip-10-0-132-49.eu-central-1.compute.internal')].metadata.name}"
etcd-ip-10-0-132-49.eu-central-1.compute.internal

$ omc etcd status
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|         ENDPOINT          |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| https://10.0.202.187:2379 | 9f38784f0a8ae43  | 3.4.14  | 147 MB  | false     | false      |        24 |    5682273 |            5682273 |        |
| https://10.0.132.49:2379  | 83b81478d4b02409 | 3.4.14  | 148 MB  | false     | false      |        24 |    5682423 |            5682423 |        |
| https://10.0.178.163:2379 | dd17c7ce8efc0349 | 3.4.14  | 147 MB  | true      | false      |        24 |    5682537 |            5682537 |        |
+---------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```