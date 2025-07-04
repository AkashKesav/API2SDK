package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ID represents a standardized ID type for the application
type ID string

// ObjectID converts the ID to a MongoDB ObjectID
func (id ID) ObjectID() primitive.ObjectID {
	if id == "" {
		return primitive.NilObjectID
	}
	objID, err := primitive.ObjectIDFromHex(string(id))
	if err != nil {
		return primitive.NilObjectID
	}
	return objID
}

// String returns the string representation of the ID
func (id ID) String() string {
	return string(id)
}

// IsZero checks if the ID is empty
func (id ID) IsZero() bool {
	return id == ""
}

// NewID creates a new ID from a MongoDB ObjectID
func NewID() ID {
	return ID(primitive.NewObjectID().Hex())
}

// NewIDFromObjectID creates an ID from an existing ObjectID
func NewIDFromObjectID(objID primitive.ObjectID) ID {
	return ID(objID.Hex())
}

// NewIDFromString creates an ID from a string
func NewIDFromString(s string) ID {
	return ID(s)
}

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        ID        `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
	Version   int64     `bson:"version" json:"version"` // For optimistic locking
}

// SetID sets the ID field
func (b *BaseModel) SetID(id ID) {
	b.ID = id
}

// GetID returns the ID field
func (b *BaseModel) GetID() ID {
	return b.ID
}

// SetTimestamps sets the created and updated timestamps
func (b *BaseModel) SetTimestamps() {
	now := time.Now()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
}

// IncrementVersion increments the version for optimistic locking
func (b *BaseModel) IncrementVersion() {
	b.Version++
}

// SoftDeletable defines the interface for soft-deletable models.
type SoftDeletable interface {
	SoftDelete()
	Restore()
	GetIsDeleted() bool
}

// SoftDeleteModel provides soft delete functionality
type SoftDeleteModel struct {
	BaseModel
	DeletedAt *time.Time `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
	IsDeleted bool       `bson:"isDeleted" json:"isDeleted"`
}

// SoftDelete marks the entity as deleted
func (s *SoftDeleteModel) SoftDelete() {
	now := time.Now()
	s.DeletedAt = &now
	s.IsDeleted = true
	s.UpdatedAt = now
}

// Restore restores a soft-deleted entity
func (s *SoftDeleteModel) Restore() {
	s.DeletedAt = nil
	s.IsDeleted = false
	s.UpdatedAt = time.Now()
}

// GetIsDeleted returns the soft delete status
func (s *SoftDeleteModel) GetIsDeleted() bool {
	return s.IsDeleted
}

// AuditableModel provides audit trail functionality
type AuditableModel struct {
	BaseModel
	CreatedBy ID `bson:"createdBy,omitempty" json:"createdBy,omitempty"`
	UpdatedBy ID `bson:"updatedBy,omitempty" json:"updatedBy,omitempty"`
}

// SetCreatedBy sets the creator ID
func (a *AuditableModel) SetCreatedBy(userID ID) {
	a.CreatedBy = userID
}

// SetUpdatedBy sets the updater ID
func (a *AuditableModel) SetUpdatedBy(userID ID) {
	a.UpdatedBy = userID
}
