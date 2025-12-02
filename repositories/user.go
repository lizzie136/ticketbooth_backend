package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ticketbooth-backend/db"
	"ticketbooth-backend/models"
)

type UserRepository struct {
	db *db.DB
}

func NewUserRepository(db *db.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user record and returns the insert ID.
func (r *UserRepository) CreateUser(user *models.User) (int64, error) {
	query := "INSERT INTO `user` (username, name, last_name, email, hashed_password, date_created, date_updated) " +
		"VALUES (?, ?, ?, ?, ?, NOW(), NOW())"

	result, err := r.db.Exec(query, user.Username, user.FirstName, user.LastName, user.Email, user.HashedPassword)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateUser updates provided fields for a user.
func (r *UserRepository) UpdateUser(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	setClauses := make([]string, 0, len(updates)+1)
	args := make([]interface{}, 0, len(updates)+1)

	for column, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", column))
		args = append(args, value)
	}
	setClauses = append(setClauses, "date_updated = NOW()")

	query := fmt.Sprintf("UPDATE `user` SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetUserByID fetches a single user by ID.
func (r *UserRepository) GetUserByID(id int) (*models.User, error) {
	query := "SELECT id, username, name, last_name, email, hashed_password, date_created, date_updated FROM `user` WHERE id = ?"

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.HashedPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail fetches a user by email.
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := "SELECT id, username, name, last_name, email, hashed_password, date_created, date_updated FROM `user` WHERE email = ?"
	return r.scanUser(query, email)
}

// GetUserByUsername fetches a user by username.
func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := "SELECT id, username, name, last_name, email, hashed_password, date_created, date_updated FROM `user` WHERE username = ?"
	return r.scanUser(query, username)
}

func (r *UserRepository) scanUser(query string, arg interface{}) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(query, arg).Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.HashedPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
