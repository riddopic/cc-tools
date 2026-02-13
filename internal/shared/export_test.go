package shared

// ExportFileExists exports fileExists for testing.
func ExportFileExists(path string, deps *Dependencies) bool {
	return fileExists(path, deps)
}

// ProjectHelperDeps returns the deps field from a ProjectHelper for testing.
func (p *ProjectHelper) ProjectHelperDeps() *Dependencies {
	return p.deps
}
