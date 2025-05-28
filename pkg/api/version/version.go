package version

import (
	"fmt"
	"net/http"
	"strings"
)

// APIVersion represents a specific API version
type APIVersion struct {
	Major      int
	Minor      int
	Deprecated bool
	SunsetDate string
}

// VersionManager handles API versioning
type VersionManager struct {
	versions map[string]APIVersion
	default_ string
}

// NewVersionManager creates a new version manager
func NewVersionManager(defaultVersion string) *VersionManager {
	return &VersionManager{
		versions: make(map[string]APIVersion),
		default_: defaultVersion,
	}
}

// RegisterVersion registers a new API version
func (vm *VersionManager) RegisterVersion(version string, major, minor int, deprecated bool, sunsetDate string) {
	vm.versions[version] = APIVersion{
		Major:      major,
		Minor:      minor,
		Deprecated: deprecated,
		SunsetDate: sunsetDate,
	}
}

// GetVersion extracts API version from request
func (vm *VersionManager) GetVersion(r *http.Request) string {
	// Check Accept header first
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/vnd.task.") {
		parts := strings.Split(accept, "application/vnd.task.")
		if len(parts) > 1 {
			version := strings.Split(parts[1], "+")[0]
			if _, ok := vm.versions[version]; ok {
				return version
			}
		}
	}

	// Check URL path
	parts := strings.Split(r.URL.Path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, "v") {
			version := strings.TrimPrefix(part, "v")
			if _, ok := vm.versions[version]; ok {
				// Remove version from path
				r.URL.Path = fmt.Sprintf("/%s", strings.Join(append(parts[:i], parts[i+1:]...), "/"))
				return version
			}
		}
	}

	return vm.default_
}

// VersionMiddleware handles API versioning
func (vm *VersionManager) VersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := vm.GetVersion(r)
		apiVersion := vm.versions[version]

		// Set version headers
		w.Header().Set("X-API-Version", version)
		
		// Handle deprecated versions
		if apiVersion.Deprecated {
			w.Header().Set("Warning", fmt.Sprintf("299 - \"Deprecated API version %s. Please upgrade before %s\"", version, apiVersion.SunsetDate))
		}

		// Store version in context
		ctx := r.Context()
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
} 