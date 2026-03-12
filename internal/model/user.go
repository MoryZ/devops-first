package model

import "gorm.io/gorm"

// User represents a user in the system
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"size:255;uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"` // Don't expose password
	Email    string `json:"email"`
	Remark   string `json:"remark"`
	Active   bool   `gorm:"default:true" json:"active"`
}

// TableName specifies the table name
func (User) TableName() string {
	return "users"
}

// AutoMigrate creates the users table if it doesn't exist
func AutoMigrateUsers(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
