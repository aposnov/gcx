// Package smcfg defines the shared config loader interface for the synth provider.
package smcfg

import (
	"context"

	"github.com/grafana/grafanactl/internal/config"
)

// Loader can load SM credentials and the current namespace from config.
type Loader interface {
	LoadSMConfig(ctx context.Context) (baseURL, token, namespace string, err error)
}

// RESTConfigLoader can load a Grafana REST config for Prometheus queries.
type RESTConfigLoader interface {
	LoadRESTConfig(ctx context.Context) (config.NamespacedRESTConfig, error)
}

// ConfigLoader can load the full config for datasource discovery.
type ConfigLoader interface {
	LoadConfig(ctx context.Context) (*config.Config, error)
}

// DatasourceUIDSaver can persist a discovered Prometheus datasource UID to the SM provider config.
type DatasourceUIDSaver interface {
	SaveMetricsDatasourceUID(ctx context.Context, uid string) error
}

// StatusLoader combines SM config loading with Grafana REST config and full config loading.
// Used by status/timeline commands that need SM API + Prometheus + datasource discovery.
type StatusLoader interface {
	Loader
	RESTConfigLoader
	ConfigLoader
	DatasourceUIDSaver
}
