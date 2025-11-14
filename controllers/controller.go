package controllers

import (
	"net/http"
	"os"
	"time"

	"authgo/data"
	"authgo/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Controller groups handlers for users and tasks
type Controller struct {
	userSvc *data.UserService
	taskSvc *data.TaskService
	secret  string
}

// NewController constructs Controller
func NewController(us *data.UserService, ts *data.TaskService) *Controller {
	return &Controller{
		userSvc: us,
		taskSvc: ts,
		secret:  os.Getenv("JWT_SECRET"),
	}
}

func tokenForUser(u models.User, secret string) (string, error) {
	claims := jwt.MapClaims{
		"username": u.Username,
		"role":     u.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"nbf":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Register handles POST /register
func (ctl *Controller) Register(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
		return
	}
	u, err := ctl.userSvc.CreateUser(input.Username, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// issue token
	tok, err := tokenForUser(u, ctl.secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"username": u.Username,
		"role":     u.Role,
		"token":    tok,
	})
}

// Login handles POST /login
func (ctl *Controller) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
		return
	}
	u, err := ctl.userSvc.Authenticate(input.Username, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	tok, err := tokenForUser(u, ctl.secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"username": u.Username,
		"role":     u.Role,
		"token":    tok,
	})
}

// Promote handles POST /promote/:username (admin only)
func (ctl *Controller) Promote(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
		return
	}
	updated, err := ctl.userSvc.PromoteUser(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to promote user"})
		return
	}
	if updated.Username == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": updated.Username, "role": updated.Role})
}

// GetTasks handles GET /tasks (authenticated: all users)
func (ctl *Controller) GetTasks(c *gin.Context) {
	tasks, err := ctl.taskSvc.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tasks"})
		return
	}
	resp := make([]models.TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		id := ""
		if !t.ID.IsZero() {
			id = t.ID.Hex()
		}
		resp = append(resp, models.TaskResponse{
			ID:          id,
			Title:       t.Title,
			Description: t.Description,
			DueDate:     t.DueDate,
			Status:      t.Status,
		})
	}
	c.JSON(http.StatusOK, resp)
}

// GetTaskByID handles GET /tasks/:id (authenticated)
func (ctl *Controller) GetTaskByID(c *gin.Context) {
	id := c.Param("id")
	t, err := ctl.taskSvc.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch task"})
		return
	}
	if t.ID.IsZero() {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	resp := models.TaskResponse{
		ID:          t.ID.Hex(),
		Title:       t.Title,
		Description: t.Description,
		DueDate:     t.DueDate,
		Status:      t.Status,
	}
	c.JSON(http.StatusOK, resp)
}

// CreateTask handles POST /tasks (admin only)
func (ctl *Controller) CreateTask(c *gin.Context) {
	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json (title required)"})
		return
	}
	created, err := ctl.taskSvc.CreateTask(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}
	c.JSON(http.StatusCreated, models.TaskResponse{
		ID:          created.ID.Hex(),
		Title:       created.Title,
		Description: created.Description,
		DueDate:     created.DueDate,
		Status:      created.Status,
	})
}

// UpdateTask handles PUT /tasks/:id (admin only)
func (ctl *Controller) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	updated, err := ctl.taskSvc.UpdateTask(id, input)
	if err != nil {
		if err.Error() == "no fields to update" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}
	if updated.ID.IsZero() {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, models.TaskResponse{
		ID:          updated.ID.Hex(),
		Title:       updated.Title,
		Description: updated.Description,
		DueDate:     updated.DueDate,
		Status:      updated.Status,
	})
}

// DeleteTask handles DELETE /tasks/:id (admin only)
func (ctl *Controller) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	ok, err := ctl.taskSvc.DeleteTask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}
