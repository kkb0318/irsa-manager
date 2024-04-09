package kubernetes

// Owner contains options for setting the field manager.
type Owner struct {
	// Field sets the field manager name for the given server-side apply patch.
	Field string
}
