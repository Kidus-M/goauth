package data

import (
	"context"
	"errors"
	"time"

	"go_mango/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskService works with a Mongo collection to perform CRUD
type TaskService struct {
	collection *mongo.Collection
	timeout    time.Duration
}

// NewTaskService returns a new TaskService bound to the provided collection
func NewTaskService(collection *mongo.Collection) *TaskService {
	return &TaskService{
		collection: collection,
		timeout:    5 * time.Second,
	}
}

// GetAllTasks returns all tasks stored in the collection
func (s *TaskService) GetAllTasks() ([]models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	cur, err := s.collection.Find(ctx, bson.D{}, options.Find())
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var tasks []models.Task
	for cur.Next(ctx) {
		var t models.Task
		if err := cur.Decode(&t); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTaskByID returns a task by its hex string ID
func (s *TaskService) GetTaskByID(hexID string) (models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(hexID)
	if err != nil {
		return models.Task{}, errors.New("invalid id")
	}

	var task models.Task
	if err := s.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&task); err != nil {
		if err == mongo.ErrNoDocuments {
			return models.Task{}, nil
		}
		return models.Task{}, err
	}
	return task, nil
}

// CreateTask inserts a new task and returns the inserted document (with ID)
func (s *TaskService) CreateTask(input models.Task) (models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// Ensure Title is present (binding also checks this at controller level)
	if input.Title == "" {
		return models.Task{}, errors.New("title required")
	}

	res, err := s.collection.InsertOne(ctx, input)
	if err != nil {
		return models.Task{}, err
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return models.Task{}, errors.New("failed to convert inserted id")
	}
	input.ID = oid
	return input, nil
}

// UpdateTask updates existing fields of a task by hex ID. Returns updated task.
func (s *TaskService) UpdateTask(hexID string, updated models.Task) (models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(hexID)
	if err != nil {
		return models.Task{}, errors.New("invalid id")
	}

	updateDoc := bson.M{}
	if updated.Title != "" {
		updateDoc["title"] = updated.Title
	}
	if updated.Description != "" {
		updateDoc["description"] = updated.Description
	}
	if updated.DueDate != "" {
		updateDoc["due_date"] = updated.DueDate
	}
	if updated.Status != "" {
		updateDoc["status"] = updated.Status
	}
	if len(updateDoc) == 0 {
		return models.Task{}, errors.New("no fields to update")
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var result models.Task
	if err := s.collection.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": updateDoc}, opts).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return models.Task{}, nil
		}
		return models.Task{}, err
	}
	return result, nil
}

// DeleteTask deletes a task by hex ID. Returns true if deleted.
func (s *TaskService) DeleteTask(hexID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(hexID)
	if err != nil {
		return false, errors.New("invalid id")
	}

	res, err := s.collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}
