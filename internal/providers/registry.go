package providers

// All returns all compile-time registered providers.
// When new providers are implemented, they should be added to the
// returned slice.
//
// Note: provider registration is done at the cmd layer to avoid import cycles.
// See cmd/grafanactl/root/command.go for the full provider list.
func All() []Provider {
	return []Provider{}
}
