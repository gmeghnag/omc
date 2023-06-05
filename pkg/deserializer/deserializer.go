package deserializer

// specific labels https://github.com/seans3/kubernetes/blob/6108dac6708c026b172f3928e137c206437791da/pkg/printers/internalversion/printers_test.go#L1979
import (

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiregistration "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	nodeapi "k8s.io/kubernetes/pkg/apis/node"

	appsv1 "github.com/openshift/openshift-apiserver/pkg/apps/apis/apps"
	imagev1 "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
	"github.com/openshift/openshift-apiserver/pkg/project/apis/project"
	corev1 "k8s.io/api/core/v1"

	authorizationv1 "github.com/openshift/api/authorization/v1"
	//"k8s.io/kubernetes/pkg/apis/rbac"
	//rbac "k8s.io/api/rbac/v1"

	"k8s.io/kubernetes/pkg/apis/coordination"
	"k8s.io/kubernetes/pkg/apis/networking"
	"k8s.io/kubernetes/pkg/apis/rbac"

	template "github.com/openshift/openshift-apiserver/pkg/template/apis/template"
	storage "k8s.io/kubernetes/pkg/apis/storage"

	// "k8s.io/client-go/kubernetes/scheme"
	//"k8s.io/apimachinery/pkg/api/meta"

	//runtime "k8s.io/apimachinery/pkg/runtime"
	//utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"github.com/openshift/openshift-apiserver/pkg/build/apis/build"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	//core "k8s.io/kubernetes/pkg/apis/core"
	//ocpinternal "github.com/openshift/openshift-apiserver/pkg/apps/printers/internalversion"

	"k8s.io/kubernetes/pkg/apis/core"

	// cliprint "k8s.io/cli-runtime/pkg/printers"
	quotav1 "github.com/openshift/api/quota/v1"
	securityv1 "github.com/openshift/api/security/v1"
	"github.com/openshift/openshift-apiserver/pkg/route/apis/route"
	runtime "k8s.io/apimachinery/pkg/runtime"

	//
	admissionregistration "k8s.io/kubernetes/pkg/apis/admissionregistration"
	apiserverinternal "k8s.io/kubernetes/pkg/apis/apiserverinternal"
	apps "k8s.io/kubernetes/pkg/apis/apps"
	autoscaling "k8s.io/kubernetes/pkg/apis/autoscaling"
	batch "k8s.io/kubernetes/pkg/apis/batch"
	certificates "k8s.io/kubernetes/pkg/apis/certificates"
	discovery "k8s.io/kubernetes/pkg/apis/discovery"
	flowcontrol "k8s.io/kubernetes/pkg/apis/flowcontrol"
	policy "k8s.io/kubernetes/pkg/apis/policy"
	resource "k8s.io/kubernetes/pkg/apis/resource"
	scheduling "k8s.io/kubernetes/pkg/apis/scheduling"
)

