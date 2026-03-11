package checks

import (
	"context"
	"fmt"
	"strconv"

	"github.com/grafana/grafanactl/internal/providers/synth/probes"
	"github.com/grafana/grafanactl/internal/providers/synth/smcfg"
	"github.com/grafana/grafanactl/internal/resources"
	"github.com/grafana/grafanactl/internal/resources/adapter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// staticDescriptor is the resource descriptor for SM Check resources.
//
//nolint:gochecknoglobals // Static descriptor used in init() self-registration pattern.
var staticDescriptor = resources.Descriptor{
	GroupVersion: schema.GroupVersion{
		Group:   "syntheticmonitoring.ext.grafana.app",
		Version: "v1alpha1",
	},
	Kind:     "Check",
	Singular: "check",
	Plural:   "checks",
}

// staticAliases are the short aliases for Check resources.
//
//nolint:gochecknoglobals // Static descriptor used in init() self-registration pattern.
var staticAliases = []string{"checks"}

// NewAdapterFactory returns a lazy adapter.Factory for SM checks.
// The factory captures the smcfg.Loader and constructs clients on first invocation.
func NewAdapterFactory(loader smcfg.Loader) adapter.Factory {
	return func(ctx context.Context) (adapter.ResourceAdapter, error) {
		baseURL, token, namespace, err := loader.LoadSMConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load SM config for checks adapter: %w", err)
		}

		checksClient := NewClient(baseURL, token)
		probesClient := probes.NewClient(baseURL, token)

		return &ResourceAdapter{
			client:       checksClient,
			probesClient: probesClient,
			namespace:    namespace,
		}, nil
	}
}

// ResourceAdapter bridges the checks.Client to the grafanactl resources pipeline.
type ResourceAdapter struct {
	client       *Client
	probesClient *probes.Client
	namespace    string
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

// StaticDescriptor returns the static descriptor for SM Check resources.
// Used for registration without constructing an adapter instance.
func StaticDescriptor() resources.Descriptor {
	return staticDescriptor
}

// StaticAliases returns the static aliases for SM Check resources.
func StaticAliases() []string {
	return staticAliases
}

// StaticGVK returns the static GroupVersionKind for SM Check resources.
func StaticGVK() schema.GroupVersionKind {
	return staticDescriptor.GroupVersionKind()
}

// probeNameMap fetches all probes and builds an ID→name map.
func (a *ResourceAdapter) probeNameMap(ctx context.Context) (map[int64]string, error) {
	probeList, err := a.probesClient.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching probe list for name resolution: %w", err)
	}

	nameMap := make(map[int64]string, len(probeList))
	for _, p := range probeList {
		nameMap[p.ID] = p.Name
	}

	return nameMap, nil
}

// probeIDMap fetches all probes and builds a name→ID map.
func (a *ResourceAdapter) probeIDMap(ctx context.Context) (map[string]int64, error) {
	probeList, err := a.probesClient.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching probe list for ID resolution: %w", err)
	}

	idMap := make(map[string]int64, len(probeList))
	for _, p := range probeList {
		idMap[p.Name] = p.ID
	}

	return idMap, nil
}

// List returns all check resources as unstructured objects.
func (a *ResourceAdapter) List(ctx context.Context, _ metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	checkList, err := a.client.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list checks: %w", err)
	}

	nameMap, err := a.probeNameMap(ctx)
	if err != nil {
		return nil, err
	}

	result := &unstructured.UnstructuredList{}
	for _, check := range checkList {
		res, err := ToResource(check, a.namespace, nameMap)
		if err != nil {
			return nil, fmt.Errorf("failed to convert check %d to resource: %w", check.ID, err)
		}

		result.Items = append(result.Items, res.Object)
	}

	return result, nil
}

