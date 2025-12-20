package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	authdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/auth"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (authdomain.User, error) {
	const query = `
		SELECT id::text, email, password_hash, is_active, failed_login_attempts, locked_until, token_version
		FROM users
		WHERE email = $1
	`

	var user authdomain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.TokenVersion,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authdomain.User{}, authdomain.ErrNotFound
		}
		return authdomain.User{}, err
	}
	return user, nil
}

func (r *AuthRepository) FindUserByID(ctx context.Context, id string) (authdomain.User, error) {
	const query = `
		SELECT id::text, email, password_hash, is_active, failed_login_attempts, locked_until, token_version
		FROM users
		WHERE id = $1
	`

	var user authdomain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.TokenVersion,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authdomain.User{}, authdomain.ErrNotFound
		}
		return authdomain.User{}, err
	}
	return user, nil
}

func (r *AuthRepository) GetUserAuthState(ctx context.Context, id string) (authdomain.AuthState, error) {
	const query = `
		SELECT is_active, token_version
		FROM users
		WHERE id = $1
	`

	var state authdomain.AuthState
	err := r.pool.QueryRow(ctx, query, id).Scan(&state.IsActive, &state.TokenVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authdomain.AuthState{}, authdomain.ErrNotFound
		}
		return authdomain.AuthState{}, err
	}
	return state, nil
}

func (r *AuthRepository) ListUserRoles(ctx context.Context, userID string) ([]string, error) {
	const query = `
		SELECT r.name
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, name)
	}
	return roles, rows.Err()
}

func (r *AuthRepository) ListUserPermissions(ctx context.Context, userID string) ([]string, error) {
	const query = `
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		permissions = append(permissions, name)
	}
	return permissions, rows.Err()
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token authdomain.RefreshToken) error {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt, token.IPAddress, token.UserAgent)
	return err
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, tokenHash string) (authdomain.RefreshToken, error) {
	const query = `
		SELECT id::text, user_id::text, token_hash, COALESCE(replaced_by_hash, ''), expires_at, revoked_at, created_at, ip_address, user_agent
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var token authdomain.RefreshToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ReplacedByHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
		&token.IPAddress,
		&token.UserAgent,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authdomain.RefreshToken{}, authdomain.ErrNotFound
		}
		return authdomain.RefreshToken{}, err
	}
	return token, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash, replacedByHash string) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = now(), replaced_by_hash = NULLIF($2, '')
		WHERE token_hash = $1 AND revoked_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, tokenHash, replacedByHash)
	return err
}

func (r *AuthRepository) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *AuthRepository) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	const query = `
		DELETE FROM refresh_tokens
		WHERE expires_at < now()
	`

	tag, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *AuthRepository) RecordLoginFailure(ctx context.Context, userID string, maxAttempts int, lockoutSeconds int64) error {
	const query = `
		UPDATE users
		SET failed_login_attempts = failed_login_attempts + 1,
			locked_until = CASE
				WHEN $2 > 0 AND failed_login_attempts + 1 >= $2 AND $3 > 0
					THEN now() + ($3 || ' seconds')::interval
				ELSE locked_until
			END
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID, maxAttempts, lockoutSeconds)
	return err
}

func (r *AuthRepository) ResetLoginFailures(ctx context.Context, userID string) error {
	const query = `
		UPDATE users
		SET failed_login_attempts = 0, locked_until = NULL
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
