// Package v1 contains API Schema definitions for the logging v1 API group
// +kubebuilder:object:generate=true
// +groupName=logging.openshift.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "logging.openshift.io", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// +kubebuilder:rbac:groups=console.openshift.io,resources=consolelinks;consoleexternalloglinks,verbs=get;create;update;delete
// +kubebuilder:rbac:groups=logging.openshift.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=core,resources=pods;pods/exec;services;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts;services/finalizers,verbs=*
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes;routes/custom-host,verbs="*"
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=*
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=*
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules;servicemonitors,verbs=*
// +kubebuilder:rbac:groups=oauth.openshift.io,resources=oauthclients,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=*
// +kubebuilder:rbac:urls=/metrics,verbs=get
// +kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews;subjectaccessreviews,verbs=create
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=*
// +kubebuilder:rbac:groups=config.openshift.io,resources=proxies,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=create;delete
// +kubebuilder:rbac:groups=apps,resourceNames=elasticsearch-operator,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;create;update
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=get;list;watch;create;update