// Get returns a single check resource by name (numeric string ID).
func (a *ResourceAdapter) Get(ctx context.Context, name string, _ metav1.GetOptions) (*unstructured.Unstructured, error) {
	id, err := strconv.ParseInt(name, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("check name must be a numeric ID, got %q: %w", name, err)
	}

	check, err := a.client.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get check %d: %w", id, err)
	}

	nameMap, err := a.probeNameMap(ctx)
	if err != nil {
		return nil, err
	}

	res, err := ToResource(*check, a.namespace, nameMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert check %d to resource: %w", id, err)
	}

	obj := res.ToUnstructured()
	return &obj, nil
}

// Create creates a new check resource from an unstructured object.
func (a *ResourceAdapter) Create(ctx context.Context, obj *unstructured.Unstructured, _ metav1.CreateOptions) (*unstructured.Unstructured, error) {
	res, err := resources.FromUnstructured(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert unstructured to resource: %w", err)
	}

	spec, _, err := FromResource(res)
	if err != nil {
		return nil, fmt.Errorf("failed to extract check spec from resource: %w", err)
	}

	tenant, err := a.client.GetTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SM tenant: %w", err)
	}

	idMap, err := a.probeIDMap(ctx)
	if err != nil {
		return nil, err
	}

	probeIDs, err := resolveProbeIDs(spec.Probes, idMap)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve probe IDs for create: %w", err)
	}
	check := SpecToCheck(spec, 0, tenant.ID, probeIDs)

	created, err := a.client.Create(ctx, check)
	if err != nil {
		return nil, fmt.Errorf("failed to create check: %w", err)
	}

	nameMap := invertIDMap(idMap)
	createdRes, err := ToResource(*created, a.namespace, nameMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert created check to resource: %w", err)
	}

	createdObj := createdRes.ToUnstructured()
	return &createdObj, nil
}

// Update updates an existing check resource from an unstructured object.
func (a *ResourceAdapter) Update(ctx context.Context, obj *unstructured.Unstructured, _ metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	res, err := resources.FromUnstructured(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert unstructured to resource: %w", err)
	}

	spec, id, err := FromResource(res)
	if err != nil {
		return nil, fmt.Errorf("failed to extract check spec from resource: %w", err)
	}

	tenant, err := a.client.GetTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SM tenant: %w", err)
	}

	idMap, err := a.probeIDMap(ctx)
	if err != nil {
		return nil, err
	}

	probeIDs, err := resolveProbeIDs(spec.Probes, idMap)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve probe IDs for update: %w", err)
	}
	check := SpecToCheck(spec, id, tenant.ID, probeIDs)

	updated, err := a.client.Update(ctx, check)
	if err != nil {
		return nil, fmt.Errorf("failed to update check %d: %w", id, err)
	}

	nameMap := invertIDMap(idMap)
	updatedRes, err := ToResource(*updated, a.namespace, nameMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert updated check to resource: %w", err)
	}

	updatedObj := updatedRes.ToUnstructured()
	return &updatedObj, nil
}

// Delete removes a check resource by name (numeric string ID).
func (a *ResourceAdapter) Delete(ctx context.Context, name string, _ metav1.DeleteOptions) error {
	id, err := strconv.ParseInt(name, 10, 64)
	if err != nil {
		return fmt.Errorf("check name must be a numeric ID, got %q: %w", name, err)
	}

	if err := a.client.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete check %d: %w", id, err)
	}

	return nil
}

// resolveProbeIDs converts probe names to IDs using the provided map.
// Returns an error if any name is not found in the map.
func resolveProbeIDs(names []string, idMap map[string]int64) ([]int64, error) {
	ids := make([]int64, 0, len(names))
	for _, name := range names {
		id, ok := idMap[name]
		if !ok {
			return nil, fmt.Errorf("probe %q not found", name)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// invertIDMap converts a name→ID map to an ID→name map.
func invertIDMap(idMap map[string]int64) map[int64]string {
	nameMap := make(map[int64]string, len(idMap))
	for name, id := range idMap {
		nameMap[id] = name
	}
	return nameMap
}
