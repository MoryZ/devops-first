package model

import "gorm.io/gorm"

// BPMDefinition stores a pipeline's BPM graph JSON.
type BPMDefinition struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	UserID         uint   `gorm:"not null;index:idx_user_pipeline_bpm,priority:1" json:"user_id"`
	PipelineID     string `gorm:"size:128;not null;index:idx_user_pipeline_bpm,priority:2" json:"pipeline_id"`
	DefinitionJSON string `gorm:"type:longtext;not null" json:"definition_json"`
}

func (BPMDefinition) TableName() string {
	return "bpm_definitions"
}

func AutoMigrateBPMDefinitions(db *gorm.DB) error {
	return db.AutoMigrate(&BPMDefinition{})
}
