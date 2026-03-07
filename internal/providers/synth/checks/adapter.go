package checks

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/grafana/grafanactl/internal/resources"
)

// ToResource converts an API Check + probe map to a K8s-envelope Resource.
// probeNames maps probe ID → name for display in the YAML file.
// Server-managed fields (id, tenantId, created, modified, channels) are stripped.
func ToResource(check Check, namespace string, probeNames map[int64]string) (*resources.Resource, error) {
	// Resolve probe IDs to names for the YAML spec.
	probeNameList := make([]string, 0, len(check.Probes))
	for _, id := range check.Probes {
		name, ok := probeNames[id]
		if !ok {
			name = strconv.FormatInt(id, 10) // fallback to numeric string if name unknown
		}
		probeNameList = append(probeNameList, name)
	}

	spec := CheckSpec{
		Job:              check.Job,
		Target:           check.Target,
		Frequency:        check.Frequency,
		Offset:           check.Offset,
		Timeout:          check.Timeout,
		Enabled:          check.Enabled,
		Labels:           check.Labels,
		Settings:         check.Settings,
		Probes:           probeNameList,
		BasicMetricsOnly: check.BasicMetricsOnly,
		AlertSensitivity: check.AlertSensitivity,
	}

	// Marshal spec to generic map for the K8s envelope.
	specData, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshalling check spec: %w", err)
	}

	var specMap map[string]any
	if err := json.Unmarshal(specData, &specMap); err != nil {
		return nil, fmt.Errorf("unmarshalling check spec to map: %w", err)
	}

	obj := map[string]any{
		"apiVersion": APIVersion,
		"kind":       Kind,
		"metadata": map[string]any{
			"name":      strconv.FormatInt(check.ID, 10),
			"namespace": namespace,
		},
		"spec": specMap,
	}

	return resources.MustFromObject(obj, resources.SourceInfo{}), nil
}

// FromResource converts a K8s-envelope Resource back to a CheckSpec.
// The numeric check ID is read from metadata.name (if set and parseable).
// Probe names in spec.probes are left as names — callers resolve them to IDs.
func FromResource(res *resources.Resource) (*CheckSpec, int64, error) {
	obj := res.Object.Object

	specRaw, ok := obj["spec"]
	if !ok {
		return nil, 0, errors.New("resource has no spec field")
	}

	specMap, ok := specRaw.(map[string]any)
	if !ok {
		return nil, 0, errors.New("resource spec is not a map")
	}

	specData, err := json.Marshal(specMap)
	if err != nil {
		return nil, 0, fmt.Errorf("marshalling spec: %w", err)
	}

	var spec CheckSpec
	if err := json.Unmarshal(specData, &spec); err != nil {
		return nil, 0, fmt.Errorf("unmarshalling spec to CheckSpec: %w", err)
	}

	// Parse the numeric ID from metadata.name (0 means "create new").
	var id int64
	name := res.Raw.GetName()
	if name != "" {
		if parsed, err := strconv.ParseInt(name, 10, 64); err == nil {
			id = parsed
		}
	}

	return &spec, id, nil
}

// SpecToCheck converts a CheckSpec + resolved probe IDs to an API Check struct.
// tenantID must be fetched from the server before calling this.
func SpecToCheck(spec *CheckSpec, id, tenantID int64, probeIDs []int64) Check {
	return Check{
		ID:               id,
		TenantID:         tenantID,
		Job:              spec.Job,
		Target:           spec.Target,
		Frequency:        spec.Frequency,
		Offset:           spec.Offset,
		Timeout:          spec.Timeout,
		Enabled:          spec.Enabled,
		Labels:           spec.Labels,
		Settings:         spec.Settings,
		Probes:           probeIDs,
		BasicMetricsOnly: spec.BasicMetricsOnly,
		AlertSensitivity: spec.AlertSensitivity,
	}
}

// FileNamer returns a function that produces the file path for a check resource.
// Path convention: checks/{id}.yaml.
func FileNamer(outputFormat string) func(*resources.Resource) string {
	return func(res *resources.Resource) string {
		return fmt.Sprintf("checks/%s.%s", res.Raw.GetName(), outputFormat)
	}
}
