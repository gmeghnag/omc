package get

// specific labels https://github.com/seans3/kubernetes/blob/6108dac6708c026b172f3928e137c206437791da/pkg/printers/internalversion/printers_test.go#L1979
import (
	appsv1 "github.com/openshift/openshift-apiserver/pkg/apps/apis/apps"

	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"k8s.io/kubernetes/pkg/apis/admissionregistration"
	"k8s.io/kubernetes/pkg/apis/apiserverinternal"
	"k8s.io/kubernetes/pkg/apis/apps"
	autoscaling "k8s.io/kubernetes/pkg/apis/autoscaling"
	batch "k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/certificates"
	"k8s.io/kubernetes/pkg/apis/coordination"
	flowcontrol "k8s.io/kubernetes/pkg/apis/flowcontrol"
	"k8s.io/kubernetes/pkg/apis/networking"
	nodeapi "k8s.io/kubernetes/pkg/apis/node"
	"k8s.io/kubernetes/pkg/apis/policy"
	rbac "k8s.io/kubernetes/pkg/apis/rbac"
	"k8s.io/kubernetes/pkg/apis/resource"
	"k8s.io/kubernetes/pkg/apis/scheduling"

	discovery "k8s.io/kubernetes/pkg/apis/discovery"
	storage "k8s.io/kubernetes/pkg/apis/storage"

	templateapi "github.com/openshift/openshift-apiserver/pkg/template/apis/template"

	"github.com/openshift/openshift-apiserver/pkg/build/apis/build"
	"k8s.io/apimachinery/pkg/runtime/schema"

	authorizationv1 "github.com/openshift/api/authorization/v1"
	configv1 "github.com/openshift/api/config/v1"
	quotav1 "github.com/openshift/api/quota/v1"
	securityv1 "github.com/openshift/api/security/v1"
	imagev1 "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
	projectv1helpers "github.com/openshift/openshift-apiserver/pkg/project/apis/project"
	"github.com/openshift/openshift-apiserver/pkg/route/apis/route"
	runtime "k8s.io/apimachinery/pkg/runtime"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func addAdmissionRegistrationTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "admissionregistration.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&admissionregistration.MutatingWebhookConfiguration{},
		&admissionregistration.ValidatingAdmissionPolicyBinding{},
		&admissionregistration.ValidatingWebhookConfiguration{},
		&admissionregistration.ValidatingAdmissionPolicy{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addApiextensionsTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "apiextensions.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&apiextensionsv1.CustomResourceDefinition{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addApiextensionsV1Beta1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "apiextensions.k8s.io", Version: "v1beta1"}
	types := []runtime.Object{
		&apiextensionsv1beta1.CustomResourceDefinition{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addApiServerInternalTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "apiserverinternal.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&apiserverinternal.StorageVersion{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addAutoscalingV1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "autoscaling", Version: "v1"}
	types := []runtime.Object{
		&autoscaling.HorizontalPodAutoscaler{},
		&autoscaling.Scale{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addAutoscalingV2Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "autoscaling", Version: "v2"}
	types := []runtime.Object{
		&autoscaling.HorizontalPodAutoscaler{},
		&autoscaling.Scale{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addCertificatesTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "certificates.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&certificates.CertificateSigningRequest{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addCoordinationTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "coordination.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&coordination.Lease{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addAppsV1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "apps.openshift.io", Version: "v1"}
	types := []runtime.Object{
		&appsv1.DeploymentConfig{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addBuildTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "build.openshift.io", Version: "v1"}
	types := []runtime.Object{
		&build.Build{},
		&build.BuildConfig{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addRBACTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&rbac.ClusterRole{},
		&rbac.ClusterRoleBinding{},
		&rbac.Role{},
		&rbac.RoleBinding{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addAuthorizationTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "authorization.openshift.io", Version: "v1"}
	types := []runtime.Object{
		&authorizationv1.ClusterRole{},
		&authorizationv1.ClusterRoleBinding{},
		&authorizationv1.RoleBindingRestriction{},
		&authorizationv1.SubjectRulesReview{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

// OCP
func addImageTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "image.openshift.io", Version: "v1"}
	types := []runtime.Object{
		&imagev1.Image{},
		&imagev1.ImageTag{},
		&imagev1.ImageStream{},
		&imagev1.ImageStreamTag{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addBatchTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "batch", Version: "v1"}
	types := []runtime.Object{
		&batch.CronJob{},
		&batch.Job{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addFlowControlTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta1"}
	types := []runtime.Object{
		&flowcontrol.FlowSchema{},
		&flowcontrol.PriorityLevelConfiguration{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addFlowControlV1B2Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta2"}
	types := []runtime.Object{
		&flowcontrol.FlowSchema{},
		&flowcontrol.PriorityLevelConfiguration{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addDiscoveryTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "discovery.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&discovery.EndpointSlice{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addAppsTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "apps", Version: "v1"}
	types := []runtime.Object{
		&apps.ControllerRevision{},
		&apps.Deployment{},
		&apps.DaemonSet{},
		&apps.ReplicaSet{},
		&apps.StatefulSet{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addNetworkingTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "networking.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&networking.ClusterCIDR{},
		&networking.IngressClass{},
		&networking.Ingress{},
		&networking.NetworkPolicy{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addPolicyV1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "policy", Version: "v1"}
	types := []runtime.Object{
		&policy.PodDisruptionBudget{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addPolicyV1B1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "policy", Version: "v1"}
	types := []runtime.Object{
		&policy.PodSecurityPolicy{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addProjectV1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "project.openshift.io", Version: "v1"}
	types := []runtime.Object{
		&projectv1helpers.Project{},
		&projectv1helpers.ProjectRequest{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addQuotaV1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "quota.openshift.io", Version: "v1"}
	types := []runtime.Object{
		&quotav1.AppliedClusterResourceQuota{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}
func addRouteV1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "route.openshift.io", Version: "v1"}
	types := []runtime.Object{
		&route.Route{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addSecurityV1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "security.openshift.io", Version: "v1"}
	types := []runtime.Object{
		&securityv1.PodSecurityPolicyReview{},
		&securityv1.PodSecurityPolicySelfSubjectReview{},
		&securityv1.PodSecurityPolicySubjectReview{},
		&securityv1.SecurityContextConstraints{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addStorageV1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "storage.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&storage.StorageClass{},
		&storage.CSINode{},
		&storage.CSIDriver{},
		&storage.VolumeAttachment{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addStorageV1B1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "storage.k8s.io", Version: "v1beta1"}
	types := []runtime.Object{
		&storage.CSIStorageCapacity{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addResourceV1A2Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "resource", Version: "v1alpha2"}
	types := []runtime.Object{
		&resource.ResourceClass{},
		&resource.ResourceClaim{},
		&resource.ResourceClaimTemplate{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addSchedulingTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "scheduling.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&scheduling.PriorityClass{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}
func addTemplateV1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "template.openshift.io", Version: "v1"}
	types := []runtime.Object{
		&templateapi.Template{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addNodeTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "node.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&nodeapi.RuntimeClass{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addApiRegistrationTypes(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "apiregistration.k8s.io", Version: "v1"}
	types := []runtime.Object{
		&apiregistrationv1.APIService{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}

func addConfigV1Types(scheme *runtime.Scheme) error {
	GroupVersion := schema.GroupVersion{Group: "config.openshift.io", Version: "v1"}
	types := []runtime.Object{
		&configv1.ClusterVersion{},
	}
	scheme.AddKnownTypes(GroupVersion, types...)
	return nil
}
