# `omc alert`
It allows to retrieve the alerting rules [<sup class="omc-apex">1</sup>](https://docs.openshift.com/container-platform/4.12/monitoring/monitoring-overview.html#:~:text=Alerting%20rules)[<sup>2</sup>](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) from `Prometheus` built-in component.

## `groups`
Retrieve `PrometheusRules` grouped by component:
```yaml
omc prometheus alertgroup insights -oyaml
```

1.  :man_raising_hand: I'm a code annotation! I can contain `code`, __formatted
    text__, images, ... basically anything that can be written in Markdown.

<details>
<summary>OUTPUT</summary>

```yaml
data:
  groups:
  - evaluationTime: 0.000948389
    file: /etc/prometheus/rules/prometheus-k8s-rulefiles-0/openshift-insights-insights-prometheus-rules-1b383bfe-cf4a-4c42-8cd2-16e1da9546cf.yaml
    interval: 30
    lastEvaluation: "2023-02-07T10:16:52.230581232Z"
    limit: 0
    name: insights
    rules:
    - alerts: []
      annotations:
        description: 'Insights operator is disabled. In order to enable Insights and
          benefit from recommendations specific to your cluster, please follow steps
          listed in the documentation: https://docs.openshift.com/container-platform/latest/support/remote_health_monitoring/enabling-remote-health-reporting.html'
        summary: Insights operator is disabled.
      duration: 300
      evaluationTime: 0.000333577
      health: ok
      labels:
        severity: info
      lastEvaluation: "2023-02-07T10:16:52.230589049Z"
      name: InsightsDisabled
      query: cluster_operator_conditions{condition="Disabled",name="insights"} ==
        1
      state: inactive
      type: alerting
    - alerts:
      - activeAt: "2023-01-12T10:19:22.229190558Z"
        annotations:
          description: Simple content access (SCA) is not enabled. Once enabled, Insights
            Operator can automatically import the SCA certificates from Red Hat OpenShift
            Cluster Manager making it easier to use the content provided by your Red
            Hat subscriptions when creating container images. See https://docs.openshift.com/container-platform/latest/cicd/builds/running-entitled-builds.html
            for more information.
          summary: Simple content access certificates are not available.
        labels:
          alertname: SimpleContentAccessNotAvailable
          condition: SCANotAvailable
          endpoint: metrics
          instance: 10.30.0.6:9099
          job: cluster-version-operator
          name: insights
          namespace: openshift-cluster-version
          pod: cluster-version-operator-688999f8cd-m7lb8
          reason: NotFound
          service: cluster-version-operator
          severity: info
        state: firing
        value: "1e+00"
      annotations:
        description: Simple content access (SCA) is not enabled. Once enabled, Insights
          Operator can automatically import the SCA certificates from Red Hat OpenShift
          Cluster Manager making it easier to use the content provided by your Red
          Hat subscriptions when creating container images. See https://docs.openshift.com/container-platform/latest/cicd/builds/running-entitled-builds.html
          for more information.
        summary: Simple content access certificates are not available.
      duration: 300
      evaluationTime: 0.000599797
      health: ok
      labels:
        severity: info
      lastEvaluation: "2023-02-07T10:16:52.230924882Z"
      name: SimpleContentAccessNotAvailable
      query: max_over_time(cluster_operator_conditions{condition="SCANotAvailable",name="insights",reason="NotFound"}[5m])
        == 1
      state: firing
      type: alerting
status: success
```

</details>

## `rules`
PrometheusRules present in the cluster; it's possible to filter them, as an example by the status, if we are interested only for the `firing` ones we can execute:
``` 
$ omc prometheus alertrules -s firing
RULE                                 STATE    AGE   ALERTS   ACTIVE SINCE
ClusterNotUpgradeable                firing   10s   1        10 Jan 23 10:20 UTC
UpdateAvailable                      firing   17s   1        22 Jan 23 04:59 UTC
SimpleContentAccessNotAvailable      firing   23s   1        12 Jan 23 10:19 UTC
APIRemovedInNextEUSReleaseInUse      firing   22s   2        05 Feb 23 17:50 UTC
Watchdog                             firing   1s    1        10 Jan 23 10:45 UTC
AlertmanagerReceiversNotConfigured   firing   15s   1        22 Sep 22 14:32 UTC
KubeJobCompletion                    firing   8s    1        18 Oct 22 19:37 UTC
KubeJobFailed                        firing   8s    3        22 Sep 22 14:33 UTC
CsvAbnormalFailedOver2Min            firing   21s   1        12 Jan 23 10:14 UTC
```

<details>
<summary> Get AlertingRules by group name</summary>

```
$ omc prometheus alertrule --group etcd
RULE                              STATE      AGE   ALERTS   ACTIVE SINCE
etcdMembersDown                   inactive   9s    0        ----
etcdNoLeader                      inactive   9s    0        ----
etcdGRPCRequestsSlow              inactive   9s    0        ----
etcdMemberCommunicationSlow       inactive   9s    0        ----
etcdHighNumberOfFailedProposals   inactive   9s    0        ----
etcdHighFsyncDurations            inactive   9s    0        ----
etcdHighFsyncDurations            inactive   9s    0        ----
etcdHighCommitDurations           inactive   9s    0        ----
etcdBackendQuotaLowSpace          inactive   9s    0        ----
etcdExcessiveDatabaseGrowth       inactive   9s    0        ----
```
</details>


