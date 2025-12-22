package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	redisinfra "github.com/mzulfanw/boilerplate-go-fiber/infrastructure/redis"
	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	authdomain "github.com/mzulfanw/boilerplate-go-fiber/internal/domain/auth"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid refresh token")
	ErrInvalidAccessToken = errors.New("invalid access token")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrUserLocked         = errors.New("user is locked")
	ErrInvalidResetToken  = errors.New("invalid reset token")
	ErrInvalidPassword    = errors.New("invalid password")
)

const passwordHashCost = 12

type Service struct {
	repo             Repository
	tokenManager     TokenManager
	cache            redisinfra.Cache
	refreshTTL       time.Duration
	passwordResetTTL time.Duration
	resetCooldown    time.Duration
	maxLoginAttempts int
	lockoutDuration  time.Duration
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

func NewService(cfg config.Config, repo Repository, tokenManager TokenManager) (*Service, error) {
	if repo == nil {
		return nil, errors.New("auth: repository is nil")
	}
	if tokenManager == nil {
		return nil, errors.New("auth: token manager is nil")
	}

	return &Service{
		repo:             repo,
		tokenManager:     tokenManager,
		cache:            nil,
		refreshTTL:       cfg.RefreshTokenTTL,
		passwordResetTTL: cfg.AuthPasswordResetTTL,
		resetCooldown:    cfg.AuthPasswordResetCooldown,
		maxLoginAttempts: cfg.AuthMaxLoginAttempts,
		lockoutDuration:  cfg.AuthLockoutDuration,
	}, nil
}

func NewServiceWithCache(cfg config.Config, repo Repository, tokenManager TokenManager, cache redisinfra.Cache) (*Service, error) {
	service, err := NewService(cfg, repo, tokenManager)
	if err != nil {
		return nil, err
	}
	service.cache = cache
	return service, nil
}

func (s *Service) Login(ctx context.Context, email, password, ip, userAgent string) (TokenPair, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if normalizedEmail == "" || password == "" {
		return TokenPair{}, ErrInvalidCredentials
	}

	user, err := s.repo.FindUserByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, authdomain.ErrNotFound) {
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, err
	}
	if !user.IsActive {
		return TokenPair{}, ErrUserDisabled
	}
	if user.LockedUntil != nil {
		if user.LockedUntil.After(time.Now()) {
			return TokenPair{}, ErrUserLocked
		}
		if s.maxLoginAttempts > 0 {
			if err := s.repo.ResetLoginFailures(ctx, user.ID); err != nil {
				return TokenPair{}, err
			}
		}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if s.maxLoginAttempts > 0 {
			lockSeconds := int64(s.lockoutDuration.Seconds())
			if err := s.repo.RecordLoginFailure(ctx, user.ID, s.maxLoginAttempts, lockSeconds); err != nil {
				return TokenPair{}, err
			}
		}
		return TokenPair{}, ErrInvalidCredentials
	}
	if s.maxLoginAttempts > 0 {
		if err := s.repo.ResetLoginFailures(ctx, user.ID); err != nil {
			return TokenPair{}, err
		}
	}

	roles, err := s.repo.ListUserRoles(ctx, user.ID)
	if err != nil {
		return TokenPair{}, err
	}
	perms, err := s.repo.ListUserPermissions(ctx, user.ID)
	if err != nil {
		return TokenPair{}, err
	}

	accessToken, expiresIn, err := s.createAccessToken(user.ID, roles, perms, user.TokenVersion)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, tokenHash, expiresAt, err := s.newRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.repo.CreateRefreshToken(ctx, authdomain.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		IPAddress: ip,
		UserAgent: userAgent,
	}); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken, ip, userAgent string) (TokenPair, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return TokenPair{}, ErrInvalidToken
	}

	tokenHash := hashToken(refreshToken)
	storedToken, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, authdomain.ErrNotFound) {
			return TokenPair{}, ErrInvalidToken
		}
		return TokenPair{}, err
	}
	if storedToken.RevokedAt != nil || time.Now().After(storedToken.ExpiresAt) {
		if storedToken.RevokedAt != nil && storedToken.ReplacedByHash != "" {
			if err := s.repo.RevokeAllRefreshTokens(ctx, storedToken.UserID); err != nil {
				return TokenPair{}, err
			}
		}
		return TokenPair{}, ErrInvalidToken
	}

	user, err := s.repo.FindUserByID(ctx, storedToken.UserID)
	if err != nil {
		if errors.Is(err, authdomain.ErrNotFound) {
			return TokenPair{}, ErrInvalidToken
		}
		return TokenPair{}, err
	}
	if !user.IsActive {
		return TokenPair{}, ErrInvalidToken
	}
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return TokenPair{}, ErrInvalidToken
	}

	roles, err := s.repo.ListUserRoles(ctx, user.ID)
	if err != nil {
		return TokenPair{}, err
	}
	perms, err := s.repo.ListUserPermissions(ctx, user.ID)
	if err != nil {
		return TokenPair{}, err
	}

	accessToken, expiresIn, err := s.createAccessToken(user.ID, roles, perms, user.TokenVersion)
	if err != nil {
		return TokenPair{}, err
	}

	newRefreshToken, newTokenHash, expiresAt, err := s.newRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.repo.RevokeRefreshToken(ctx, tokenHash, newTokenHash); err != nil {
		return TokenPair{}, err
	}
	if err := s.repo.CreateRefreshToken(ctx, authdomain.RefreshToken{
		UserID:    user.ID,
		TokenHash: newTokenHash,
		ExpiresAt: expiresAt,
		IPAddress: ip,
		UserAgent: userAgent,
	}); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}

	tokenHash := hashToken(refreshToken)
	if err := s.repo.RevokeRefreshToken(ctx, tokenHash, ""); err != nil {
		return err
	}
	return nil
}

