package shared

// Dependencies holds all external dependencies for the shared package.
type Dependencies struct {
	FS SharedFS
}

// NewDefaultDependencies creates production dependencies.
func NewDefaultDependencies() *Dependencies {
	return &Dependencies{
		FS: &RealFS{},
	}
}
