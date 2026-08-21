package constants

// UserRole enumerates platform roles.
const (
	RoleUser  = "member"
	RoleAdmin = "Administrator"
)

// ValidRoles returns all accepted roles.
func ValidRoles() []string {
	return []string{RoleUser}
}
