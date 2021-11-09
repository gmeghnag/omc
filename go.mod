module omc

go 1.16

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/maistra/api v0.0.0-20210601141927-1cbee4cb8ce8
	github.com/olekukonko/tablewriter v0.0.5
	github.com/openshift/api v0.0.0-20210906075240-3611f00b94fd
	github.com/openshift/machine-api-operator v0.2.1-0.20210917195819-eb6706653664
	github.com/openshift/machine-config-operator v0.0.1-0.20210917223957-ff7e847c56ac
	github.com/operator-framework/api v0.10.7
	github.com/spf13/cast v1.4.0 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	go.etcd.io/etcd/api/v3 v3.5.0
	gopkg.in/yaml.v2 v2.4.0
	istio.io/client-go v0.0.0-20210503213042-e6eb157f0d81
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/kubectl v0.22.1
	maistra.io/api v0.0.0-20210823191737-259bd5854df6 // indirect
	sigs.k8s.io/yaml v1.2.0

)

replace (
	github.com/istio/api => istio.io/api v0.0.0-20191111210003-35e06ef8d838
	github.com/istio/client-go => istio.io/client-go v1.12.0-alpha.5.0.20211103172409-6d920928bee3 // indirect
	github.com/maistra/api => maistra.io/api v0.0.0-20210601141927-1cbee4cb8ce8
	sigs.k8s.io/cluster-api-provider-aws => github.com/openshift/cluster-api-provider-aws v0.2.1-0.20210121023454-5ffc5f422a80
	sigs.k8s.io/cluster-api-provider-azure => github.com/openshift/cluster-api-provider-azure v0.1.0-alpha.3.0.20210626224711-5d94c794092f
	sigs.k8s.io/cluster-api-provider-openstack => github.com/openshift/cluster-api-provider-openstack v0.0.0-20210302164104-8498241fa4bd
)
