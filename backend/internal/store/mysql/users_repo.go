package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
}

const (
	RoleStaff = "staff"
	RoleAdmin = "admin"
)

type UsersRepo struct {
	db *sqlx.DB
}

func NewUsersRepo(db *sqlx.DB) *UsersRepo {
	return &UsersRepo{db: db}
}

func (r *UsersRepo) Create(ctx context.Context, email, passwordHash, role string) (*User, error) {
	u := &User{ID: uuid.NewString(), Email: email, PasswordHash: passwordHash, Role: role}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, role) VALUES (?, ?, ?, ?)`,
		u.ID, u.Email, u.PasswordHash, u.Role)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return r.GetByID(ctx, u.ID)
}

func (r *UsersRepo) GetByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.db.GetContext(ctx, &u, `SELECT id, email, password_hash, role, created_at FROM users WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *UsersRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.GetContext(ctx, &u, `SELECT id, email, password_hash, role, created_at FROM users WHERE email = ?`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}
