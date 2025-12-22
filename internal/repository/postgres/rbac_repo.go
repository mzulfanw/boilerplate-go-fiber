package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	rbacdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/rbac"
)

type RBACRepository struct {
	pool *pgxpool.Pool
}

func NewRBACRepository(pool *pgxpool.Pool) *RBACRepository {
	return &RBACRepository{pool: pool}
}

func (r *RBACRepository) ListRoles(ctx context.Context, filter rbacdomain.ListFilterRole) (rbacdomain.ListRole, error) {
	where, args := buildRolesListFilter(filter)

	countQuery := `SELECT COUNT(*) FROM roles`
	if where != "" {
		countQuery += " " + where
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return rbacdomain.ListRole{}, err
	}

	limit := filter.Pagination.Limit()
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Pagination.Offset()
	if offset < 0 {
		offset = 0
	}

	listQuery := fmt.Sprintf(`
		SELECT id::text, name, COALESCE(description, ''), created_at
		FROM roles
		%s
		ORDER BY name
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	listArgs := append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return rbacdomain.ListRole{}, err
	}
	defer rows.Close()

	var roles []rbacdomain.Role
	for rows.Next() {
		var role rbacdomain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return rbacdomain.ListRole{}, err
		}
		roles = append(roles, role)
	}
	return rbacdomain.ListRole{
		Role:  roles,
		Total: total,
	}, nil
}

func buildRolesListFilter(filter rbacdomain.ListFilterRole) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if search := strings.TrimSpace(filter.Search); search != "" {
		args = append(args, "%"+search+"%")
		index := len(args)
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", index, index))
	}

	if filter.CreatedFrom != nil {
		args = append(args, *filter.CreatedFrom)
		index := len(args)
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", index))
	}

	if filter.CreatedTo != nil {
		args = append(args, *filter.CreatedTo)
		index := len(args)
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", index))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *RBACRepository) GetRole(ctx context.Context, id string) (rbacdomain.Role, error) {
	const query = `
		SELECT id::text, name, COALESCE(description, ''), created_at
		FROM roles
		WHERE id = $1
	`

	var role rbacdomain.Role
	err := r.pool.QueryRow(ctx, query, id).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rbacdomain.Role{}, rbacdomain.ErrNotFound
		}
		return rbacdomain.Role{}, mapRBACError(err)
	}
	return role, nil
}

func (r *RBACRepository) CreateRole(ctx context.Context, name, description string) (rbacdomain.Role, error) {
	const query = `
		INSERT INTO roles (name, description)
		VALUES ($1, $2)
		RETURNING id::text, name, COALESCE(description, ''), created_at
	`

	var role rbacdomain.Role
	desc := normalizeDescription(description)
	err := r.pool.QueryRow(ctx, query, name, desc).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return rbacdomain.Role{}, mapRBACError(err)
	}
	return role, nil
}

func (r *RBACRepository) UpdateRole(ctx context.Context, id, name, description string) (rbacdomain.Role, error) {
	const query = `
		UPDATE roles
		SET name = $2, description = $3
		WHERE id = $1
		RETURNING id::text, name, COALESCE(description, ''), created_at
	`

	var role rbacdomain.Role
	desc := normalizeDescription(description)
	err := r.pool.QueryRow(ctx, query, id, name, desc).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rbacdomain.Role{}, rbacdomain.ErrNotFound
		}
		return rbacdomain.Role{}, mapRBACError(err)
	}
	return role, nil
}

func (r *RBACRepository) DeleteRole(ctx context.Context, id string) error {
	const query = `
		DELETE FROM roles
		WHERE id = $1
	`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return mapRBACError(err)
	}
	if tag.RowsAffected() == 0 {
		return rbacdomain.ErrNotFound
	}
	return nil
}

func (r *RBACRepository) ListPermissions(ctx context.Context, filter rbacdomain.ListFilterPermission) (rbacdomain.ListPermission, error) {
	where, args := buildPermissionFilters(filter)

	countQuery := `SELECT COUNT(*) FROM permissions`
	if where != "" {
		countQuery += " " + where
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return rbacdomain.ListPermission{}, err
	}

	limit := filter.Pagination.Limit()
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Pagination.Offset()
	if offset < 0 {
		offset = 0
	}

	listQuery := fmt.Sprintf(`
		SELECT id::text, name, COALESCE(description, ''), created_at
		FROM permissions
		%s
		ORDER BY name
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	listArgs := append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return rbacdomain.ListPermission{}, err
	}
	defer rows.Close()

	var permissions []rbacdomain.Permission
	for rows.Next() {
		var permission rbacdomain.Permission
		if err := rows.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt); err != nil {
			return rbacdomain.ListPermission{}, err
		}
		permissions = append(permissions, permission)
	}
	return rbacdomain.ListPermission{
		Permission: permissions,
		Total:      total,
	}, nil
}

