package model

import (
	"time"

	"gorm.io/gorm"
)

// ExecutionLog represents a single log line from a pipeline execution
type ExecutionLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BatchID   string    `gorm:"size:64;not null;index:idx_batch_log" json:"batch_id"`
	Stage     string    `gorm:"size:64;not null;index:idx_stage" json:"stage"` // source, build, deploy, custom
	LogLine   string    `gorm:"type:text;not null" json:"log_line"`
	LogLevel  string    `gorm:"size:16;not null" json:"log_level"` // info, warn, error, debug
	LineNo    int       `gorm:"not null" json:"line_no"`
	CreatedAt time.Time `json:"created_at"`
}

func (ExecutionLog) TableName() string {
	return "execution_logs"
}

func AutoMigrateExecutionLog(db *gorm.DB) error {
	return db.AutoMigrate(&ExecutionLog{})
}