func (s *Service) ParseAccessToken(tokenString string) (AccessClaims, error) {
	if s == nil || s.tokenManager == nil {
		return AccessClaims{}, ErrInvalidAccessToken
	}
	claims, err := s.tokenManager.ParseAccessToken(tokenString)
	if err != nil {
		return AccessClaims{}, ErrInvalidAccessToken
	}
	return claims, nil
}

func (s *Service) ValidateAccessToken(ctx context.Context, tokenString string) (AccessClaims, error) {
	claims, err := s.ParseAccessToken(tokenString)
	if err != nil {
		return AccessClaims{}, err
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return AccessClaims{}, ErrInvalidAccessToken
	}

	state, err := s.repo.GetUserAuthState(ctx, claims.Subject)
	if err != nil {
		if errors.Is(err, authdomain.ErrNotFound) {
			return AccessClaims{}, ErrInvalidAccessToken
		}
		return AccessClaims{}, err
	}
	if !state.IsActive {
		return AccessClaims{}, ErrUserDisabled
	}
	if state.TokenVersion != claims.TokenVersion {
		return AccessClaims{}, ErrInvalidAccessToken
	}

	return claims, nil
}

func (s *Service) CleanupExpiredRefreshTokens(ctx context.Context) (int64, error) {
	return s.repo.DeleteExpiredRefreshTokens(ctx)
}

type PasswordResetRequest struct {
	Email      string
	Token      string
	ExpiresAt  time.Time
	ShouldSend bool
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) (PasswordResetRequest, error) {
	normalized := strings.TrimSpace(strings.ToLower(email))
	if normalized == "" {
		return PasswordResetRequest{}, ErrInvalidCredentials
	}
	if s.cache == nil {
		return PasswordResetRequest{}, errors.New("auth: cache is nil")
	}

	user, err := s.repo.FindUserByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, authdomain.ErrNotFound) {
			return PasswordResetRequest{ShouldSend: false}, nil
		}
		return PasswordResetRequest{}, err
	}
	if !user.IsActive {
		return PasswordResetRequest{ShouldSend: false}, nil
	}

	if s.resetCooldown > 0 {
		cooldownKey := fmt.Sprintf("auth:password_reset:cooldown:%s", user.ID)
		ok, err := s.cache.SetIfNotExists(ctx, cooldownKey, "1", s.resetCooldown)
		if err != nil {
			return PasswordResetRequest{}, err
		}
		if !ok {
			return PasswordResetRequest{ShouldSend: false}, nil
		}
	}

	rawToken, err := generateToken(32)
	if err != nil {
		return PasswordResetRequest{}, err
	}
	tokenHash := hashToken(rawToken)
	tokenKey := fmt.Sprintf("auth:password_reset:token:%s", tokenHash)
	userKey := fmt.Sprintf("auth:password_reset:user:%s", user.ID)

	if existing, err := s.cache.GetString(ctx, userKey); err == nil && existing != "" {
		_ = s.cache.Delete(ctx, fmt.Sprintf("auth:password_reset:token:%s", existing))
	}

	ttl := s.passwordResetTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	if err := s.cache.SetWithTTL(ctx, tokenKey, user.ID, ttl); err != nil {
		return PasswordResetRequest{}, err
	}
	if err := s.cache.SetWithTTL(ctx, userKey, tokenHash, ttl); err != nil {
		return PasswordResetRequest{}, err
	}

	return PasswordResetRequest{
		Email:      user.Email,
		Token:      rawToken,
		ExpiresAt:  time.Now().Add(ttl),
		ShouldSend: true,
	}, nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	trimmedToken := strings.TrimSpace(token)
	trimmedPassword := strings.TrimSpace(newPassword)
	if trimmedToken == "" {
		return ErrInvalidResetToken
	}
	if trimmedPassword == "" {
		return ErrInvalidPassword
	}
	if s.cache == nil {
		return errors.New("auth: cache is nil")
	}

	tokenHash := hashToken(trimmedToken)
	tokenKey := fmt.Sprintf("auth:password_reset:token:%s", tokenHash)

	userID, err := s.cache.GetString(ctx, tokenKey)
	if err != nil {
		if errors.Is(err, redisinfra.ErrKeyNotFound) {
			return ErrInvalidResetToken
		}
		return err
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, authdomain.ErrNotFound) {
			return ErrInvalidResetToken
		}
		return err
	}
	if !user.IsActive {
		return ErrInvalidResetToken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(trimmedPassword), passwordHashCost)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePassword(ctx, user.ID, string(hashed)); err != nil {
		return err
	}
	if err := s.repo.RevokeAllRefreshTokens(ctx, user.ID); err != nil {
		return err
	}

	_ = s.cache.Delete(ctx, tokenKey)
	_ = s.cache.Delete(ctx, fmt.Sprintf("auth:password_reset:user:%s", user.ID))
	return nil
}

func (s *Service) createAccessToken(userID string, roles, permissions []string, tokenVersion int) (string, int64, error) {
	if s == nil || s.tokenManager == nil {
		return "", 0, errors.New("auth: token manager is nil")
	}
	return s.tokenManager.GenerateAccessToken(userID, roles, permissions, tokenVersion)
}

func (s *Service) newRefreshToken() (string, string, time.Time, error) {
	raw, err := generateToken(32)
	if err != nil {
		return "", "", time.Time{}, err
	}
	return raw, hashToken(raw), time.Now().Add(s.refreshTTL), nil
}

func generateToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
