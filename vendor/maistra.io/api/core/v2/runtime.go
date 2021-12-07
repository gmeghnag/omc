package v2

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	v1 "maistra.io/api/core/v1"
)

// ControlPlaneRuntimeConfig configures execution parameters for control plane
// componets.
type ControlPlaneRuntimeConfig struct {
	// Components allows specifying execution parameters for specific control plane
	// componets.  The key of the map is the component name to which the settings
	// should be applied.
	// +optional
	Components map[ControlPlaneComponentName]*ComponentRuntimeConfig `json:"components,omitempty"`
	// Defaults will be merged into specific component config.
	// .Values.global.defaultResources, e.g.
	// +optional
	Defaults *DefaultRuntimeConfig `json:"defaults,omitempty"`
}

// ControlPlaneComponentName simple type for control plane component names
type ControlPlaneComponentName string

const (
	// ControlPlaneComponentNameSecurity - security (citadel)
	ControlPlaneComponentNameSecurity ControlPlaneComponentName = "security"
	// ControlPlaneComponentNameGalley - galley
	ControlPlaneComponentNameGalley ControlPlaneComponentName = "galley"
	// ControlPlaneComponentNamePilot - pilot
	ControlPlaneComponentNamePilot ControlPlaneComponentName = "pilot"
	// ControlPlaneComponentNameMixer - mixer
	ControlPlaneComponentNameMixer ControlPlaneComponentName = "mixer"
	// ControlPlaneComponentNameMixerPolicy - mixer.policy
	ControlPlaneComponentNameMixerPolicy ControlPlaneComponentName = "mixer.policy"
	// ControlPlaneComponentNameMixerTelemetry - mixer.telemetry
	ControlPlaneComponentNameMixerTelemetry ControlPlaneComponentName = "mixer.telemetry"
	// ControlPlaneComponentNameGlobalOauthProxy - global.oauthproxy
	ControlPlaneComponentNameGlobalOauthProxy ControlPlaneComponentName = "global.oauthproxy"
	// ControlPlaneComponentNameSidecarInjectoryWebhook - sidecarInjectorWebhook
	ControlPlaneComponentNameSidecarInjectoryWebhook ControlPlaneComponentName = "sidecarInjectorWebhook"
	// ControlPlaneComponentNameTracing - tracing
	ControlPlaneComponentNameTracing ControlPlaneComponentName = "tracing"
	// ControlPlaneComponentNameTracingJaeger - tracing.jaeger
	ControlPlaneComponentNameTracingJaeger ControlPlaneComponentName = "tracing.jaeger"
	// ControlPlaneComponentNameTracingJaegerElasticsearch - tracing.jaeger.elasticsearch
	ControlPlaneComponentNameTracingJaegerElasticsearch ControlPlaneComponentName = "tracing.jaeger.elasticsearch"
	// ControlPlaneComponentNameTracingJaegerAgent - tracing.jaeger.agent
	ControlPlaneComponentNameTracingJaegerAgent ControlPlaneComponentName = "tracing.jaeger.agent"
	// ControlPlaneComponentNameTracingJaegerAllInOne - tracing.jaeger.allInOne
	ControlPlaneComponentNameTracingJaegerAllInOne ControlPlaneComponentName = "tracing.jaeger.allInOne"
	// ControlPlaneComponentNameTracingJaegerCollector - tracing.jaeger.collector
	ControlPlaneComponentNameTracingJaegerCollector ControlPlaneComponentName = "tracing.jaeger.collector"
	// ControlPlaneComponentNameTracingJaegerQuery - tracing.jaeger.query
	ControlPlaneComponentNameTracingJaegerQuery ControlPlaneComponentName = "tracing.jaeger.query"
	// ControlPlaneComponentNamePrometheus - prometheus
	ControlPlaneComponentNamePrometheus ControlPlaneComponentName = "prometheus"
	// ControlPlaneComponentNameKiali - kiali
	ControlPlaneComponentNameKiali ControlPlaneComponentName = "kiali"
	// ControlPlaneComponentNameGrafana - grafana
	ControlPlaneComponentNameGrafana ControlPlaneComponentName = "grafana"
	// ControlPlaneComponentNameThreeScale - 3scale
	ControlPlaneComponentNameThreeScale ControlPlaneComponentName = "3scale"
	// ControlPlaneComponentNameWASMCacher - wasm-extensions cacher
	ControlPlaneComponentNameWASMCacher ControlPlaneComponentName = "wasmExtensions.cacher"
)

