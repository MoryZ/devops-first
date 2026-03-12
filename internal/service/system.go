package service

import (
	"fmt"
	"time"

	"devops-first/internal/database"
	"devops-first/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateSystem creates a new system for the user
func CreateSystem(userID uint, name, description string) (*model.System, error) {
	system := &model.System{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        name,
		Description: description,
		Status:      "active",
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
	}
	if err := database.GetDB().Create(system).Error; err != nil {
		return nil, fmt.Errorf("create system failed: %w", err)
	}
	return system, nil
}

// ListSystems lists all systems for a user
func ListSystems(userID uint) ([]model.System, error) {
	var systems []model.System
	if err := database.GetDB().Where("user_id = ?", userID).Order("updated_at DESC").Find(&systems).Error; err != nil {
		return nil, fmt.Errorf("list systems failed: %w", err)
	}
	return systems, nil
}

// GetSystem gets a system by ID (checks ownership)
func GetSystem(userID uint, systemID string) (*model.System, error) {
	var system model.System
	if err := database.GetDB().Where("id = ? AND user_id = ?", systemID, userID).First(&system).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("system not found")
		}
		return nil, fmt.Errorf("get system failed: %w", err)
	}
	return &system, nil
}

// CreatePlan creates a new plan under a system
func CreatePlan(systemID, version, status, plannedDate, description string) (*model.Plan, error) {
	plan := &model.Plan{
		ID:          uuid.New().String(),
		SystemID:    systemID,
		Version:     version,
		Status:      status,
		PlannedDate: plannedDate,
		Description: description,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
	}
	if err := database.GetDB().Create(plan).Error; err != nil {
		return nil, fmt.Errorf("create plan failed: %w", err)
	}
	return plan, nil
}

// ListPlans lists all plans under a system
func ListPlans(systemID string) ([]model.Plan, error) {
	var plans []model.Plan
	if err := database.GetDB().Where("system_id = ?", systemID).Order("version DESC").Find(&plans).Error; err != nil {
		return nil, fmt.Errorf("list plans failed: %w", err)
	}
	return plans, nil
}

// CreatePipeline creates a new pipeline under a system (and optionally a plan)
func CreatePipeline(systemID, planID, name, appType, description string) (*model.PipelineInfo, error) {
	pipeline := &model.PipelineInfo{
		ID:          uuid.New().String(),
		SystemID:    systemID,
		PlanID:      planID, // can be empty
		Name:        name,
		AppType:     appType, // java, node, sql
		Description: description,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
	}
	if err := database.GetDB().Create(pipeline).Error; err != nil {
		return nil, fmt.Errorf("create pipeline failed: %w", err)
	}
	return pipeline, nil
}

// ListPipelinesForPlan lists all pipelines under a specific plan
func ListPipelinesForPlan(planID string) ([]model.PipelineInfo, error) {
	var pipelines []model.PipelineInfo
	if err := database.GetDB().Where("plan_id = ?", planID).Order("created_at DESC").Find(&pipelines).Error; err != nil {
		return nil, fmt.Errorf("list pipelines for plan failed: %w", err)
	}
	return pipelines, nil
}

// ListPipelinesForSystem lists all pipelines under a system (across all plans)
func ListPipelinesForSystem(systemID string) ([]model.PipelineInfo, error) {
	var pipelines []model.PipelineInfo
	if err := database.GetDB().Where("system_id = ?", systemID).Order("created_at DESC").Find(&pipelines).Error; err != nil {
		return nil, fmt.Errorf("list pipelines for system failed: %w", err)
	}
	return pipelines, nil
}

// GetPipeline gets a pipeline by ID
func GetPipeline(pipelineID string) (*model.PipelineInfo, error) {
	var pipeline model.PipelineInfo
	if err := database.GetDB().Where("id = ?", pipelineID).First(&pipeline).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pipeline not found")
		}
		return nil, fmt.Errorf("get pipeline failed: %w", err)
	}
	return &pipeline, nil
}