func buildPermissionFilters(filter rbacdomain.ListFilterPermission) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if search := strings.TrimSpace(filter.Search); search != "" {
		args = append(args, "%"+search+"%")
		index := len(args)
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", index, index))
	}

	if filter.CreatedFrom != nil {
		args = append(args, *filter.CreatedFrom)
		index := len(args)
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", index))
	}

	if filter.CreatedTo != nil {
		args = append(args, *filter.CreatedTo)
		index := len(args)
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", index))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *RBACRepository) GetPermission(ctx context.Context, id string) (rbacdomain.Permission, error) {
	const query = `
		SELECT id::text, name, COALESCE(description, ''), created_at
		FROM permissions
		WHERE id = $1
	`

	var permission rbacdomain.Permission
	err := r.pool.QueryRow(ctx, query, id).Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rbacdomain.Permission{}, rbacdomain.ErrNotFound
		}
		return rbacdomain.Permission{}, mapRBACError(err)
	}
	return permission, nil
}

func (r *RBACRepository) CreatePermission(ctx context.Context, name, description string) (rbacdomain.Permission, error) {
	const query = `
		INSERT INTO permissions (name, description)
		VALUES ($1, $2)
		RETURNING id::text, name, COALESCE(description, ''), created_at
	`

	var permission rbacdomain.Permission
	desc := normalizeDescription(description)
	err := r.pool.QueryRow(ctx, query, name, desc).Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt)
	if err != nil {
		return rbacdomain.Permission{}, mapRBACError(err)
	}
	return permission, nil
}

func (r *RBACRepository) UpdatePermission(ctx context.Context, id, name, description string) (rbacdomain.Permission, error) {
	const query = `
		UPDATE permissions
		SET name = $2, description = $3
		WHERE id = $1
		RETURNING id::text, name, COALESCE(description, ''), created_at
	`

	var permission rbacdomain.Permission
	desc := normalizeDescription(description)
	err := r.pool.QueryRow(ctx, query, id, name, desc).Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rbacdomain.Permission{}, rbacdomain.ErrNotFound
		}
		return rbacdomain.Permission{}, mapRBACError(err)
	}
	return permission, nil
}

func (r *RBACRepository) DeletePermission(ctx context.Context, id string) error {
	const query = `
		DELETE FROM permissions
		WHERE id = $1
	`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return mapRBACError(err)
	}
	if tag.RowsAffected() == 0 {
		return rbacdomain.ErrNotFound
	}
	return nil
}

func (r *RBACRepository) ListRolePermissions(ctx context.Context, roleID string) ([]rbacdomain.Permission, error) {
	if err := r.ensureRoleExists(ctx, roleID); err != nil {
		return nil, err
	}

	const query = `
		SELECT p.id::text, p.name, COALESCE(p.description, ''), p.created_at
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.name
	`

	rows, err := r.pool.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []rbacdomain.Permission
	for rows.Next() {
		var permission rbacdomain.Permission
		if err := rows.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *RBACRepository) ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := ensureRoleExistsTx(ctx, tx, roleID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
		return mapRBACError(err)
	}

	if len(permissionIDs) > 0 {
		batch := &pgx.Batch{}
		queued := 0
		for _, permissionID := range permissionIDs {
			if strings.TrimSpace(permissionID) == "" {
				continue
			}
			batch.Queue(
				`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				roleID,
				permissionID,
			)
			queued++
		}
		results := tx.SendBatch(ctx, batch)
		for i := 0; i < queued; i++ {
			if _, err := results.Exec(); err != nil {
				_ = results.Close()
				return mapRBACError(err)
			}
		}
		if err := results.Close(); err != nil {
			return mapRBACError(err)
		}
	}

	return tx.Commit(ctx)
}

func (r *RBACRepository) ensureRoleExists(ctx context.Context, roleID string) error {
	return ensureRoleExistsTx(ctx, r.pool, roleID)
}

func ensureRoleExistsTx(ctx context.Context, querier interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}, roleID string) error {
	var exists bool
	if err := querier.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM roles WHERE id = $1)`, roleID).Scan(&exists); err != nil {
		return mapRBACError(err)
	}
	if !exists {
		return rbacdomain.ErrNotFound
	}
	return nil
}

func normalizeDescription(value string) interface{} {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func mapRBACError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return rbacdomain.ErrConflict
		case "23503":
			return rbacdomain.ErrInvalidInput
		case "22P02":
			return rbacdomain.ErrInvalidInput
		}
	}
	return err
}