func RawObjectToRuntimeObject(rawObject []byte, schema *runtime.Scheme) runtime.Object {
	codec := serializer.NewCodecFactory(schema)
	decode := codec.UniversalDeserializer()
	obj, _, err := decode.Decode([]byte(rawObject), nil, nil)
	if err != nil {
		//fmt.Println("loglevel2", err)
	}
	switch obj.(type) {
	case *admissionregistration.MutatingWebhookConfiguration:
		return &admissionregistration.MutatingWebhookConfiguration{}
	case *admissionregistration.ValidatingAdmissionPolicyBinding:
		return &admissionregistration.ValidatingAdmissionPolicyBinding{}
	case *admissionregistration.ValidatingWebhookConfiguration:
		return &admissionregistration.ValidatingWebhookConfiguration{}
	case *admissionregistration.ValidatingAdmissionPolicy:
		return &admissionregistration.ValidatingAdmissionPolicy{}
	case *apiregistration.APIService:
		return &apiregistration.APIService{}
	case *apiserverinternal.StorageVersion:
		return &apiserverinternal.StorageVersion{}
	case *apps.StatefulSet:
		return &apps.StatefulSet{}
	case *apps.ReplicaSet:
		return &apps.ReplicaSet{}
	case *apps.Deployment:
		return &apps.Deployment{}
	case *apps.DaemonSet:
		return &apps.DaemonSet{}
	case *apps.ControllerRevision:
		return &apps.ControllerRevision{}
	case *appsv1.DeploymentConfig:
		return &appsv1.DeploymentConfig{}
	case *authorizationv1.ClusterRole:
		return &rbac.ClusterRole{}
	case *authorizationv1.ClusterRoleBinding:
		return &rbac.ClusterRoleBinding{}
	case *authorizationv1.RoleBindingRestriction:
		return &authorizationv1.RoleBindingRestriction{}
	case *authorizationv1.SubjectRulesReview:
		return &authorizationv1.SubjectRulesReview{}
	case *autoscaling.HorizontalPodAutoscaler:
		return &autoscaling.HorizontalPodAutoscaler{}
	case *autoscaling.Scale:
		return &autoscaling.Scale{}
	case *batch.CronJob:
		return &batch.CronJob{}
	case *batch.Job:
		return &batch.Job{}
	case *build.Build:
		return &build.Build{}
	case *build.BuildConfig:
		return &build.BuildConfig{}
	case *certificates.CertificateSigningRequest:
		return &certificates.CertificateSigningRequest{}
	case *coordination.Lease:
		return &coordination.Lease{}
	case *corev1.Pod:
		return &core.Pod{}
	case *corev1.PodTemplate:
		return &core.PodTemplate{}
	case *corev1.ReplicationController:
		return &core.ReplicationController{}
	case *corev1.Service:
		return &core.Service{}
	case *corev1.Endpoints:
		return &core.Endpoints{}
	case *corev1.Namespace:
		return &core.Namespace{}
	case *corev1.Secret:
		return &core.Secret{}
	case *corev1.ServiceAccount:
		return &core.ServiceAccount{}
	case *corev1.Node:
		return &core.Node{}
	case *corev1.PersistentVolume:
		return &core.PersistentVolume{}
	case *corev1.PersistentVolumeClaim:
		return &core.PersistentVolumeClaim{}
	case *corev1.Event:
		return &core.Event{}
	case *corev1.ComponentStatus:
		return &core.ComponentStatus{}
	case *corev1.ConfigMap:
		return &core.ConfigMap{}
	case *corev1.ResourceQuota:
		return &core.ResourceQuota{}
	case *discovery.EndpointSlice:
		return &discovery.EndpointSlice{}
	case *flowcontrol.FlowSchema:
		return &flowcontrol.FlowSchema{}
	case *flowcontrol.PriorityLevelConfiguration:
		return &flowcontrol.PriorityLevelConfiguration{}
	case *imagev1.Image:
		return &imagev1.Image{}
	case *imagev1.ImageStream:
		return &imagev1.ImageStream{}
	case *imagev1.ImageStreamTag:
		return &imagev1.ImageStreamTag{}
	case *imagev1.ImageTag:
		return &imagev1.ImageTag{}
	case *networking.ClusterCIDR:
		return &networking.ClusterCIDR{}
	case *networking.IngressClass:
		return &networking.IngressClass{}
	case *networking.Ingress:
		return &networking.Ingress{}
	case *networking.NetworkPolicy:
		return &networking.NetworkPolicy{}
	case *nodeapi.RuntimeClass:
		return &nodeapi.RuntimeClass{}
	case *policy.PodDisruptionBudget:
		return &policy.PodDisruptionBudget{}
	case *policy.PodSecurityPolicy:
		return &policy.PodSecurityPolicy{}
	case *project.Project:
		return &project.Project{}
	case *project.ProjectRequest:
		return &project.ProjectRequest{}
	case *quotav1.AppliedClusterResourceQuota:
		return &quotav1.AppliedClusterResourceQuota{}
	case *resource.ResourceClass:
		return &resource.ResourceClass{}
	case *resource.ResourceClaim:
		return &resource.ResourceClaim{}
	case *resource.ResourceClaimTemplate:
		return &resource.ResourceClaimTemplate{}
	case *rbac.ClusterRole:
		return &rbac.ClusterRole{}
	case *rbac.ClusterRoleBinding:
		return &rbac.ClusterRoleBinding{}
	case *rbac.Role:
		return &rbac.Role{}
	case *rbac.RoleBinding:
		return &rbac.RoleBinding{}
	case *route.Route:
		return &route.Route{}
	case *scheduling.PriorityClass:
		return &scheduling.PriorityClass{}
	case *securityv1.PodSecurityPolicyReview:
		return &securityv1.PodSecurityPolicyReview{}
	case *securityv1.PodSecurityPolicySelfSubjectReview:
		return &securityv1.PodSecurityPolicySelfSubjectReview{}
	case *securityv1.PodSecurityPolicySubjectReview:
		return &securityv1.PodSecurityPolicySubjectReview{}
	case *storage.CSIStorageCapacity:
		return &storage.CSIStorageCapacity{}
	case *storage.StorageClass:
		return &storage.StorageClass{}
	case *storage.CSINode:
		return &storage.CSINode{}
	case *storage.CSIDriver:
		return &storage.CSIDriver{}
	case *storage.VolumeAttachment:
		return &storage.VolumeAttachment{}
	case *template.Template:
		return &template.Template{}
	}
	//fmt.Println("RUNTIME UNKNOW")
	return &runtime.Unknown{}
}
