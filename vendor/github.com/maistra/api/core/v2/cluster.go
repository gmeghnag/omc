package v2

// ControlPlaneClusterConfig configures aspects related to clustering.
type ControlPlaneClusterConfig struct {
	// .Values.global.multiCluster.clusterName, defaults to Kubernetes
	// +optional
	Name string `json:"name,omitempty"`
	// .Values.global.network
	// XXX: not sure what the difference is between this and cluster name
	// +optional
	Network string `json:"network,omitempty"`
	// .Values.global.multiCluster.enabled, if not null
	// +optional
	MultiCluster *MultiClusterConfig `json:"multiCluster,omitempty"`
	// .Values.global.meshExpansion.enabled, if not null
	// XXX: it's not clear whether or not there is any overlap with MultiCluster,
	// i.e. does MultiCluster require mesh expansion ports to be configured on
	// the ingress gateway?
	// +optional
	MeshExpansion *MeshExpansionConfig `json:"meshExpansion,omitempty"`
}

// MultiClusterConfig configures aspects related to multi-cluster.
// implies the following:
// adds external to RequestedNetworkView (ISTIO_META_REQUESTED_NETWORK_VIEW) for egress gateway
// adds "global" and "{{ valueOrDefault .DeploymentMeta.Namespace \"default\" }}.global" to pod dns search suffixes
type MultiClusterConfig struct {
	Enablement `json:",inline"`
	// .Values.global.meshNetworks
	// XXX: if non-empty, local cluster network should be configured as:
	//  <spec.cluster.network>:
	//      endpoints:
	//      - fromRegistry: <spec.cluster.name>
	//      gateways:
	//      - service: <ingress-gateway-service-name>
	//        port: 443 # mtls port
	// +optional
	MeshNetworks map[string]MeshNetworkConfig `json:"meshNetworks,omitempty"`
}

// MeshExpansionConfig configures aspects related to mesh expansion
type MeshExpansionConfig struct {
	Enablement `json:",inline"`
	// .Values.global.meshExpansion.useILB, true if not null, otherwise uses ingress gateway
	// +optional
	ILBGateway *GatewayConfig `json:"ilbGateway,omitempty"`
}

// MeshNetworkConfig configures mesh networks for a multi-cluster mesh.
type MeshNetworkConfig struct {
	Endpoints []MeshEndpointConfig `json:"endpoints,omitempty"`
	Gateways  []MeshGatewayConfig  `json:"gateways,omitempty"`
}

// MeshEndpointConfig specifies the endpoint of a mesh network.  Only one of
// FromRegistry or FromCIDR may be specified
type MeshEndpointConfig struct {
	// +optional
	FromRegistry string `json:"fromRegistry,omitempty"`
	// +optional
	FromCIDR string `json:"fromCIDR,omitempty"`
}

// MeshGatewayConfig specifies the gateway which should be used for accessing
// the network
type MeshGatewayConfig struct {
	// +optional
	// +deprecated
	Service string `json:"service,omitempty"`
	// +optional
	RegistryServiceName string `json:"registryServiceName,omitempty"`
	// +optional
	Address string `json:"address,omitempty"`
	// +optional
	Port int32 `json:"port,omitempty"`
}
