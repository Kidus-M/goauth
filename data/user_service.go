package data

import (
	"context"
	"errors"
	"time"

	"task_manager/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserService manages users in MongoDB
type UserService struct {
	collection *mongo.Collection
	timeout    time.Duration
}

// NewUserService constructs a UserService
func NewUserService(coll *mongo.Collection) *UserService {
	// ensure unique username index
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return &UserService{collection: coll, timeout: 5 * time.Second}
}

// CreateUser hashes password and creates user. If DB empty -> first user becomes admin.
func (s *UserService) CreateUser(username, password string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if username == "" || password == "" {
		return models.User{}, errors.New("username and password required")
	}
	// check existing
	count, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return models.User{}, err
	}

	// hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	role := "user"
	if count == 0 {
		role = "admin" // first user is admin
	}

	u := models.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	}

	res, err := s.collection.InsertOne(ctx, u)
	if err != nil {
		// duplicate user will error because index created
		if mongo.IsDuplicateKeyError(err) {
			return models.User{}, errors.New("username already exists")
		}
		return models.User{}, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		u.ID = oid
	}
	u.PasswordHash = "" // don't return hash
	return u, nil
}

// Authenticate validates username/password and returns user if ok
func (s *UserService) Authenticate(username, password string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	var u models.User
	if err := s.collection.FindOne(ctx, bson.M{"username": username}).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, errors.New("invalid credentials")
		}
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return models.User{}, errors.New("invalid credentials")
	}

	// clear hash for returning
	u.PasswordHash = ""
	return u, nil
}

// FindByUsername returns user (without password hash)
func (s *UserService) FindByUsername(username string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	var u models.User
	if err := s.collection.FindOne(ctx, bson.M{"username": username}).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, nil
		}
		return models.User{}, err
	}
	u.PasswordHash = ""
	return u, nil
}

// PromoteUser sets role to admin; returns updated user
func (s *UserService) PromoteUser(username string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	update := bson.M{"$set": bson.M{"role": "admin"}}
	var updated models.User
	if err := s.collection.FindOneAndUpdate(ctx, bson.M{"username": username}, update, opts).Decode(&updated); err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, nil
		}
		return models.User{}, err
	}
	updated.PasswordHash = ""
	return updated, nil
}

// IsEmpty checks whether users collection is empty
func (s *UserService) IsEmpty() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	count, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// GetByID returns user by hex id (if needed)
func (s *UserService) GetByID(hexID string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(hexID)
	if err != nil {
		return models.User{}, errors.New("invalid id")
	}
	var u models.User
	if err := s.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, nil
		}
		return models.User{}, err
	}
	u.PasswordHash = ""
	return u, nil
}
