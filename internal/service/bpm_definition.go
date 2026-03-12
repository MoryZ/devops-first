package service

import (
	"encoding/json"
	"fmt"

	"devops-first/internal/database"
	"devops-first/internal/model"
	"gorm.io/gorm"
)

type BPMDefinitionRequest struct {
	PipelineID string                 `json:"pipeline_id" binding:"required"`
	Definition map[string]interface{} `json:"definition" binding:"required"`
}

type BPMDefinitionResponse struct {
	PipelineID string                 `json:"pipeline_id"`
	Definition map[string]interface{} `json:"definition"`
}

type BPMDefinitionService struct{}

func NewBPMDefinitionService() *BPMDefinitionService {
	return &BPMDefinitionService{}
}

func defaultBPMDefinition() map[string]interface{} {
	return map[string]interface{}{
		"nodes": []map[string]interface{}{
			{"id": "start_1", "type": "start", "name": "开始", "x": 80, "y": 120},
			{"id": "end_1", "type": "end", "name": "结束", "x": 520, "y": 120},
		},
		"edges": []map[string]interface{}{},
	}
}

func (s *BPMDefinitionService) Upsert(userID uint, req BPMDefinitionRequest) error {
	body, err := json.Marshal(req.Definition)
	if err != nil {
		return fmt.Errorf("marshal definition: %w", err)
	}

	db := database.GetDB()
	var existing model.BPMDefinition
	err = db.Where("user_id = ? AND pipeline_id = ?", userID, req.PipelineID).First(&existing).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("query bpm definition: %w", err)
	}

	entity := model.BPMDefinition{
		UserID:         userID,
		PipelineID:     req.PipelineID,
		DefinitionJSON: string(body),
	}

	if existing.ID == 0 {
		if err := db.Create(&entity).Error; err != nil {
			return fmt.Errorf("create bpm definition: %w", err)
		}
		return nil
	}

	if err := db.Model(&existing).Updates(entity).Error; err != nil {
		return fmt.Errorf("update bpm definition: %w", err)
	}
	return nil
}

func (s *BPMDefinitionService) Get(userID uint, pipelineID string) (BPMDefinitionResponse, error) {
	db := database.GetDB()
	var row model.BPMDefinition
	err := db.Where("user_id = ? AND pipeline_id = ?", userID, pipelineID).First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return BPMDefinitionResponse{PipelineID: pipelineID, Definition: defaultBPMDefinition()}, nil
		}
		return BPMDefinitionResponse{}, fmt.Errorf("query bpm definition: %w", err)
	}

	var definition map[string]interface{}
	if err := json.Unmarshal([]byte(row.DefinitionJSON), &definition); err != nil {
		return BPMDefinitionResponse{}, fmt.Errorf("unmarshal definition: %w", err)
	}
	if definition == nil {
		definition = defaultBPMDefinition()
	}

	return BPMDefinitionResponse{PipelineID: pipelineID, Definition: definition}, nil
}