// ControlPlaneComponentNames - supported runtime components
var ControlPlaneComponentNames = []ControlPlaneComponentName{
	ControlPlaneComponentNameSecurity,
	ControlPlaneComponentNameGalley,
	ControlPlaneComponentNamePilot,
	ControlPlaneComponentNameMixer,
	ControlPlaneComponentNameMixerPolicy,
	ControlPlaneComponentNameMixerTelemetry,
	ControlPlaneComponentNameGlobalOauthProxy,
	ControlPlaneComponentNameSidecarInjectoryWebhook,
	ControlPlaneComponentNameTracing,
	ControlPlaneComponentNameTracingJaeger,
	ControlPlaneComponentNameTracingJaegerElasticsearch,
	ControlPlaneComponentNameTracingJaegerAgent,
	ControlPlaneComponentNameTracingJaegerAllInOne,
	ControlPlaneComponentNameTracingJaegerCollector,
	ControlPlaneComponentNameTracingJaegerQuery,
	ControlPlaneComponentNamePrometheus,
	ControlPlaneComponentNameKiali,
	ControlPlaneComponentNameGrafana,
	ControlPlaneComponentNameThreeScale,
	ControlPlaneComponentNameWASMCacher,
}

// ComponentRuntimeConfig allows for partial customization of a component's
// runtime configuration (Deployment, PodTemplate, auto scaling, pod disruption, etc.)
type ComponentRuntimeConfig struct {
	// Deployment specific overrides
	// +optional
	Deployment *DeploymentRuntimeConfig `json:"deployment,omitempty"`

	// Pod specific overrides
	// +optional
	Pod *PodRuntimeConfig `json:"pod,omitempty"`

	// .Values.*.resource, imagePullPolicy, etc.
	// +optional
	Container *ContainerConfig `json:"container,omitempty"`
}

// DeploymentRuntimeConfig allow customization of a component's Deployment
// resource, including additional labels/annotations, replica count, autoscaling,
// rollout strategy, etc.
type DeploymentRuntimeConfig struct {
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	// .Values.*.replicaCount
	Replicas *int32 `json:"replicas,omitempty"`

	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	// +patchStrategy=retainKeys
	// .Values.*.rollingMaxSurge, rollingMaxUnavailable, etc.
	Strategy *appsv1.DeploymentStrategy `json:"strategy,omitempty" patchStrategy:"retainKeys"`

	// Autoscaling specifies the configuration for a HorizontalPodAutoscaler
	// to be applied to this deployment.  Null indicates no auto scaling.
	// .Values.*.autoscale* fields
	// +optional
	AutoScaling *AutoScalerConfig `json:"autoScaling,omitempty"`
}

// CommonDeploymentRuntimeConfig represents deployment settings common to both
// default and component specific settings
type CommonDeploymentRuntimeConfig struct {
	// .Values.global.podDisruptionBudget.enabled, if not null
	// XXX: this is currently a global setting, not per component.  perhaps
	// this should only be available on the defaults?
	// +optional
	PodDisruption *PodDisruptionBudget `json:"podDisruption,omitempty"`
}

// AutoScalerConfig is used to configure autoscaling for a deployment
type AutoScalerConfig struct {
	Enablement `json:",inline"`
	// lower limit for the number of pods that can be set by the autoscaler, default 1.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.
	// +optional
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
	// target average CPU utilization (represented as a percentage of requested CPU) over all the pods;
	// if not specified the default autoscaling policy will be used.
	// +optional
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage,omitempty"`
}

// PodRuntimeConfig is used to customize pod configuration for a component
type PodRuntimeConfig struct {
	CommonPodRuntimeConfig `json:",inline"`

	// Metadata allows additional annotations/labels to be applied to the pod
	// .Values.*.podAnnotations
	// XXX: currently, additional lables are not supported
	// +optional
	Metadata *MetadataConfig `json:"metadata,omitempty"`

	// If specified, the pod's scheduling constraints
	// +optional
	// .Values.podAntiAffinityLabelSelector, podAntiAffinityTermLabelSelector, nodeSelector
	// NodeAffinity is not supported at this time
	// PodAffinity is not supported at this time
	Affinity *Affinity `json:"affinity,omitempty"`
}

