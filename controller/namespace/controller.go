package namespace

import (
	"log/slog"

	"github.com/temporalio/kube-temporal/pkg/controller/base"
	"github.com/temporalio/kube-temporal/pkg/controller/config"
	"github.com/temporalio/kube-temporal/pkg/types"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cloudv1alpha1 "github.com/temporalio/resource-operator/api/cloud/v1alpha1"
)

// +kubebuilder:rbac:groups=cloud.temporal.io,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.temporal.io,resources=namespaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.temporal.io,resources=namespaces/finalizers,verbs=update

const (
	controllerName = "namespace"
)

type NewOption func(*controller)

// New returns a new [types.Controller] that handles [cloudv1alpha1.Namespace]
// resources.
func New(
	cfg *config.Config,
	logger *slog.Logger,
) types.Controller {
	return &controller{
		Controller: base.New(cfg, logger),
	}
}

type controller struct {
	base.Controller
}

// BindManager binds the Controller to the supplied [ctrlrt.Manager].
func (c *controller) BindManager(m ctrlrt.Manager) error {
	c.Controller.Client = m.GetClient()
	c.Controller.Reader = m.GetAPIReader()
	builder := ctrlrt.NewControllerManagedBy(m).
		Named(controllerName).
		For(&cloudv1alpha1.Namespace{}).
		WithEventFilter(predicate.GenerationChangedPredicate{})
	return builder.Complete(c)
}
