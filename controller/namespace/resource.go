package namespace

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudv1alpha1 "github.com/temporalio/kube-temporal/api/cloud/v1alpha1"
	"github.com/temporalio/kube-temporal/pkg/kube/transform"
	"github.com/temporalio/kube-temporal/pkg/types"
)

// resourceFromRequest returns a Resource representing the requested Kubernetes
// namespaced object.
func (c *controller) resourceFromRequest(
	ctx context.Context,
	req ctrlrt.Request,
) (types.Resource, error) {
	cr := &cloudv1alpha1.Namespace{}
	// We use APIReader in order to prevent stale reads.
	if err := c.apiReader.Get(ctx, req.NamespacedName, cr); err != nil {
		return nil, err
	}
	return &resource{cr}, nil
}

// resource wraps a Namespace CR and implements the types.Resource
// interface.
type resource struct {
	// cr is the Kubernetes-native custom resource.
	cr *cloudv1alpha1.Namespace
}

// IsBeingDeleted returns true if the Kubernetes resource has a non-zero
// deletion timestamp
func (r *resource) IsBeingDeleted() bool {
	return !r.cr.DeletionTimestamp.IsZero()
}

// ClientObject returns the Kubernetes controller-runtime Client representation
// of the Resource
func (r *resource) ClientObject() client.Object {
	return r.cr
}

// RuntimeObject returns the Kubernetes apimachinery/runtime representation of
// the Resource
func (r *resource) RuntimeObject() runtime.Object {
	return r.cr
}

// GroupVersionKind returns the GroupVersionKind that the Resource
// represents.
func (r *resource) GroupVersionKind() schema.GroupVersionKind {
	return cloudv1alpha1.SchemeGroupVersion.WithKind("Namespace")
}

// Conditions returns the Resource's collection of Conditions from its status
// field. Satisfies the `pkg/condition.Manager` interface.
func (r *resource) Conditions() []*metav1.Condition {
	return r.cr.Status.Conditions
}

// ReplaceConditions replaces the Resource's collection of Conditions with the
// supplied collection. Satisfies the `pkg/condition.Manager` interface.
func (r *resource) ReplaceConditions(conditions []*metav1.Condition) {
	r.cr.Status.Conditions = conditions
}

// DeepCopy returns a copy of the resource
func (r *resource) DeepCopy() types.Resource {
	crCopy := r.cr.DeepCopy()
	return &resource{crCopy}
}

// Unstructured returns the resource as a [unstructured.Unstructured]
func (r *resource) Unstructured() (*unstructured.Unstructured, error) {
	return transform.ToUnstructured(r.RuntimeObject())
}

var _ types.Resource = (*resource)(nil)
