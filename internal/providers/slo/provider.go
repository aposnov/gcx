package slo

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/grafana/grafanactl/internal/config"
	"github.com/grafana/grafanactl/internal/providers"
	"github.com/grafana/grafanactl/internal/providers/slo/definitions"
	"github.com/grafana/grafanactl/internal/providers/slo/reports"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() { //nolint:gochecknoinits // Self-registration pattern (like database/sql drivers).
	providers.Register(&SLOProvider{})
}

// SLOProvider manages Grafana SLO resources.
type SLOProvider struct{}

// Name returns the unique identifier for this provider.
func (p *SLOProvider) Name() string { return "slo" }

// ShortDesc returns a one-line description of the provider.
func (p *SLOProvider) ShortDesc() string { return "Manage Grafana SLO resources." }

// Commands returns the Cobra commands contributed by this provider.
func (p *SLOProvider) Commands() []*cobra.Command {
	loader := &configLoader{}

	sloCmd := &cobra.Command{
		Use:   "slo",
		Short: p.ShortDesc(),
	}

	// Bind config flags on the parent — all subcommands inherit these.
	loader.bindFlags(sloCmd.PersistentFlags())

	sloCmd.AddCommand(definitions.Commands(loader))
	sloCmd.AddCommand(reports.Commands(loader))

	return []*cobra.Command{sloCmd}
}

// Validate checks that the given provider configuration is valid.
// The SLO provider uses Grafana's built-in authentication, so no extra keys
// are required.
func (p *SLOProvider) Validate(cfg map[string]string) error {
	return nil
}

// ConfigKeys returns the configuration keys used by this provider.
// The SLO provider uses Grafana's built-in authentication and does not require
// additional provider-specific keys.
func (p *SLOProvider) ConfigKeys() []providers.ConfigKey {
	return nil
}

// configLoader is a minimal config loading helper that avoids importing
// cmd/grafanactl/config (which would create an import cycle via internal/providers).
type configLoader struct {
	configFile string
	ctxName    string
}

func (l *configLoader) bindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&l.configFile, "config", "", "Path to the configuration file to use")
	flags.StringVar(&l.ctxName, "context", "", "Name of the context to use")
}

// LoadRESTConfig loads the REST config from the config file, applying
// env var overrides and context flags. It mirrors the logic in
// cmd/grafanactl/config.Options.LoadRESTConfig.
func (l *configLoader) LoadRESTConfig(ctx context.Context) (config.NamespacedRESTConfig, error) {
	source := l.configSource()

	overrides := []config.Override{
		// Apply env vars into the current context.
		func(cfg *config.Config) error {
			if cfg.CurrentContext == "" {
				cfg.CurrentContext = config.DefaultContextName
			}

			if !cfg.HasContext(cfg.CurrentContext) {
				cfg.SetContext(cfg.CurrentContext, true, config.Context{})
			}

			curCtx := cfg.Contexts[cfg.CurrentContext]
			if curCtx.Grafana == nil {
				curCtx.Grafana = &config.GrafanaConfig{}
			}

			if err := env.Parse(curCtx); err != nil {
				return err
			}

			// Resolve GRAFANA_PROVIDER_{NAME}_{KEY} environment variables.
			const providerEnvPrefix = "GRAFANA_PROVIDER_"
			for _, envVar := range os.Environ() {
				parts := strings.SplitN(envVar, "=", 2)
				if len(parts) != 2 {
					continue
				}

				key, val := parts[0], parts[1]
				if !strings.HasPrefix(key, providerEnvPrefix) {
					continue
				}

				suffix := key[len(providerEnvPrefix):]
				nameParts := strings.SplitN(suffix, "_", 2)
				if len(nameParts) != 2 || nameParts[0] == "" || nameParts[1] == "" {
					continue
				}

				providerName := strings.ToLower(nameParts[0])
				configKey := strings.ReplaceAll(strings.ToLower(nameParts[1]), "_", "-")

				if curCtx.Providers == nil {
					curCtx.Providers = make(map[string]map[string]string)
				}
				if curCtx.Providers[providerName] == nil {
					curCtx.Providers[providerName] = make(map[string]string)
				}
				curCtx.Providers[providerName][configKey] = val
			}

			return nil
		},
	}

	// Apply context flag override.
	if l.ctxName != "" {
		overrides = append(overrides, func(cfg *config.Config) error {
			if !cfg.HasContext(l.ctxName) {
				return config.ContextNotFound(l.ctxName)
			}
			cfg.CurrentContext = l.ctxName
			return nil
		})
	}

	// Validate after loading.
	overrides = append(overrides, func(cfg *config.Config) error {
		if !cfg.HasContext(cfg.CurrentContext) {
			return config.ContextNotFound(cfg.CurrentContext)
		}
		return cfg.GetCurrentContext().Validate()
	})

	loaded, err := config.Load(ctx, source, overrides...)
	if err != nil {
		return config.NamespacedRESTConfig{}, err
	}

	if !loaded.HasContext(loaded.CurrentContext) {
		return config.NamespacedRESTConfig{}, fmt.Errorf("context %q not found", loaded.CurrentContext)
	}

	return loaded.GetCurrentContext().ToRESTConfig(ctx), nil
}

func (l *configLoader) configSource() config.Source {
	if l.configFile != "" {
		return config.ExplicitConfigFile(l.configFile)
	}
	return config.StandardLocation()
}
