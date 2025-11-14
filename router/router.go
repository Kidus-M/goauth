package router

import (
	"github.com/gin-gonic/gin"
	"go_mango/controllers"
)

// SetupRouter creates a gin.Engine and registers routes with provided controller
func SetupRouter(taskController *controllers.TaskController) *gin.Engine {
	r := gin.Default()

	// Tasks routes
	r.GET("/tasks", taskController.GetTasks)
	r.GET("/tasks/:id", taskController.GetTaskByID)
	r.POST("/tasks", taskController.CreateTask)
	r.PUT("/tasks/:id", taskController.UpdateTask)
	r.DELETE("/tasks/:id", taskController.DeleteTask)

	return r
}
