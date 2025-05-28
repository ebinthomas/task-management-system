package version

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionManager_RegisterVersion(t *testing.T) {
	vm := NewVersionManager("1.0")

	// Test registering a version
	vm.RegisterVersion("1.0", 1, 0, false, "")
	version, exists := vm.versions["1.0"]
	assert.True(t, exists)
	assert.Equal(t, 1, version.Major)
	assert.Equal(t, 0, version.Minor)
	assert.False(t, version.Deprecated)
	assert.Empty(t, version.SunsetDate)

	// Test registering a deprecated version
	vm.RegisterVersion("0.9", 0, 9, true, "2024-12-31")
	version, exists = vm.versions["0.9"]
	assert.True(t, exists)
	assert.True(t, version.Deprecated)
	assert.Equal(t, "2024-12-31", version.SunsetDate)
}

func TestVersionManager_GetVersion(t *testing.T) {
	vm := NewVersionManager("1.0")
	vm.RegisterVersion("1.0", 1, 0, false, "")
	vm.RegisterVersion("1.1", 1, 1, false, "")
	vm.RegisterVersion("0.9", 0, 9, true, "2024-12-31")

	tests := []struct {
		name           string
		acceptHeader   string
		path          string
		expectedPath  string
		expectedVersion string
	}{
		{
			name:            "Accept Header Version",
			acceptHeader:    "application/vnd.task.v1.1+json",
			path:           "/api/tasks",
			expectedPath:   "/api/tasks",
			expectedVersion: "1.1",
		},
		{
			name:            "URL Path Version",
			acceptHeader:    "",
			path:           "/v1.0/api/tasks",
			expectedPath:   "/api/tasks",
			expectedVersion: "1.0",
		},
		{
			name:            "Default Version",
			acceptHeader:    "",
			path:           "/api/tasks",
			expectedPath:   "/api/tasks",
			expectedVersion: "1.0",
		},
		{
			name:            "Invalid Version",
			acceptHeader:    "application/vnd.task.v2.0+json",
			path:           "/api/tasks",
			expectedPath:   "/api/tasks",
			expectedVersion: "1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}

			version := vm.GetVersion(req)
			assert.Equal(t, tt.expectedVersion, version)
			assert.Equal(t, tt.expectedPath, req.URL.Path)
		})
	}
}

func TestVersionMiddleware(t *testing.T) {
	vm := NewVersionManager("1.0")
	vm.RegisterVersion("1.0", 1, 0, false, "")
	vm.RegisterVersion("0.9", 0, 9, true, "2024-12-31")

	tests := []struct {
		name           string
		version        string
		deprecated     bool
		expectedHeader string
		expectWarning  bool
	}{
		{
			name:           "Current Version",
			version:        "1.0",
			deprecated:     false,
			expectedHeader: "1.0",
			expectWarning:  false,
		},
		{
			name:           "Deprecated Version",
			version:        "0.9",
			deprecated:     true,
			expectedHeader: "0.9",
			expectWarning:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := vm.VersionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
			if tt.version != "" {
				req.Header.Set("Accept", "application/vnd.task."+tt.version+"+json")
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, tt.expectedHeader, rr.Header().Get("X-API-Version"))

			if tt.expectWarning {
				assert.Contains(t, rr.Header().Get("Warning"), "Deprecated API version")
			} else {
				assert.Empty(t, rr.Header().Get("Warning"))
			}
		})
	}
} 