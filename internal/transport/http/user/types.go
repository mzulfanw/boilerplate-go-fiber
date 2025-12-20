package user

type CreateUserRequest struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	IsActive *bool    `json:"is_active,omitempty"`
	RoleIDs  []string `json:"role_ids,omitempty"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
	IsActive *bool   `json:"is_active"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserRolesRequest struct {
	RoleIDs []string `json:"role_ids"`
}

type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type UserRolesResponse struct {
	UserID string         `json:"user_id"`
	Roles  []RoleResponse `json:"roles"`
}
