package model

import "gorm.io/gorm"

type System struct {
	ID          string `gorm:"primaryKey;type:VARCHAR(36)"`
	UserID      uint   `gorm:"not null;index:idx_user_id"`
	Name        string `gorm:"not null;size:255;index:idx_user_system"`
	Description string `gorm:"type:TEXT"`
	Status      string `gorm:"size:50;default:'active'"` // planning, active, archived
	CreatedAt   int64 `gorm:"autoCreateTime:milli"`
	UpdatedAt   int64 `gorm:"autoUpdateTime:milli"`
}

type Plan struct {
	ID        string `gorm:"primaryKey;type:VARCHAR(36)"`
	SystemID  string `gorm:"not null;index:idx_system_id"`
	Version   string `gorm:"not null;size:50"`                    // e.g., "1.1.0"
	Status    string `gorm:"size:50;default:'planning'"` // planning, developing, released
	PlannedDate string `gorm:"size:50"`                   // date string, e.g., "2026-03-20"
	Description string `gorm:"type:TEXT"`
	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli"`
}

type PipelineInfo struct {
	ID          string `gorm:"primaryKey;type:VARCHAR(36)"`
	SystemID    string `gorm:"not null;index:idx_system_id"`
	PlanID      string `gorm:"index:idx_plan_id"` // nullable
	Name        string `gorm:"not null;size:255"`
	AppType     string `gorm:"not null;size:50"` // java, node, sql
	Description string `gorm:"type:TEXT"`
	CreatedAt   int64 `gorm:"autoCreateTime:milli"`
	UpdatedAt   int64 `gorm:"autoUpdateTime:milli"`
}

// AutoMigrateSystem migrates System, Plan, and PipelineInfo tables
func AutoMigrateSystem(db *gorm.DB) error {
	return db.AutoMigrate(&System{}, &Plan{}, &PipelineInfo{})
}
