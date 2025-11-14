package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go_mango/data"
	"go_mango/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TaskController coordinates HTTP -> service
type TaskController struct {
	service *data.TaskService
}

// NewTaskController constructs controller with provided service
func NewTaskController(s *data.TaskService) *TaskController {
	return &TaskController{service: s}
}

// helper to map DB Task -> TaskResponse
func toResponse(t models.Task) models.TaskResponse {
	id := ""
	if !t.ID.IsZero() {
		id = t.ID.Hex()
	}
	return models.TaskResponse{
		ID:          id,
		Title:       t.Title,
		Description: t.Description,
		DueDate:     t.DueDate,
		Status:      t.Status,
	}
}

// GetTasks handles GET /tasks
func (tc *TaskController) GetTasks(c *gin.Context) {
	tasks, err := tc.service.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tasks"})
		return
	}
	resp := make([]models.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, toResponse(t))
	}
	c.JSON(http.StatusOK, resp)
}

// GetTaskByID handles GET /tasks/:id
func (tc *TaskController) GetTaskByID(c *gin.Context) {
	id := c.Param("id")
	// validate hex
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	task, err := tc.service.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching task"})
		return
	}
	// not found
	if task.ID.IsZero() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, toResponse(task))
}

// CreateTask handles POST /tasks
func (tc *TaskController) CreateTask(c *gin.Context) {
	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON, title required"})
		return
	}
	created, err := tc.service.CreateTask(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}
	c.JSON(http.StatusCreated, toResponse(created))
}

// UpdateTask handles PUT /tasks/:id
func (tc *TaskController) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	// validate hex
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	updated, err := tc.service.UpdateTask(id, input)
	if err != nil {
		// no fields to update
		if err.Error() == "no fields to update" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update task"})
		return
	}
	// not found
	if updated.ID.IsZero() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, toResponse(updated))
}

// DeleteTask handles DELETE /tasks/:id
func (tc *TaskController) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	// validate hex
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	ok, err := tc.service.DeleteTask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete task"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}
