package model

import "gorm.io/gorm"

// PipelineConfig stores per-user CI/CD pipeline settings.
type PipelineConfig struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	UserID          uint   `gorm:"not null;index:idx_user_pipeline,priority:1" json:"user_id"`
	PipelineID      string `gorm:"size:128;not null;index:idx_user_pipeline,priority:2" json:"pipeline_id"`
	Name            string `gorm:"size:255;not null" json:"name"`
	ReleaseUnitID   string `gorm:"size:128" json:"release_unit_id"`
	RepositoryType  string `gorm:"size:64" json:"repository_type"`
	AutoMerge       bool   `gorm:"default:true" json:"auto_merge"`
	AutoTag         bool   `gorm:"default:true" json:"auto_tag"`
	DisplayOrder    int    `gorm:"default:0" json:"display_order"`
	RepoURL         string `gorm:"size:1024" json:"repo_url"`
	Branch          string `gorm:"size:255" json:"branch"`
	GitUsername     string `gorm:"size:255" json:"git_username"`
	GitToken        string `gorm:"size:512" json:"git_token"`
	GitCredentialKey string `gorm:"size:255;default:'github'" json:"git_credential_key"`
	GitUsernameField string `gorm:"size:255;default:'username'" json:"git_username_field"`
	GitTokenField    string `gorm:"size:255;default:'token'" json:"git_token_field"`
	ProjectPath     string `gorm:"size:1024" json:"project_path"`
	BuildType       string `gorm:"size:64;default:'maven'" json:"build_type"` // maven, gradle, npm, none
	MavenCommand    string `gorm:"size:512" json:"maven_command"`
	GradleCommand   string `gorm:"size:512" json:"gradle_command"`
	NPMCommand      string `gorm:"size:512" json:"npm_command"`
	DeployType      string `gorm:"size:64;default:'docker'" json:"deploy_type"` // docker, jar, war
	DockerImage     string `gorm:"size:512" json:"docker_image"`
	DockerContainer string `gorm:"size:255" json:"docker_container"`
	DockerRunArgs   string `gorm:"size:1024" json:"docker_run_args"`
	MainStagesJSON  string `gorm:"type:text" json:"main_stages_json"`
	EnvStagesJSON   string `gorm:"type:text" json:"env_stages_json"`
}

func (PipelineConfig) TableName() string {
	return "pipeline_configs"
}

func AutoMigratePipelineConfigs(db *gorm.DB) error {
	return db.AutoMigrate(&PipelineConfig{})
}
