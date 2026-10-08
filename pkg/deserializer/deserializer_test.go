package deserializer

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/kubernetes/pkg/apis/core"
)

func newTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return s
}

// TestRuntimeObjectForGVK_MatchesRawAndCaches checks the memoized resolver
// returns the same concrete internal type as the direct resolver, serves a
// cache hit on the second call, and keeps distinct GVKs isolated.
func TestRuntimeObjectForGVK_MatchesRawAndCaches(t *testing.T) {
	ResetRuntimeObjectTypeCache()
	t.Cleanup(ResetRuntimeObjectTypeCache)

	s := newTestScheme(t)
	podGVK := schema.GroupVersionKind{Version: "v1", Kind: "Pod"}
	rawPod := []byte(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":"p"}}`)

	direct := RawObjectToRuntimeObject(rawPod, s)
	if _, ok := direct.(*core.Pod); !ok {
		t.Fatalf("RawObjectToRuntimeObject: expected *core.Pod, got %T", direct)
	}

	first := RuntimeObjectForGVK(podGVK, rawPod, s)
	if reflect.TypeOf(first) != reflect.TypeOf(direct) {
		t.Fatalf("RuntimeObjectForGVK returned %T, want %T", first, direct)
	}
	if _, ok := gvkTypeCache.Load(podGVK); !ok {
		t.Fatalf("expected Pod GVK to be cached after first resolution")
	}

	// Second call must be served from cache and stay the same type, while
	// returning a distinct (freshly allocated) instance.
	second := RuntimeObjectForGVK(podGVK, rawPod, s)
	if reflect.TypeOf(second) != reflect.TypeOf(first) {
		t.Fatalf("cached call returned %T, want %T", second, first)
	}
	if second == first {
		t.Fatalf("expected a fresh object instance per call, got the same pointer")
	}

	svcGVK := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
	rawSvc := []byte(`{"apiVersion":"v1","kind":"Service","metadata":{"name":"s"}}`)
	svc := RuntimeObjectForGVK(svcGVK, rawSvc, s)
	if _, ok := svc.(*core.Service); !ok {
		t.Fatalf("expected *core.Service for Service GVK, got %T (GVK cache cross-contamination?)", svc)
	}
}

// TestRuntimeObjectForGVK_DoesNotCacheUnknown ensures an unrecognised object
// resolves to runtime.Unknown and is NOT cached, so a later recognisable object
// of the same GVK can still resolve to its real type.
func TestRuntimeObjectForGVK_DoesNotCacheUnknown(t *testing.T) {
	ResetRuntimeObjectTypeCache()
	t.Cleanup(ResetRuntimeObjectTypeCache)

	s := newTestScheme(t)
	gvk := schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Gizmo"}
	rawUnknown := []byte(`{"apiVersion":"example.com/v1","kind":"Gizmo","metadata":{"name":"g"}}`)

	obj := RuntimeObjectForGVK(gvk, rawUnknown, s)
	if _, ok := obj.(*runtime.Unknown); !ok {
		t.Fatalf("expected *runtime.Unknown for an unregistered type, got %T", obj)
	}
	if _, ok := gvkTypeCache.Load(gvk); ok {
		t.Fatalf("runtime.Unknown must not be cached by GVK")
	}
}
