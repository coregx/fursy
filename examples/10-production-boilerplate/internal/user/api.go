package user

import (
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
	authGroup := r.Group("/api/auth")
	authGroup.POST("/login", api.login)
	authGroup.POST("/register", api.register)

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
func (api *API) register(c *fursy.Box[RegisterRequest, UserResponse]) error {
	user, err := api.service.Register(c.Request.Context(), *c.ReqBody)
	if err != nil {
		return MapError(c.Context, err)
	}

	// Todo: fix this dummy url.
	return c.Created("/api/users/me", ToResponse(user))
}

// login handles user login.
func (api *API) login(c *fursy.Box[LoginRequest, LoginResponse]) error {
	token, err := api.service.Login(c.Request.Context(), *c.ReqBody)
	if err != nil {
		return MapError(c.Context, err)
	}

	return c.OK(LoginResponse{Token: token})
}

// getProfile retrieves current user profile.
func (api *API) getProfile(c *fursy.Box[fursy.Empty, UserResponse]) error {
	// Get user ID from JWT claims (set by auth middleware)
	userID := auth.GetUserID(c.Request.Context())

	user, err := api.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		return MapError(c.Context, err)
	}

	return c.OK(ToResponse(user))
}

// updateProfile updates current user profile.
func (api *API) updateProfile(c *fursy.Box[UpdateProfileRequest, UserResponse]) error {
	userID := auth.GetUserID(c.Request.Context())

	user, err := api.service.UpdateProfile(c.Request.Context(), userID, *c.ReqBody)
	if err != nil {
		return MapError(c.Context, err)
	}

	return c.OK(ToResponse(user))
}

// changePassword changes current user password.
func (api *API) changePassword(c *fursy.Box[ChangePasswordRequest, MessageResponse]) error {
	userID := auth.GetUserID(c.Request.Context())

	err := api.service.ChangePassword(c.Request.Context(), userID, *c.ReqBody)
	if err != nil {
		return MapError(c.Context, err)
	}

	return c.OK(MessageResponse{Message: "Password changed successfully"})
}

// listUsers lists all users (admin only).
func (api *API) listUsers(c *fursy.Box[fursy.Empty, ListUsersResponse]) error {
	// Parse pagination from query params
	offset, _ := strconv.Atoi(c.Query("offset"))
	if offset == 0 {
		offset = 0
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 20
	}

	users, total, err := api.service.ListUsers(c.Request.Context(), offset, limit)
	if err != nil {
		return MapError(c.Context, err)
	}

	// Convert to response DTOs
	resp := make([]UserResponse, len(users))
	for i, user := range users {
		resp[i] = ToResponse(user)
	}

	return c.OK(ListUsersResponse{Limit: limit, Offset: offset, Total: total, Users: resp})
}

// banUser bans a user (admin only).
func (api *API) banUser(c *fursy.Box[fursy.Empty, MessageResponse]) error {
	userID := c.Param("id")

	err := api.service.BanUser(c.Request.Context(), userID)
	if err != nil {
		return MapError(c.Context, err)
	}

	return c.OK(MessageResponse{Message: "User banned successfully"})
}

// promoteToAdmin promotes a user to admin (admin only).
func (api *API) promoteToAdmin(c *fursy.Box[fursy.Empty, MessageResponse]) error {
	userID := c.Param("id")

	err := api.service.PromoteToAdmin(c.Request.Context(), userID)
	if err != nil {
		return MapError(c.Context, err)
	}

	return c.OK(MessageResponse{Message: "User promoted to admin successfully"})
}
