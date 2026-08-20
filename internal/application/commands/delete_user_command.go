package commands

import "uuid"

// DeleteUser is the command to delete a user's own account.
type DeleteUser struct {
	UserID uuid.UUID
}

// DeleteUserResult reports whether the deletion succeeded.
type DeleteUserResult struct {
	Success bool
}
