package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/AkashKesav/API2SDK/internal/models" // Corrected import path
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) (primitive.ObjectID, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByName(ctx context.Context, name string) (*models.User, error) // Changed from FindByUsername to FindByName
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error                     // Added Delete method
	FindAll(ctx context.Context, page, limit int) ([]*models.User, int64, error) // Added FindAll for admin
	CountAll(ctx context.Context) (int64, error)                                 // Added CountAll for stats
}

type mongoUserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *mongo.Database) UserRepository {
	return &mongoUserRepository{
		collection: db.Collection("users"),
	}
}

// Create inserts a new user into the database.
func (r *mongoUserRepository) Create(ctx context.Context, user *models.User) (primitive.ObjectID, error) {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return user.ID, nil
}

// FindByEmail retrieves a user by their email address.
func (r *mongoUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}

// FindByName retrieves a user by their name.
func (r *mongoUserRepository) FindByName(ctx context.Context, name string) (*models.User, error) { // Changed from FindByUsername to FindByName
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&user) // Changed from username to name
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}

// FindByID retrieves a user by their ID.
func (r *mongoUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}

// Update modifies an existing user's details in the database.
func (r *mongoUserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	// Ensure ID is not in the $set to avoid trying to change it, if it's part of the user struct passed in.
	// However, standard practice is that user.ID is used in the filter (UpdateByID) and not in the $set payload.
	// The current implementation of passing the whole 'user' object to $set is generally fine
	// as long as _id is immutable and handled correctly by MongoDB driver or conventions.
	// For clarity, one might explicitly build the update document:
	// updateFields := bson.M{
	// 	"name": user.Name,
	// 	"email": user.Email,
	// 	"password": user.Password, // if password changes are handled here
	// 	"updatedAt": user.UpdatedAt,
	// }
	// update := bson.M{
	// 	"$set": updateFields,
	// }
	update := bson.M{
		"$set": user,
	}
	_, err := r.collection.UpdateByID(ctx, user.ID, update)
	return err
}

// Delete removes a user from the database by their ID.
func (r *mongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("user not found or already deleted") // Or mongo.ErrNoDocuments if preferred
	}
	return nil
}

// FindAll retrieves all users with pagination.
func (r *mongoUserRepository) FindAll(ctx context.Context, page, limit int) ([]*models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10 // Default limit
	}
	skip := (page - 1) * limit

	var users []*models.User
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	totalCount, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
}

// CountAll counts all documents in the users collection.
func (r *mongoUserRepository) CountAll(ctx context.Context) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return count, nil
}
