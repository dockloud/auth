package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dockloud/auth/intermal/domain/user"
	"github.com/redis/go-redis/v9"
)

const userKeyPrefix = "user"

type Repository struct {
	db    *sql.DB
	cache *redis.Client
}

func NewRepository(db *sql.DB, cache *redis.Client) *Repository {
	return &Repository{
		db:    db,
		cache: cache,
	}
}

func (r *Repository) Create(ctx context.Context, username string, passwordHash string, email string) (*user.User, error) {
	var err error
	var u user.User

	query := `
		INSERT INTO users(username, email, password_hash)
		VALUES($1, $2, $3)
		RETURNING uid, username, email, password_hash, created_at
	`

	err = r.db.
		QueryRowContext(ctx, query, username, email, passwordHash).
		Scan(&u.UID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &u, nil
}

func (r *Repository) Get(ctx context.Context, uid string) (*user.User, error) {
	var u user.User
	var err error

	userKey := fmt.Sprintf("%s:%s", userKeyPrefix, uid)
	jsonUser, err := r.cache.Get(ctx, userKey).Result()

	switch {
	case err == nil:
		if err := json.Unmarshal([]byte(jsonUser), &u); err != nil {
			_ = r.cache.Del(ctx, userKey).Err()
			break
		}
		return &u, nil

	case errors.Is(err, redis.Nil):
		// cache miss

	default:
		break
	}

	query := `
		SELECT uid, username, email, password_hash, created_at
        FROM users
        WHERE uid = $1
	`

	err = r.db.
		QueryRowContext(ctx, query, uid).
		Scan(&u.UID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("error getting user by uid: %w", err)
	}

	marshaledUser, err := json.Marshal(u)
	if err != nil {
		return &u, nil
	}
	err = r.cache.Set(ctx, userKey, string(marshaledUser), 10*time.Minute).Err()
	if err != nil {
		// unable to put value to cache
	}

	return &u, nil
}

func (r *Repository) Update(
	ctx context.Context, uid string,
	username *string, email *string, passwordHash *string,
) (*user.User, error) {
	var u user.User
	var err error

	query := `
        UPDATE users
        SET
            username = COALESCE($1, username),
            email = COALESCE($2, email),
            password_hash = COALESCE($3, password_hash)
        WHERE uid = $4
		RETURNING uid, username, email, password_hash, created_at
    `

	err = r.db.
		QueryRowContext(ctx, query, username, email, passwordHash, uid).
		Scan(&u.UID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	userKey := fmt.Sprintf("%s:%s", userKeyPrefix, uid)
	err = r.cache.Del(ctx, userKey).Err()
	if err != nil {
		// unable to clear cache
	}

	return &u, nil
}

func (r *Repository) Delete(ctx context.Context, uid string) (*user.User, error) {
	var u user.User
	var err error

	query := `
		DELETE FROM users
		WHERE uid = $1
		RETURNING uid, username, email, password_hash, created_at
	`

	err = r.db.
		QueryRowContext(ctx, query, uid).
		Scan(&u.UID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("error deleting user: %w", err)
	}

	userKey := fmt.Sprintf("%s:%s", userKeyPrefix, uid)
	err = r.cache.Del(ctx, userKey).Err()
	if err != nil {
		// unable to clear cache
	}

	return &u, nil
}

func (r *Repository) Search(ctx context.Context, username *string, email *string) (*user.SearchResult, error) {
	var u user.SearchResult
	var err error

	query := `
	SELECT
		EXISTS (
			SELECT 1 FROM users
			WHERE $1::text IS NOT NULL AND username = $1::text
		) AS username_exists,
		EXISTS (
			SELECT 1 FROM users
			WHERE $2::text IS NOT NULL AND email = $2::text
		) AS email_exists;
	`

	err = r.db.
		QueryRowContext(ctx, query, username, email).
		Scan(&u.UIDExists, &u.UsernameExists, &u.EmailExists)

	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}

	return &u, nil
}
