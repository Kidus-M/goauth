package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// User stored in mongodb
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	Username     string             `bson:"username" json:"username" binding:"required"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	Role         string             `bson:"role" json:"role"` // "admin" or "user"
}
