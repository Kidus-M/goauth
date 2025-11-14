package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Task represents the DB model for a task
type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	Title       string             `bson:"title" json:"title" binding:"required"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	DueDate     string             `bson:"due_date,omitempty" json:"due_date,omitempty"`
	Status      string             `bson:"status,omitempty" json:"status,omitempty"`
}

// TaskResponse is used for API responses (ID as hex string)
type TaskResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Status      string `json:"status,omitempty"`
}
