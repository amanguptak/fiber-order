package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TokenHash string    `gorm:"not null;index"`
	IsRevoked bool      `gorm:"default:false"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}

func (token *RefreshToken) BeforeCreate(tx *gorm.DB) (err error) {
	token.ID = uuid.New()
	return
}

// index is a Database Performance Booster.

// Why use it here? We will frequently ask the database: "Hey, find all Refresh Tokens belonging to User X" (e.g., when logging out or revoking access).

// Without Index: The database has to scan EVERY row in the table (Slow 🐢).
// With Index: The database goes straight to User X's tokens (Fast ⚡).
// Why didn't we do it for Order? We probably should have! Usually, you add index to any column you use in a WHERE clause (like WHERE user_id = ?). It's a best practice for Foreign Keys
