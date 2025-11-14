package router

import (
	"authgo/controllers"
	"authgo/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures routes and middleware
func SetupRouter(ctl *controllers.Controller, authMw *middleware.AuthMiddleware) *gin.Engine {
	r := gin.Default()

	// Public auth endpoints
	r.POST("/register", ctl.Register)
	r.POST("/login", ctl.Login)

	// Routes requiring authentication
	auth := r.Group("/")
	auth.Use(authMw.AuthRequired())
	{
		// All authenticated users can read tasks
		auth.GET("/tasks", ctl.GetTasks)
		auth.GET("/tasks/:id", ctl.GetTaskByID)
	}

	// Admin-only actions
	admin := r.Group("/")
	admin.Use(authMw.AuthRequired(), authMw.RequireAdmin())
	{
		admin.POST("/tasks", ctl.CreateTask)
		admin.PUT("/tasks/:id", ctl.UpdateTask)
		admin.DELETE("/tasks/:id", ctl.DeleteTask)

		// promote endpoint
		admin.POST("/promote/:username", ctl.Promote)
	}

	return r
}
