package model

import "gorm.io/gorm"

// GlobalVariable stores user-scoped global variables and secrets referenced by pipeline nodes.
type GlobalVariable struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	UserID      uint   `gorm:"not null;index:idx_user_key,priority:1" json:"user_id"`
	Key         string `gorm:"size:255;not null;index:idx_user_key,priority:2,unique" json:"key"`
	Value       string `gorm:"type:text;not null" json:"value,omitempty"`
	IsSecret    bool   `gorm:"not null;default:false" json:"is_secret"`
	Description string `gorm:"size:512" json:"description"`
}

func (GlobalVariable) TableName() string {
	return "global_variables"
}

func AutoMigrateGlobalVariables(db *gorm.DB) error {
	return db.AutoMigrate(&GlobalVariable{})
}
