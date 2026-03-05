package providers

// registry holds all providers registered via Register().
var registry []Provider //nolint:gochecknoglobals // Self-registration pattern requires package-level state.

// Register adds a provider to the global registry.
// Providers call this from their init() function.
func Register(p Provider) {
	registry = append(registry, p)
}

// All returns all registered providers.
// Returns a non-nil empty slice when no providers have been registered.
func All() []Provider {
	if registry == nil {
		return []Provider{}
	}
	return registry
}
