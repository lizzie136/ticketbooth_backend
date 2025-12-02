package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"ticketbooth-backend/models"
	"ticketbooth-backend/repositories"
)

type UserHandler struct {
	userRepo   *repositories.UserRepository
	authSecret string
}

func NewUserHandler(userRepo *repositories.UserRepository, authSecret string) *UserHandler {
	return &UserHandler{
		userRepo:   userRepo,
		authSecret: authSecret,
	}
}

// CreateUser handles POST /api/users (admin/internal creation)
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	resp, ok := h.processUserCreation(&req, w)
	if !ok {
		return
	}

	JSON(w, http.StatusCreated, resp)
}

// SignUp handles POST /api/signup (public registration)
func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	resp, ok := h.processUserCreation(&req, w)
	if !ok {
		return
	}

	JSON(w, http.StatusCreated, resp)
}

// UpdateUser handles PUT /api/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		BadRequest(w, "Invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	updates := make(map[string]interface{})

	if req.FirstName != nil {
		firstName := strings.TrimSpace(*req.FirstName)
		if firstName == "" {
			BadRequest(w, "firstName cannot be empty")
			return
		}
		updates["name"] = firstName
	}

	if req.LastName != nil {
		lastName := strings.TrimSpace(*req.LastName)
		if lastName == "" {
			BadRequest(w, "lastName cannot be empty")
			return
		}
		updates["last_name"] = lastName
	}

	if req.Email != nil {
		email := normalizeEmail(*req.Email)
		if email == "" {
			BadRequest(w, "email must be valid")
			return
		}
		updates["email"] = email
	}

	if req.Username != nil {
		username := normalizeUsername(*req.Username)
		if username == "" {
			BadRequest(w, "username must be valid")
			return
		}
		updates["username"] = username
	}

	if len(updates) == 0 {
		BadRequest(w, "At least one field must be provided for update")
		return
	}

	if err := h.userRepo.UpdateUser(id, updates); err != nil {
		if err == sql.ErrNoRows {
			NotFound(w, "User not found")
			return
		}
		if isDuplicateEntryError(err) {
			Conflict(w, "USER_EXISTS", "A user with those details already exists")
			return
		}
		InternalServerError(w, "Failed to update user")
		return
	}

	updatedUser, err := h.userRepo.GetUserByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			NotFound(w, "User not found")
			return
		}
		InternalServerError(w, "Failed to load updated user")
		return
	}

	JSON(w, http.StatusOK, userToResponse(updatedUser))
}

func userToResponse(user *models.User) *models.UserResponse {
	resp := &models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}

	if user.CreatedAt != nil {
		created := user.CreatedAt.Format(time.RFC3339)
		resp.CreatedAt = &created
	}

	if user.UpdatedAt != nil {
		updated := user.UpdatedAt.Format(time.RFC3339)
		resp.UpdatedAt = &updated
	}

	return resp
}

func isDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "Duplicate entry")
}

func (h *UserHandler) processUserCreation(req *models.CreateUserRequest, w http.ResponseWriter) (*models.UserResponse, bool) {
	firstName, lastName, email, username, password, errMsg := normalizeCreateUserPayload(req)
	if errMsg != "" {
		BadRequest(w, errMsg)
		return nil, false
	}

	resp, err := h.createUserRecord(firstName, lastName, email, username, password)
	if err != nil {
		if isDuplicateEntryError(err) {
			Conflict(w, "USER_EXISTS", "A user with that email already exists")
			return nil, false
		}
		InternalServerError(w, "Failed to create user")
		return nil, false
	}

	return resp, true
}

func (h *UserHandler) createUserRecord(firstName, lastName, email, username, password string) (*models.UserResponse, error) {
	user := &models.User{
		Username:       username,
		FirstName:      firstName,
		LastName:       lastName,
		Email:          email,
		HashedPassword: h.hashPassword(password),
	}

	userID, err := h.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	createdUser, err := h.userRepo.GetUserByID(int(userID))
	if err != nil {
		return nil, err
	}

	return userToResponse(createdUser), nil
}

func normalizeCreateUserPayload(req *models.CreateUserRequest) (string, string, string, string, string, string) {
	firstName := strings.TrimSpace(req.FirstName)
	if firstName == "" {
		return "", "", "", "", "", "firstName is required"
	}

	lastName := strings.TrimSpace(req.LastName)
	if lastName == "" {
		return "", "", "", "", "", "lastName is required"
	}

	email := normalizeEmail(req.Email)
	if email == "" {
		return "", "", "", "", "", "email must be valid"
	}

	username := normalizeUsername(req.Username)
	if username == "" {
		return "", "", "", "", "", "username must be provided"
	}

	password := strings.TrimSpace(req.Password)
	if len(password) < 8 {
		return "", "", "", "", "", "password must be at least 8 characters"
	}

	return firstName, lastName, email, username, password, ""
}

func normalizeEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return ""
	}
	return email
}

func normalizeUsername(username string) string {
	username = strings.TrimSpace(strings.ToLower(username))
	if username == "" {
		return ""
	}
	return username
}

func (h *UserHandler) hashPassword(password string) string {
	sum := sha256.Sum256([]byte(h.authSecret + ":" + password))
	return hex.EncodeToString(sum[:])
}

func (h *UserHandler) generateToken(user *models.User) string {
	payload := fmt.Sprintf("%d:%s:%d", user.ID, user.Email, time.Now().Unix())
	sig := sha256.Sum256([]byte(h.authSecret + ":" + payload))
	token := payload + ":" + hex.EncodeToString(sig[:])
	return base64.StdEncoding.EncodeToString([]byte(token))
}

// Login handles POST /api/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Password) == "" {
		BadRequest(w, "password is required")
		return
	}

	var user *models.User
	var err error

	if req.Email != nil && strings.TrimSpace(*req.Email) != "" {
		email := normalizeEmail(*req.Email)
		if email == "" {
			BadRequest(w, "email must be valid")
			return
		}
		user, err = h.userRepo.GetUserByEmail(email)
	} else if req.Username != nil && strings.TrimSpace(*req.Username) != "" {
		username := normalizeUsername(*req.Username)
		if username == "" {
			BadRequest(w, "username must be valid")
			return
		}
		user, err = h.userRepo.GetUserByUsername(username)
	} else {
		BadRequest(w, "email or username is required")
		return
	}

	if err != nil {
		if err == sql.ErrNoRows {
			Unauthorized(w, "Invalid credentials")
			return
		}
		InternalServerError(w, "Failed to fetch user")
		return
	}

	if h.hashPassword(req.Password) != user.HashedPassword {
		Unauthorized(w, "Invalid credentials")
		return
	}

	token := h.generateToken(user)
	response := &models.LoginResponse{
		Token: token,
		User:  userToResponse(user),
	}
	JSON(w, http.StatusOK, response)
}
