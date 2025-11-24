package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// Repository defines data access for User aggregate.
type Repository interface {
	// Get retrieves user by ID.
	Get(ctx context.Context, id string) (*User, error)

	// GetByEmail retrieves user by email.
	GetByEmail(ctx context.Context, email Email) (*User, error)

	// List retrieves users with pagination.
	List(ctx context.Context, offset, limit int) ([]*User, error)

	// Count returns total number of users.
	Count(ctx context.Context) (int, error)

	// Create saves new user.
	Create(ctx context.Context, user *User) error

	// Update updates existing user.
	Update(ctx context.Context, user *User) error

	// Delete deletes user.
	Delete(ctx context.Context, id string) error
}

// repositoryImpl implements Repository using database/sql.
type repositoryImpl struct {
	db *sql.DB
}

// NewRepository creates a new User repository.
func NewRepository(db *sql.DB) Repository {
	return &repositoryImpl{db: db}
}

// Get retrieves user by ID.
func (r *repositoryImpl) Get(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, email, password_hash, name, role, status, created_at, updated_at
	          FROM users WHERE id = ?`

	var (
		userID       string
		email        string
		passwordHash string
		name         string
		role         string
		status       string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&userID, &email, &passwordHash, &name, &role, &status, &createdAt, &updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        userID,
		Email:     emailVO,
		Password:  FromHash(passwordHash),
		Name:      name,
		Role:      Role(role),
		Status:    Status(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// GetByEmail retrieves user by email.
func (r *repositoryImpl) GetByEmail(ctx context.Context, email Email) (*User, error) {
	query := `SELECT id, email, password_hash, name, role, status, created_at, updated_at
	          FROM users WHERE email = ?`

	var (
		userID       string
		userEmail    string
		passwordHash string
		name         string
		role         string
		status       string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := r.db.QueryRowContext(ctx, query, email.String()).Scan(
		&userID, &userEmail, &passwordHash, &name, &role, &status, &createdAt, &updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	emailVO, err := NewEmail(userEmail)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        userID,
		Email:     emailVO,
		Password:  FromHash(passwordHash),
		Name:      name,
		Role:      Role(role),
		Status:    Status(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// List retrieves users with pagination.
func (r *repositoryImpl) List(ctx context.Context, offset, limit int) ([]*User, error) {
	query := `SELECT id, email, password_hash, name, role, status, created_at, updated_at
	          FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		var (
			userID       string
			email        string
			passwordHash string
			name         string
			role         string
			status       string
			createdAt    time.Time
			updatedAt    time.Time
		)

		if err := rows.Scan(&userID, &email, &passwordHash, &name, &role, &status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		emailVO, err := NewEmail(email)
		if err != nil {
			return nil, err
		}

		users = append(users, &User{
			ID:        userID,
			Email:     emailVO,
			Password:  FromHash(passwordHash),
			Name:      name,
			Role:      Role(role),
			Status:    Status(status),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	return users, rows.Err()
}

// Count returns total number of users.
func (r *repositoryImpl) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// Create saves new user.
func (r *repositoryImpl) Create(ctx context.Context, user *User) error {
	query := `INSERT INTO users (id, email, password_hash, name, role, status, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email.String(),
		user.Password.Hash(),
		user.Name,
		string(user.Role),
		string(user.Status),
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

// Update updates existing user.
func (r *repositoryImpl) Update(ctx context.Context, user *User) error {
	query := `UPDATE users SET email = ?, password_hash = ?, name = ?, role = ?, status = ?, updated_at = ?
	          WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		user.Email.String(),
		user.Password.Hash(),
		user.Name,
		string(user.Role),
		string(user.Status),
		user.UpdatedAt,
		user.ID,
	)

	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrUserAlreadyExists
	}

	return err
}

// Delete deletes user.
func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}

// Domain errors.
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)
