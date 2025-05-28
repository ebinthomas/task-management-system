package auth

import (
	"context"
)

// GetUserFromContext extracts the user from the context
func GetUserFromContext(ctx context.Context) (User, error) {
	claims, ok := ctx.Value("claims").(*Claims)
	if !ok {
		return User{}, ErrUserNotFound
	}

	return User{
		ID:    claims.UserID,
		Roles: claims.Roles,
	}, nil
}

// CanAccessResource checks if a user can access a specific resource
func CanAccessResource(ctx context.Context, resourceType, resourceID string) error {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return err
	}

	// Admins can access all resources
	if HasRole(user, "admin") {
		return nil
	}

	// Check resource ownership
	if !hasResourceAccess(user.ID, resourceType, resourceID) {
		return ErrResourceNotOwned
	}

	return nil
}

// Example service-level authorization check
type TaskService struct {
	// ... other fields ...
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID string, update interface{}) error {
	// Check if user can access this task
	if err := CanAccessResource(ctx, "task", taskID); err != nil {
		return err
	}

	// Proceed with update
	return nil
} 