// CommonPodRuntimeConfig represents pod settings common to both defaults and
// component specific configuration
type CommonPodRuntimeConfig struct {
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	// .Values.nodeSelector
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	// .Values.tolerations
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// .Values.global.priorityClassName
	// XXX: currently, this is only a global setting.  maybe only allow setting in global runtime defaults?
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`
}

// Affinity is the structure used by Istio for specifying Pod affinity
// XXX: istio does not support full corev1.Affinity settings, hence the special
// types here.
type Affinity struct {
	// XXX: use corev1.PodAntiAffinity instead, the only things not supported are namespaces and weighting
	// +optional
	PodAntiAffinity PodAntiAffinity `json:"podAntiAffinity,omitempty"`
}

// PodAntiAffinity configures anti affinity for pod scheduling
type PodAntiAffinity struct {
	// +optional
	RequiredDuringScheduling []PodAntiAffinityTerm `json:"requiredDuringScheduling,omitempty"`
	// +optional
	PreferredDuringScheduling []PodAntiAffinityTerm `json:"preferredDuringScheduling,omitempty"`
}

// PodAntiAffinityTerm is a simplified version of corev1.PodAntiAffinityTerm
type PodAntiAffinityTerm struct {
	metav1.LabelSelectorRequirement `json:",inline"`
	// This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching
	// the labelSelector in the specified namespaces, where co-located is defined as running on a node
	// whose value of the label with key topologyKey matches that of any node on which any of the
	// selected pods is running.
	// Empty topologyKey is not allowed.
	// +optional
	TopologyKey string `json:"topologyKey,omitempty"`
}

// ContainerConfig to be applied to containers in a pod, in a deployment
type ContainerConfig struct {
	CommonContainerConfig `json:",inline"`
	// +optional
	Image string `json:"imageName,omitempty"`
	// +optional
	Env map[string]string `json:"env,omitempty"`
}

// CommonContainerConfig represents container settings common to both defaults
// and component specific configuration.
type CommonContainerConfig struct {
	// +optional
	ImageRegistry string `json:"imageRegistry,omitempty"`
	// +optional
	ImageTag string `json:"imageTag,omitempty"`
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// PodDisruptionBudget details
// XXX: currently only configurable globally (i.e. no component values.yaml equivalent)
type PodDisruptionBudget struct {
	Enablement `json:",inline"`
	// +optional
	MinAvailable *intstr.IntOrString `json:"minAvailable,omitempty"`
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
}

// DefaultRuntimeConfig specifies default execution parameters to apply to
// control plane deployments/pods when no specific component overrides have been
// specified.  These settings will be merged with component specific settings.
type DefaultRuntimeConfig struct {
	// Deployment defaults
	// +optional
	Deployment *CommonDeploymentRuntimeConfig `json:"deployment,omitempty"`
	// Pod defaults
	// +optional
	Pod *CommonPodRuntimeConfig `json:"pod,omitempty"`
	// Container overrides to be merged with component specific overrides.
	// +optional
	Container *CommonContainerConfig `json:"container,omitempty"`
}

// MetadataConfig represents additional metadata to be applied to resources
type MetadataConfig struct {
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ComponentServiceConfig is used to customize the service associated with a component.
type ComponentServiceConfig struct {
	// Metadata represents addtional annotations/labels to be applied to the
	// component's service.
	// +optional
	Metadata *MetadataConfig `json:"metadata,omitempty"`
	// NodePort specifies a NodePort for the component's Service.
	// .Values.<component>.service.nodePort.port, ...enabled is true if not null
	// +optional
	NodePort *int32 `json:"nodePort,omitempty"`
	// Ingress specifies details for accessing the component's service through
	// a k8s Ingress or OpenShift Route.
	// +optional
	Ingress *ComponentIngressConfig `json:"ingress,omitempty"`
}

// ComponentIngressConfig is used to customize a k8s Ingress or OpenShift Route
// for the service associated with a component.
type ComponentIngressConfig struct {
	Enablement `json:",inline"`
	// Metadata represents additional metadata to be applied to the ingress/route.
	// +optional
	Metadata *MetadataConfig `json:"metadata,omitempty"`
	// Hosts represents a list of host names to configure.  Note, OpenShift route
	// only supports a single host name per route.  An empty host name implies
	// a default host name for the Route.
	// XXX: is a host name required for k8s Ingress?
	// +optional
	Hosts []string `json:"hosts,omitempty"`
	// ContextPath represents the context path to the service.
	// +optional
	ContextPath string `json:"contextPath,omitempty"`
	// TLS is used to configure TLS for the Ingress/Route
	// XXX: should this be something like RawExtension, as the configuration differs between Route and Ingress?
	// +optional
	TLS *v1.HelmValues `json:"tls,omitempty"`
}

// ComponentPersistenceConfig is used to configure persistance for a component.
type ComponentPersistenceConfig struct {
	Enablement `json:",inline"`
	// StorageClassName for the PersistentVolumeClaim
	// +optional
	StorageClassName string `json:"storageClassName,omitempty"`
	// AccessMode for the PersistentVolumeClaim
	// +optional
	AccessMode corev1.PersistentVolumeAccessMode `json:"accessMode,omitempty"`
	// Resources to request for the PersistentVolumeClaim
	// +optional
	Resources *corev1.ResourceRequirements `json:"capacity,omitempty"`
}
