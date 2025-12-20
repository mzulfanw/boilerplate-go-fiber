package user

import "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/response"

type CreateUserRequest struct {
	Email    string   `json:"email" validate:"required,notblank"`
	Password string   `json:"password" validate:"required,notblank"`
	IsActive *bool    `json:"is_active,omitempty"`
	RoleIDs  []string `json:"role_ids,omitempty"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email" validate:"required_without_all=Password IsActive,notblank"`
	Password *string `json:"password" validate:"required_without_all=Email IsActive,notblank"`
	IsActive *bool   `json:"is_active" validate:"required_without_all=Email Password"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserListResponse struct {
	Items []UserResponse    `json:"items"`
	Meta  response.PageMeta `json:"meta"`
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
