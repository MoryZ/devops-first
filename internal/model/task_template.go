package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// TaskTemplate represents a pre-defined task template for pipeline stages
type TaskTemplate struct {
	ID               string          `gorm:"type:varchar(64);primaryKey" json:"id"`
	UserID           string          `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name             string          `gorm:"type:varchar(255);not null" json:"name"`       // e.g., "Java构建"
	Category         string          `gorm:"type:varchar(100);not null;index" json:"category"` // e.g., "构建"
	SubCategory      string          `gorm:"type:varchar(100)" json:"sub_category"`        // e.g., "Maven"
	Description      string          `gorm:"type:text" json:"description"`
	PresetFields     json.RawMessage `gorm:"type:json" json:"preset_fields"`
	AdvancedSettings json.RawMessage `gorm:"type:json" json:"advanced_settings"`
	Steps            []TaskTemplateStep  `gorm:"foreignKey:TemplateID" json:"steps"`
	Plugins          []TaskTemplatePlugin `gorm:"foreignKey:TemplateID" json:"plugins"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// TaskTemplateStep represents a single step in a task template
type TaskTemplateStep struct {
	ID             string          `gorm:"type:varchar(64);primaryKey" json:"id"`
	TemplateID     string          `gorm:"type:varchar(64);not null;index" json:"template_id"`
	StepOrder      int             `json:"step_order"`
	Name           string          `gorm:"type:varchar(255);not null" json:"name"`
	Command        string          `gorm:"type:longtext" json:"command"`
	ShellSpecified bool            `json:"shell_specified"`
	Envs           json.RawMessage `gorm:"type:json" json:"envs"`
	CreatedAt      time.Time       `json:"created_at"`
}

// TaskTemplatePlugin represents a plugin associated with a template
type TaskTemplatePlugin struct {
	ID           string          `gorm:"type:varchar(64);primaryKey" json:"id"`
	TemplateID   string          `gorm:"type:varchar(64);not null;index" json:"template_id"`
	PluginName   string          `gorm:"type:varchar(255);not null" json:"plugin_name"`
	PluginConfig json.RawMessage `gorm:"type:json" json:"plugin_config"`
	CreatedAt    time.Time       `json:"created_at"`
}

// TableName specifies the table name for TaskTemplate
func (TaskTemplate) TableName() string {
	return "task_templates"
}

// TableName specifies the table name for TaskTemplateStep
func (TaskTemplateStep) TableName() string {
	return "task_template_steps"
}

// TableName specifies the table name for TaskTemplatePlugin
func (TaskTemplatePlugin) TableName() string {
	return "task_template_plugins"
}

// AutoMigrateTaskTemplates creates task template tables
func AutoMigrateTaskTemplates(db *gorm.DB) error {
	return db.AutoMigrate(&TaskTemplate{}, &TaskTemplateStep{}, &TaskTemplatePlugin{})
}
