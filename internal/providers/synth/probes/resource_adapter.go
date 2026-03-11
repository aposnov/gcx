package probes

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/grafana/grafanactl/internal/providers/synth/smcfg"
	"github.com/grafana/grafanactl/internal/resources"
	"github.com/grafana/grafanactl/internal/resources/adapter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// staticDescriptor is the resource descriptor for SM Probe resources.
//
//nolint:gochecknoglobals // Static descriptor used in init() self-registration pattern.
var staticDescriptor = resources.Descriptor{
	GroupVersion: schema.GroupVersion{
		Group:   "syntheticmonitoring.ext.grafana.app",
		Version: "v1alpha1",
	},
	Kind:     "Probe",
	Singular: "probe",
	Plural:   "probes",
}

// staticAliases are the short aliases for Probe resources.
//
//nolint:gochecknoglobals // Static descriptor used in init() self-registration pattern.
var staticAliases = []string{"probes"}

// NewAdapterFactory returns a lazy adapter.Factory for SM probes.
// The factory captures the smcfg.Loader and constructs the client on first invocation.
func NewAdapterFactory(loader smcfg.Loader) adapter.Factory {
	return func(ctx context.Context) (adapter.ResourceAdapter, error) {
		baseURL, token, namespace, err := loader.LoadSMConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load SM config for probes adapter: %w", err)
		}

		return &ResourceAdapter{
			client:    NewClient(baseURL, token),
			namespace: namespace,
		}, nil
	}
}

// ResourceAdapter bridges the probes.Client to the grafanactl resources pipeline.
// Probes are read-only in the SM API; Create, Update, and Delete return errors.
type ResourceAdapter struct {
	client    *Client
	namespace string
}

var _ adapter.ResourceAdapter = &ResourceAdapter{}

// Descriptor returns the resource descriptor this adapter serves.
func (a *ResourceAdapter) Descriptor() resources.Descriptor {
	return staticDescriptor
}

// Aliases returns short names for selector resolution.
func (a *ResourceAdapter) Aliases() []string {
	return staticAliases
}

// StaticDescriptor returns the static descriptor for SM Probe resources.
// Used for registration without constructing an adapter instance.
func StaticDescriptor() resources.Descriptor {
	return staticDescriptor
}

// StaticAliases returns the static aliases for SM Probe resources.
func StaticAliases() []string {
	return staticAliases
}

// StaticGVK returns the static GroupVersionKind for SM Probe resources.
func StaticGVK() schema.GroupVersionKind {
	return staticDescriptor.GroupVersionKind()
}

// List returns all probe resources as unstructured objects.
func (a *ResourceAdapter) List(ctx context.Context, _ metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	probeList, err := a.client.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list probes: %w", err)
	}

	result := &unstructured.UnstructuredList{}
	for _, probe := range probeList {
		res, err := ToResource(probe, a.namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to convert probe %d to resource: %w", probe.ID, err)
		}

		result.Items = append(result.Items, res.Object)
	}

	return result, nil
}

// Get returns a single probe resource by name (numeric string ID).
// Since the probes API has no single-probe GET endpoint, this fetches the full
// list and filters by ID.
func (a *ResourceAdapter) Get(ctx context.Context, name string, _ metav1.GetOptions) (*unstructured.Unstructured, error) {
	id, err := strconv.ParseInt(name, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("probe name must be a numeric ID, got %q: %w", name, err)
	}

	probeList, err := a.client.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list probes: %w", err)
	}

	for _, probe := range probeList {
		if probe.ID == id {
			res, err := ToResource(probe, a.namespace)
			if err != nil {
				return nil, fmt.Errorf("failed to convert probe %d to resource: %w", id, err)
			}

			obj := res.ToUnstructured()
			return &obj, nil
		}
	}

	return nil, fmt.Errorf("probe %d not found", id)
}

// Create is not supported for probes — they are managed by Grafana and read-only.
func (a *ResourceAdapter) Create(_ context.Context, _ *unstructured.Unstructured, _ metav1.CreateOptions) (*unstructured.Unstructured, error) {
	return nil, errors.ErrUnsupported
}

// Update is not supported for probes — they are managed by Grafana and read-only.
func (a *ResourceAdapter) Update(_ context.Context, _ *unstructured.Unstructured, _ metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, errors.ErrUnsupported
}

// Delete is not supported for probes — they are managed by Grafana and read-only.
func (a *ResourceAdapter) Delete(_ context.Context, _ string, _ metav1.DeleteOptions) error {
	return errors.ErrUnsupported
}
