package auth

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents our custom JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID string   `json:"uid"`
	Roles  []string `json:"roles"`
}

// Role represents a user role and its permissions
type Role struct {
	Name        string
	Permissions map[string][]string // endpoint -> allowed methods
}

// AuthConfig holds the middleware configuration
type AuthConfig struct {
	JWTSecret     []byte
	AllowedRoles  map[string]Role
	PublicPaths   []string // paths that don't require authentication
}

// matchPath checks if a request path matches a pattern
func matchPath(pattern, path string) bool {
	// Convert pattern to regex
	regexPattern := strings.ReplaceAll(pattern, "{id}", "[a-zA-Z0-9-]+")
	regexPattern = "^" + regexPattern + "$"
	
	match, err := regexp.MatchString(regexPattern, path)
	if err != nil {
		return false
	}
	return match
}

// AuthMiddleware handles JWT validation and role-based access control
func AuthMiddleware(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path is public
			for _, path := range config.PublicPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, ErrNoAuthHeader.Error(), http.StatusUnauthorized)
				return
			}

			// Check bearer format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, ErrInvalidAuthType.Error(), http.StatusUnauthorized)
				return
			}

			// Parse and validate token
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
				return config.JWTSecret, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
				return
			}

			// Check role permissions
			hasPermission := false
			for _, userRole := range claims.Roles {
				if role, exists := config.AllowedRoles[userRole]; exists {
					for pattern, methods := range role.Permissions {
						if matchPath(pattern, r.URL.Path) {
							for _, method := range methods {
								if method == r.Method {
									hasPermission = true
									break
								}
							}
							if hasPermission {
								break
							}
						}
					}
				}
				if hasPermission {
					break
				}
			}

			if !hasPermission {
				http.Error(w, ErrInsufficientRole.Error(), http.StatusForbidden)
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), "claims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ResourceOwnershipMiddleware checks if the user owns the resource or has admin rights
func ResourceOwnershipMiddleware(resourceType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for non-resource specific endpoints
			if r.Method == http.MethodGet && !strings.Contains(r.URL.Path, "/") {
				next.ServeHTTP(w, r)
				return
			}

			claims, ok := r.Context().Value("claims").(*Claims)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Admin role can access all resources
			for _, role := range claims.Roles {
				if role == "admin" {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check resource ownership
			resourceID := extractResourceID(r.URL.Path)
			if resourceID == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !hasResourceAccess(claims.UserID, resourceType, resourceID) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper function to extract resource ID from URL path
func extractResourceID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return ""
	}
	return parts[len(parts)-1]
}

// hasResourceAccess checks if a user has access to a specific resource
func hasResourceAccess(userID, resourceType, resourceID string) bool {
	// TODO: Implement resource ownership check based on your storage
	// This is where you would query your database to check ownership
	return true
}

// Example usage of role definitions
var DefaultRoles = map[string]Role{
	"admin": {
		Name: "admin",
		Permissions: map[string][]string{
			"/api/v1/tasks":          {"GET", "POST", "PUT", "DELETE"},
			"/api/v1/tasks/{id}":     {"GET", "PUT", "DELETE"},
			"/api/v1/users":          {"GET", "POST", "PUT", "DELETE"},
			"/api/v1/users/{id}":     {"GET", "PUT", "DELETE"},
			"/api/v1/metrics":        {"GET"},
			"/api/v1/settings":       {"GET", "PUT"},
		},
	},
	"user": {
		Name: "user",
		Permissions: map[string][]string{
			"/api/v1/tasks":          {"GET", "POST"},
			"/api/v1/tasks/{id}":     {"GET", "PUT", "DELETE"},
			"/api/v1/users/me":       {"GET", "PUT"},
		},
	},
	"viewer": {
		Name: "viewer",
		Permissions: map[string][]string{
			"/api/v1/tasks":          {"GET"},
			"/api/v1/tasks/{id}":     {"GET"},
		},
	},
}

// RequireRoles middleware checks if the user has any of the required roles
func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromContext(r.Context())
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !HasRole(user, roles...) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}