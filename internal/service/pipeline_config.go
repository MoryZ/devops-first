package service

import (
	"encoding/json"
	"fmt"

	"devops-first/internal/database"
	"devops-first/internal/model"
	"gorm.io/gorm"
)

type StageItem struct {
	Name     string `json:"name"`
	Duration string `json:"duration"`
	Status   string `json:"status"`
}

type PipelineConfigRequest struct {
	PipelineID       string      `json:"pipeline_id" binding:"required"`
	Name             string      `json:"name" binding:"required"`
	ReleaseUnitID    string      `json:"release_unit_id"`
	RepositoryType   string      `json:"repository_type"`
	AutoMerge        *bool       `json:"auto_merge"`
	AutoTag          *bool       `json:"auto_tag"`
	DisplayOrder     int         `json:"display_order"`
	RepoURL          string      `json:"repo_url"`
	Branch           string      `json:"branch"`
	GitUsername      string      `json:"git_username"`
	GitToken         string      `json:"git_token"`
	GitCredentialKey string      `json:"git_credential_key"`
	GitUsernameField string      `json:"git_username_field"`
	GitTokenField    string      `json:"git_token_field"`
	ProjectPath      string      `json:"project_path"`
	BuildType        string      `json:"build_type"`
	MavenCommand     string      `json:"maven_command"`
	GradleCommand    string      `json:"gradle_command"`
	NPMCommand       string      `json:"npm_command"`
	DeployType       string      `json:"deploy_type"`
	DockerImage      string      `json:"docker_image"`
	DockerContainer  string      `json:"docker_container"`
	DockerRunArgs    string      `json:"docker_run_args"`
	MainStages       []StageItem `json:"main_stages"`
	EnvStages        []StageItem `json:"env_stages"`
}

type PipelineConfigResponse struct {
	PipelineID       string      `json:"pipeline_id"`
	Name             string      `json:"name"`
	ReleaseUnitID    string      `json:"release_unit_id"`
	RepositoryType   string      `json:"repository_type"`
	AutoMerge        bool        `json:"auto_merge"`
	AutoTag          bool        `json:"auto_tag"`
	DisplayOrder     int         `json:"display_order"`
	RepoURL          string      `json:"repo_url"`
	Branch           string      `json:"branch"`
	GitUsername      string      `json:"git_username"`
	GitToken         string      `json:"git_token"`
	GitCredentialKey string      `json:"git_credential_key"`
	GitUsernameField string      `json:"git_username_field"`
	GitTokenField    string      `json:"git_token_field"`
	ProjectPath      string      `json:"project_path"`
	BuildType        string      `json:"build_type"`
	MavenCommand     string      `json:"maven_command"`
	GradleCommand    string      `json:"gradle_command"`
	NPMCommand       string      `json:"npm_command"`
	DeployType       string      `json:"deploy_type"`
	DockerImage      string      `json:"docker_image"`
	DockerContainer  string      `json:"docker_container"`
	DockerRunArgs    string      `json:"docker_run_args"`
	MainStages       []StageItem `json:"main_stages"`
	EnvStages        []StageItem `json:"env_stages"`
}

type PipelineConfigService struct{}

func NewPipelineConfigService() *PipelineConfigService {
	return &PipelineConfigService{}
}

func (s *PipelineConfigService) Upsert(userID uint, req PipelineConfigRequest) error {
	autoMerge := true
	if req.AutoMerge != nil {
		autoMerge = *req.AutoMerge
	}
	autoTag := true
	if req.AutoTag != nil {
		autoTag = *req.AutoTag
	}

	buildType := req.BuildType
	if buildType == "" {
		buildType = "maven"
	}
	deployType := req.DeployType
	if deployType == "" {
		deployType = "docker"
	}

	mainBytes, err := json.Marshal(req.MainStages)
	if err != nil {
		return fmt.Errorf("marshal main stages: %w", err)
	}
	envBytes, err := json.Marshal(req.EnvStages)
	if err != nil {
		return fmt.Errorf("marshal env stages: %w", err)
	}

	db := database.GetDB()
	var existing model.PipelineConfig
	err = db.Where("user_id = ? AND pipeline_id = ?", userID, req.PipelineID).First(&existing).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("query pipeline config: %w", err)
		}
	}

	entity := model.PipelineConfig{
		UserID:           userID,
		PipelineID:       req.PipelineID,
		Name:             req.Name,
		ReleaseUnitID:    req.ReleaseUnitID,
		RepositoryType:   req.RepositoryType,
		AutoMerge:        autoMerge,
		AutoTag:          autoTag,
		DisplayOrder:     req.DisplayOrder,
		RepoURL:          req.RepoURL,
		Branch:           req.Branch,
		GitUsername:      req.GitUsername,
		GitToken:         req.GitToken,
		GitCredentialKey: req.GitCredentialKey,
		GitUsernameField: req.GitUsernameField,
		GitTokenField:    req.GitTokenField,
		ProjectPath:      req.ProjectPath,
		BuildType:        buildType,
		MavenCommand:     req.MavenCommand,
		GradleCommand:    req.GradleCommand,
		NPMCommand:       req.NPMCommand,
		DeployType:       deployType,
		DockerImage:      req.DockerImage,
		DockerContainer:  req.DockerContainer,
		DockerRunArgs:    req.DockerRunArgs,
		MainStagesJSON:   string(mainBytes),
		EnvStagesJSON:    string(envBytes),
	}

	if existing.ID == 0 {
		if err := db.Create(&entity).Error; err != nil {
			return fmt.Errorf("create pipeline config: %w", err)
		}
		return nil
	}

	if err := db.Model(&existing).Updates(entity).Error; err != nil {
		return fmt.Errorf("update pipeline config: %w", err)
	}
	return nil
}

