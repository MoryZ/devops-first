package model

import (
	"time"

	"gorm.io/gorm"
)

// ExecutionBatch represents a single pipeline execution record
type ExecutionBatch struct {
	ID            string    `gorm:"primaryKey;size:64" json:"id"`
	UserID        uint      `gorm:"not null;index:idx_user_batch" json:"user_id"`
	SystemID      string    `gorm:"size:128;not null;index:idx_system_batch" json:"system_id"`
	PipelineID    string    `gorm:"size:128;not null;index:idx_pipeline_batch" json:"pipeline_id"`
	PipelineName  string    `gorm:"size:255;not null" json:"pipeline_name"`
	BatchNumber   int       `gorm:"not null" json:"batch_number"`
	Status        string    `gorm:"size:32;not null;index:idx_status" json:"status"` // pending, running, success, failed, cancelled
	TriggeredBy   string    `gorm:"size:32;not null" json:"triggered_by"`            // manual, webhook, schedule
	CommitID         string `gorm:"size:128" json:"commit_id"`                       // git commit hash recorded at execution time
	StagesStatusJSON string `gorm:"type:text" json:"stages_status_json"`              // JSON map[stageName]status: pending/running/success/failed
	WorkDir          string `gorm:"size:1024" json:"work_dir"`                       // /tmp/devops-exec/{systemId}/{pipelineId}/{batchId}
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	TotalDuration int64      `gorm:"default:0" json:"total_duration"` // milliseconds
	ErrorMessage  string     `gorm:"type:text" json:"error_message"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ExecutionBatch) TableName() string {
	return "execution_batches"
}

func AutoMigrateExecutionBatch(db *gorm.DB) error {
	return db.AutoMigrate(&ExecutionBatch{})
}
