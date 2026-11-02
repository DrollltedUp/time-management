package auth

type Role string

const (
	RoleEmployee Role = "employee"
	RoleManager  Role = "manager"
)

func (r Role) Valid() bool {
	switch r {
	case RoleEmployee, RoleManager:
		return true
	}

	return false
}
