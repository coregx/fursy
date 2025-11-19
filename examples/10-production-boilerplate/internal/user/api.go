package user

import (
	"encoding/json"
	"net/http"
	"strconv"

	"example.com/production-boilerplate/internal/shared/auth"
	"github.com/coregx/fursy"
)

// API handles HTTP requests for user endpoints.
type API struct {
	service Service
}

// NewAPI creates a new User API handler.
func NewAPI(service Service) *API {
	return &API{service: service}
}

// RegisterRoutes registers user routes.
func (api *API) RegisterRoutes(r *fursy.Router, authMiddleware fursy.HandlerFunc) {
	// Public routes
	r.POST("/api/auth/register", api.register)
	r.POST("/api/auth/login", api.login)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(authMiddleware)
	{
		protected.GET("/users/me", api.getProfile)
		protected.PUT("/users/me", api.updateProfile)
		protected.POST("/users/me/password", api.changePassword)

		// Admin routes (with role middleware)
		admin := protected.Group("/users")
		admin.Use(auth.RequireRole("admin"))
		{
			admin.GET("", api.listUsers)
			admin.POST("/:id/ban", api.banUser)
			admin.POST("/:id/promote", api.promoteToAdmin)
		}
	}
}

// register handles user registration.
func (api *API) register(c *fursy.Context) error {
	var req RegisterRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		return c.Problem(fursy.BadRequest("Invalid Request: " + err.Error(),))
	}

	user, err := api.service.Register(c.Request.Context(), req)
	if err != nil {
		return MapError(c, err)
	}

	resp := ToResponse(user)
	return c.JSON(http.StatusCreated, resp)
}

// login handles user login.
func (api *API) login(c *fursy.Context) error {
	var req LoginRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		return c.Problem(fursy.BadRequest("Invalid Request: " + err.Error(),))
	}

	token, err := api.service.Login(c.Request.Context(), req)
	if err != nil {
		return MapError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": token,
	})
}

// getProfile retrieves current user profile.
func (api *API) getProfile(c *fursy.Context) error {
	// Get user ID from JWT claims (set by auth middleware)
	userID := auth.GetUserID(c.Request.Context())

	user, err := api.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		return MapError(c, err)
	}

	resp := ToResponse(user)
	return c.JSON(http.StatusOK, resp)
}

// updateProfile updates current user profile.
func (api *API) updateProfile(c *fursy.Context) error {
	userID := auth.GetUserID(c.Request.Context())

	var req UpdateProfileRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		return c.Problem(fursy.BadRequest("Invalid Request: " + err.Error(),))
	}

	user, err := api.service.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		return MapError(c, err)
	}

	resp := ToResponse(user)
	return c.JSON(http.StatusOK, resp)
}

// changePassword changes current user password.
func (api *API) changePassword(c *fursy.Context) error {
	userID := auth.GetUserID(c.Request.Context())

	var req ChangePasswordRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		return c.Problem(fursy.BadRequest("Invalid Request: " + err.Error(),))
	}

	err := api.service.ChangePassword(c.Request.Context(), userID, req)
	if err != nil {
		return MapError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// listUsers lists all users (admin only).
func (api *API) listUsers(c *fursy.Context) error {
	// Parse pagination from query params
	offset, _ := strconv.Atoi(c.Query("offset")); if offset == 0 { offset = 0 }
	limit, _ := strconv.Atoi(c.Query("limit")); if limit == 0 { limit = 20 }

	users, total, err := api.service.ListUsers(c.Request.Context(), offset, limit)
	if err != nil {
		return MapError(c, err)
	}

	// Convert to response DTOs
	resp := make([]UserResponse, len(users))
	for i, user := range users {
		resp[i] = ToResponse(user)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users":  resp,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

// banUser bans a user (admin only).
func (api *API) banUser(c *fursy.Context) error {
	userID := c.Param("id")

	err := api.service.BanUser(c.Request.Context(), userID)
	if err != nil {
		return MapError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User banned successfully",
	})
}

// promoteToAdmin promotes a user to admin (admin only).
func (api *API) promoteToAdmin(c *fursy.Context) error {
	userID := c.Param("id")

	err := api.service.PromoteToAdmin(c.Request.Context(), userID)
	if err != nil {
		return MapError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User promoted to admin successfully",
	})
}
