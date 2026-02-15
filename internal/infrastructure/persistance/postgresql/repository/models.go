package repo

import (
	"time"
)

type User struct {
	ID        string    `gorm:"type:char(26);primaryKey"`
	Email     string    `gorm:"type:varchar(256);unique;not null"`
	Password  string    `gorm:"type:char(256);not null"`
	FirstName string    `gorm:"type:varchar(256)"`
	LastName  string    `gorm:"type:varchar(256)"`
	Nickname  string    `gorm:"type:varchar(256);not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

type UserVerificationSession struct {
	ID        string `gorm:"type:char(26);primaryKey"`
	Email     string `gorm:"type:varchar(256);unique;not null"`
	Password  string `gorm:"type:char(256);not null"`
	FirstName string `gorm:"type:varchar(256)"`
	LastName  string `gorm:"type:varchar(256)"`
	Nickname  string `gorm:"type:varchar(256);not null"`
}

type UserPasswordChangingSession struct {
	ID        string    `gorm:"type:char(26);primaryKey"`
	UserID    string    `gorm:"type:char(26);not null"`
	ExpiresAt time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type UserJWTSession struct {
	UserID         string    `gorm:"type:char(26);not null;index:idx_session_user_id"`
	RefreshTokenID string    `gorm:"type:char(26);primaryKey;not null"`
	IsRevoked      bool      `gorm:"not null"`
	ExpiresAt      time.Time `gorm:"not null;index:idx_session_expiry,where:is_revoked = false"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
