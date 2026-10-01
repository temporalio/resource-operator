package namespace

import (
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	cloudv1alpha1 "github.com/temporalio/kube-temporal/api/cloud/v1alpha1"
	"github.com/temporalio/kube-temporal/pkg/controller/config"
	"github.com/temporalio/kube-temporal/pkg/types"
)

// +kubebuilder:rbac:groups=cloud.temporal.io,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.temporal.io,resources=namespaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.temporal.io,resources=namespaces/finalizers,verbs=update

const (
	controllerName = "namespace"
)

// New returns a new [types.Controller] that handles [cloudv1alpha1.Namespace]
// resources.
func New(
	cfg *config.Config,
	opts ...types.ControllerWithOption,
) *controller {
	c := &controller{
		cfg: cfg,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithLogger sets the Controller's Logger to the supplied logger.
func WithLogger(l logr.Logger) types.ControllerWithOption {
	return func(c types.Controller) {
		c.SetLogger(l)
	}
}

type controller struct {
	kc        client.Client
	apiReader client.Reader
	cfg       *config.Config
	log       logr.Logger
}

// Logger returns the Controller's logger.
func (c *controller) Logger() logr.Logger {
	return c.log
}

// SetLogger sets the Controller's logger.
func (c *controller) SetLogger(l logr.Logger) {
	c.log = l
}

// BindManager binds the Controller to the supplied [ctrlrt.Manager].
func (c *controller) BindManager(m ctrlrt.Manager) error {
	c.kc = m.GetClient()
	c.apiReader = m.GetAPIReader()
	builder := ctrlrt.NewControllerManagedBy(m).
		Named(controllerName).
		For(&cloudv1alpha1.Namespace{}).
		WithEventFilter(predicate.GenerationChangedPredicate{})
	return builder.Complete(c)
}

var _ types.Controller = (*controller)(nil)