func (s *PipelineConfigService) UpsertReleaseUnitBinding(userID uint, pipelineID, releaseUnitID string) error {
	if pipelineID == "" {
		return fmt.Errorf("pipeline_id is required")
	}

	db := database.GetDB()
	var existing model.PipelineConfig
	err := db.Where("user_id = ? AND pipeline_id = ?", userID, pipelineID).First(&existing).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("query pipeline config: %w", err)
		}

		name := pipelineID
		var pipeline model.PipelineInfo
		if err := db.Where("id = ?", pipelineID).First(&pipeline).Error; err == nil && pipeline.Name != "" {
			name = pipeline.Name
		}

		entity := model.PipelineConfig{
			UserID:        userID,
			PipelineID:    pipelineID,
			Name:          name,
			ReleaseUnitID: releaseUnitID,
		}
		if err := db.Create(&entity).Error; err != nil {
			return fmt.Errorf("create pipeline config: %w", err)
		}
		return nil
	}

	if err := db.Model(&existing).Updates(map[string]interface{}{
		"release_unit_id": releaseUnitID,
	}).Error; err != nil {
		return fmt.Errorf("update pipeline release unit binding: %w", err)
	}

	return nil
}

func (s *PipelineConfigService) List(userID uint) ([]PipelineConfigResponse, error) {
	db := database.GetDB()
	var rows []model.PipelineConfig
	if err := db.Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query pipeline configs: %w", err)
	}

	res := make([]PipelineConfigResponse, 0, len(rows))
	for _, row := range rows {
		item := PipelineConfigResponse{
			PipelineID:       row.PipelineID,
			Name:             row.Name,
			ReleaseUnitID:    row.ReleaseUnitID,
			RepositoryType:   row.RepositoryType,
			AutoMerge:        row.AutoMerge,
			AutoTag:          row.AutoTag,
			DisplayOrder:     row.DisplayOrder,
			RepoURL:          row.RepoURL,
			Branch:           row.Branch,
			GitUsername:      row.GitUsername,
			GitToken:         row.GitToken,
			GitCredentialKey: row.GitCredentialKey,
			GitUsernameField: row.GitUsernameField,
			GitTokenField:    row.GitTokenField,
			ProjectPath:      row.ProjectPath,
			BuildType:        row.BuildType,
			MavenCommand:     row.MavenCommand,
			GradleCommand:    row.GradleCommand,
			NPMCommand:       row.NPMCommand,
			DeployType:       row.DeployType,
			DockerImage:      row.DockerImage,
			DockerContainer:  row.DockerContainer,
			DockerRunArgs:    row.DockerRunArgs,
		}
		if row.MainStagesJSON != "" {
			_ = json.Unmarshal([]byte(row.MainStagesJSON), &item.MainStages)
		}
		if row.EnvStagesJSON != "" {
			_ = json.Unmarshal([]byte(row.EnvStagesJSON), &item.EnvStages)
		}
		res = append(res, item)
	}
	return res, nil
}

func (s *PipelineConfigService) Get(userID uint, pipelineID string) (*PipelineConfigResponse, error) {
	if pipelineID == "" {
		return nil, fmt.Errorf("pipeline_id is required")
	}

	db := database.GetDB()
	var row model.PipelineConfig
	if err := db.Where("user_id = ? AND pipeline_id = ?", userID, pipelineID).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("pipeline config not found")
		}
		return nil, fmt.Errorf("query pipeline config: %w", err)
	}

	item := &PipelineConfigResponse{
		PipelineID:       row.PipelineID,
		Name:             row.Name,
		ReleaseUnitID:    row.ReleaseUnitID,
		RepositoryType:   row.RepositoryType,
		AutoMerge:        row.AutoMerge,
		AutoTag:          row.AutoTag,
		DisplayOrder:     row.DisplayOrder,
		RepoURL:          row.RepoURL,
		Branch:           row.Branch,
		GitUsername:      row.GitUsername,
		GitToken:         row.GitToken,
		GitCredentialKey: row.GitCredentialKey,
		GitUsernameField: row.GitUsernameField,
		GitTokenField:    row.GitTokenField,
		ProjectPath:      row.ProjectPath,
		BuildType:        row.BuildType,
		MavenCommand:     row.MavenCommand,
		GradleCommand:    row.GradleCommand,
		NPMCommand:       row.NPMCommand,
		DeployType:       row.DeployType,
		DockerImage:      row.DockerImage,
		DockerContainer:  row.DockerContainer,
		DockerRunArgs:    row.DockerRunArgs,
	}
	if row.MainStagesJSON != "" {
		_ = json.Unmarshal([]byte(row.MainStagesJSON), &item.MainStages)
	}
	if row.EnvStagesJSON != "" {
		_ = json.Unmarshal([]byte(row.EnvStagesJSON), &item.EnvStages)
	}
	return item, nil
}


