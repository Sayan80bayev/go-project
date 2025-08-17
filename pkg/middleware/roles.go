package middleware

type Role string

const (
	RoleAdmin     Role = "ADMIN"
	RoleModerator Role = "MODERATOR"
	RoleUser      Role = "USER"
)

var RoleHierarchy = map[Role]int{
	RoleAdmin:     3,
	RoleModerator: 2,
	RoleUser:      1,
}

// CanModerate проверяет, может ли текущая роль изменить другую роль
func (r Role) CanModerate(target Role) bool {
	return RoleHierarchy[r] > RoleHierarchy[target]
}
