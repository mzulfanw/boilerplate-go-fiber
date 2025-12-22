package rbac

import "github.com/mzulfanw/boilerplate-go-fiber/internal/transport/http/response"

type RoleRequest struct {
	Name        string `json:"name" validate:"required,notblank"`
	Description string `json:"description"`
}

type PermissionRequest struct {
	Name        string `json:"name" validate:"required,notblank"`
	Description string `json:"description"`
}

type RolePermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids"`
}

type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type PermissionResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type RolePermissionsResponse struct {
	RoleID      string               `json:"role_id"`
	Permissions []PermissionResponse `json:"permissions"`
}

type RoleListResponse struct {
	Items []RoleResponse `json:"items"`
	Meta  response.PageMeta `json:"meta"`
}

type PermissionListResponse struct {
	Items []PermissionResponse `json:"items"`
	Meta  response.PageMeta    `json:"meta"`
}