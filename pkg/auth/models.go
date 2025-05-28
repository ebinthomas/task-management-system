package auth

// User represents the authenticated user
type User struct {
	ID    string
	Roles []string
}

// HasRole checks if the user has any of the specified roles
func HasRole(user User, roles ...string) bool {
	for _, requiredRole := range roles {
		for _, userRole := range user.Roles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
} 