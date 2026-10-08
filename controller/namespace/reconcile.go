package namespace

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	"github.com/temporalio/kube-temporal/pkg/condition"
	ctrlerrors "github.com/temporalio/kube-temporal/pkg/controller/errors"
	ctrlreqlog "github.com/temporalio/kube-temporal/pkg/controller/request/log"
	"github.com/temporalio/kube-temporal/pkg/controller/requeue"
	"github.com/temporalio/kube-temporal/pkg/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconcile performs a full reconciliation for the object referred to by the
// Request.
//
// If the returned error is non-nil, the Result is ignored and the request will
// be requeued using exponential backoff. The only exception is if the error is
// a TerminalError in which case no requeuing happens.
//
// If the error is nil and the returned Result has a non-zero
// result.RequeueAfter, the request will be requeued after the specified
// duration.
//
// If the error is nil and result.RequeueAfter is zero and result.Requeue is
// true, the request will be requeued using exponential backoff.
func (c *controller) Reconcile(
	ctx context.Context,
	req ctrlrt.Request,
) (ctrlrt.Result, error) {
	desired, err := c.resourceFromRequest(ctx, req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// resource wasn't found. just ignore these.
			return ctrlrt.Result{}, nil
		}
		return ctrlrt.Result{}, err
	}

	ctx = ctrlreqlog.ToContext(ctx, c.Logger(), desired)

	latest, err := c.reconcile(ctx, desired)
	return c.handleReconcileError(ctx, desired, latest, err)
}

// reconcile attempts to make the latest observed state of the resource match
// the latest desired state.
//
// It returns a copy of the resource that represents the latest observed state.
func (c *controller) reconcile(
	ctx context.Context,
	desired types.Resource,
) (types.Resource, error) {
	if desired.IsBeingDeleted() {
		return c.resourceDelete(ctx, desired)
	}
	latest, err := c.resourceSync(ctx, desired)
	if err != nil {
		return latest, err
	}
	return latest, nil
}

// resourceSync ensures that the supplied Resource's backing Temporal Cloud
// Namespace matches the supplied desired state.
//
// It returns a copy of the resource that represents the latest observed state.
func (c *controller) resourceSync(
	ctx context.Context,
	desired types.Resource,
) (types.Resource, error) {
	var err error
	log := ctrlreqlog.FromContext(ctx)
	exit := log.Trace("resource.sync")
	defer func() {
		exit(err)
	}()

	var latest types.Resource // the newly created or mutated resource

	// We clear the resource's Conditions collection at the start of each
	// reconcile loop to ensure that the Status.Conditions always contains the
	// freshest indication of current state.
	condition.Clear(desired)
	defer func() {
		c.conditionsSync(ctx, latest, err)
	}()
	return latest, nil
}

// resourceDelete ensures that the supplied Resource's backing Temporal Cluster
// is destroyed along with all child dependent resources.
//
// Returns a copy of the resource with the latest state either right before
// deletion OR after a failed attempted deletion.
func (c *controller) resourceDelete(
	ctx context.Context,
	current types.Resource,
) (types.Resource, error) {
	var err error
	log := ctrlreqlog.FromContext(ctx)
	exit := log.Trace("resource.delete")
	defer func() {
		exit(err)
	}()
	return nil, nil
}

// conditionsSync examines the supplied resource's collection of Condition
// objects and ensures that appropriate Ready and Terminal Conditions are
// present.
func (c *controller) conditionsSync(
	ctx context.Context,
	res types.Resource,
	reconcileErr error,
) {
	if lo.IsNil(res) {
		return
	}

	var err error
	log := ctrlreqlog.FromContext(ctx)
	exit := log.Trace("conditions.sync")
	defer func() {
		exit(err)
	}()

	if errors.Is(reconcileErr, &ctrlerrors.TerminalError{}) {
		// If we got a terminal error while reconciling the resource, we set
		// its Stalled condition to True to indicate that another reconcile
		// will not change the status of the resource without a change to its
		// desired state.
		condition.SetStalled(
			res,
			metav1.ConditionFalse,
			"TerminalError",
			fmt.Sprintf(
				"reconciliation encountered a terminal error: %s",
				reconcileErr.Error(),
			),
		)
	} else {
		condition.SetStalled(
			res,
			metav1.ConditionFalse,
			"ReconcileSucceeded",
			"reconcile succeeded",
		)
	}
}

// handleReconcileError will handle errors from reconcile handlers, which
// respects runtime errors.
//
// If the `latest` parameter is not nil, this function will ALWAYS patch the
// latest Status fields back to the Kubernetes API.
func (c *controller) handleReconcileError(
	ctx context.Context,
	desired types.Resource,
	latest types.Resource,
	err error,
) (ctrlrt.Result, error) {
	if lo.IsNotNil(latest) {
		// The reconciliation loop may have returned an error, but if latest is
		// not nil, there may be some changes available in the CR's Status
		// struct (example: Conditions), and we want to make sure we save those
		// changes before proceeding.
		//
		// It is okay to patch status when resource is not present due to deletion
		// because a NotFound error is thrown which will be ignored.
		_ = c.patchResourceStatus(ctx, desired, latest)
	}
	if err == nil || errors.Is(err, &ctrlerrors.TerminalError{}) {
		// We don't requeue on a terminal error...
		return ctrlrt.Result{}, nil
	}

	var requeueNeededAfter *requeue.RequeueNeededAfter
	if errors.As(err, &requeueNeededAfter) {
		after := requeueNeededAfter.Duration()
		return ctrlrt.Result{RequeueAfter: after}, nil
	}

	var requeueNeeded *requeue.RequeueNeeded
	if errors.As(err, &requeueNeeded) {
		return ctrlrt.Result{Requeue: true}, nil
	}

	return ctrlrt.Result{}, err
}

// patchResourceStatus patches the custom resource in the Kubernetes API to
// match the supplied latest resource.
func (c *controller) patchResourceStatus(
	ctx context.Context,
	desired types.Resource,
	latest types.Resource,
) error {
	dobj := desired.DeepCopy().ClientObject()
	lobj := latest.DeepCopy().ClientObject()
	patch := client.MergeFrom(dobj)

	return c.Client.Status().Patch(ctx, lobj, patch)
}
