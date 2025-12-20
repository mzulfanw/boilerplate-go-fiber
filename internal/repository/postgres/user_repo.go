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
	userdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) ListUsers(ctx context.Context, filter userdomain.ListFilter) (userdomain.ListResult, error) {
	where, args := buildUserListFilters(filter)

	countQuery := `SELECT COUNT(*) FROM users`
	if where != "" {
		countQuery += " " + where
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return userdomain.ListResult{}, err
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
		SELECT id::text, email, is_active, created_at, updated_at
		FROM users
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	listArgs := append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return userdomain.ListResult{}, err
	}
	defer rows.Close()

	users := make([]userdomain.User, 0)
	for rows.Next() {
		var user userdomain.User
		if err := rows.Scan(&user.ID, &user.Email, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return userdomain.ListResult{}, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return userdomain.ListResult{}, err
	}

	return userdomain.ListResult{
		Users: users,
		Total: total,
	}, nil
}

func buildUserListFilters(filter userdomain.ListFilter) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)
	index := 1

	if search := strings.TrimSpace(filter.Search); search != "" {
		args = append(args, "%"+search+"%")
		conditions = append(conditions, fmt.Sprintf("(email ILIKE $%d OR id::text ILIKE $%d)", index, index))
		index++
	}

	if filter.IsActive != nil {
		args = append(args, *filter.IsActive)
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", index))
		index++
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (userdomain.User, error) {
	const query = `
		SELECT id::text, email, password_hash, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user userdomain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userdomain.User{}, userdomain.ErrNotFound
		}
		return userdomain.User{}, mapUserError(err)
	}
	return user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string, isActive bool, roleIDs []string) (userdomain.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return userdomain.User{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const query = `
		INSERT INTO users (email, password_hash, is_active)
		VALUES ($1, $2, $3)
		RETURNING id::text, email, password_hash, is_active, created_at, updated_at
	`

	var user userdomain.User
	err = tx.QueryRow(ctx, query, email, passwordHash, isActive).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return userdomain.User{}, mapUserError(err)
	}

	if len(roleIDs) > 0 {
		batch := &pgx.Batch{}
		queued := 0
		for _, roleID := range roleIDs {
			if strings.TrimSpace(roleID) == "" {
				continue
			}
			batch.Queue(
				`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				user.ID,
				roleID,
			)
			queued++
		}
		results := tx.SendBatch(ctx, batch)
		for i := 0; i < queued; i++ {
			if _, err := results.Exec(); err != nil {
				_ = results.Close()
				return userdomain.User{}, mapUserError(err)
			}
		}
		if err := results.Close(); err != nil {
			return userdomain.User{}, mapUserError(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return userdomain.User{}, err
	}
	return user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, id, email, passwordHash string, isActive bool, bumpTokenVersion bool) (userdomain.User, error) {
	const query = `
		UPDATE users
		SET email = $2,
			password_hash = $3,
			is_active = $4,
			updated_at = now(),
			token_version = CASE WHEN $5 THEN token_version + 1 ELSE token_version END
		WHERE id = $1
		RETURNING id::text, email, password_hash, is_active, created_at, updated_at
	`

	var user userdomain.User
	err := r.pool.QueryRow(ctx, query, id, email, passwordHash, isActive, bumpTokenVersion).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userdomain.User{}, userdomain.ErrNotFound
		}
		return userdomain.User{}, mapUserError(err)
	}
	return user, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	const query = `
		DELETE FROM users
		WHERE id = $1
	`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return mapUserError(err)
	}
	if tag.RowsAffected() == 0 {
		return userdomain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) ListUserRoles(ctx context.Context, userID string) ([]rbacdomain.Role, error) {
	if err := r.ensureUserExists(ctx, userID); err != nil {
		return nil, err
	}

	const query = `
		SELECT r.id::text, r.name, COALESCE(r.description, ''), r.created_at
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []rbacdomain.Role
	for rows.Next() {
		var role rbacdomain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *UserRepository) ReplaceUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := ensureUserExistsTx(ctx, tx, userID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID); err != nil {
		return mapUserError(err)
	}

	if len(roleIDs) > 0 {
		batch := &pgx.Batch{}
		queued := 0
		for _, roleID := range roleIDs {
			if strings.TrimSpace(roleID) == "" {
				continue
			}
			batch.Queue(
				`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				userID,
				roleID,
			)
			queued++
		}
		results := tx.SendBatch(ctx, batch)
		for i := 0; i < queued; i++ {
			if _, err := results.Exec(); err != nil {
				_ = results.Close()
				return mapUserError(err)
			}
		}
		if err := results.Close(); err != nil {
			return mapUserError(err)
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE users SET token_version = token_version + 1 WHERE id = $1`, userID); err != nil {
		return mapUserError(err)
	}

	return tx.Commit(ctx)
}

func (r *UserRepository) ensureUserExists(ctx context.Context, userID string) error {
	return ensureUserExistsTx(ctx, r.pool, userID)
}

func ensureUserExistsTx(ctx context.Context, querier interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}, userID string) error {
	var exists bool
	if err := querier.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists); err != nil {
		return mapUserError(err)
	}
	if !exists {
		return userdomain.ErrNotFound
	}
	return nil
}

func mapUserError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return userdomain.ErrConflict
		case "23503":
			return userdomain.ErrInvalidInput
		case "22P02":
			return userdomain.ErrInvalidInput
		}
	}
	return err
}
