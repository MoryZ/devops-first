package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"devops-first/internal/model"
	"gorm.io/gorm"
)

type TaskTemplateService struct {
	db *gorm.DB
}

func NewTaskTemplateService(db *gorm.DB) *TaskTemplateService {
	return &TaskTemplateService{db: db}
}

// InitializeDefaultTemplates creates default task templates in the database
func (s *TaskTemplateService) InitializeDefaultTemplates(userID string) error {
	templates := s.getDefaultTemplates(userID)

	for _, template := range templates {
		existing := model.TaskTemplate{}
		if err := s.db.Where("user_id = ? AND name = ? AND category = ?", userID, template.Name, template.Category).
			First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := s.db.Create(&template).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// GetAllTemplates returns all templates for a user
func (s *TaskTemplateService) GetAllTemplates(userID string) ([]model.TaskTemplate, error) {
	var templates []model.TaskTemplate
	if err := s.db.
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC")
		}).
		Preload("Plugins").
		Where("user_id = ?", userID).
		Order("category, sub_category, name").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// GetTemplatesByCategory returns templates filtered by category
func (s *TaskTemplateService) GetTemplatesByCategory(userID, category string) ([]model.TaskTemplate, error) {
	var templates []model.TaskTemplate
	if err := s.db.
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC")
		}).
		Preload("Plugins").
		Where("user_id = ? AND category = ?", userID, category).
		Order("sub_category, name").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// GetTemplateByID returns a template by ID
func (s *TaskTemplateService) GetTemplateByID(id string) (*model.TaskTemplate, error) {
	template := &model.TaskTemplate{}
	if err := s.db.
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC")
		}).
		Preload("Plugins").
		First(template, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return template, nil
}

// CreateTemplate creates a new template
func (s *TaskTemplateService) CreateTemplate(template *model.TaskTemplate) error {
	template.ID = generateID()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	for i := range template.Steps {
		template.Steps[i].ID = generateID()
		template.Steps[i].TemplateID = template.ID
		template.Steps[i].CreatedAt = time.Now()
	}

	for i := range template.Plugins {
		template.Plugins[i].ID = generateID()
		template.Plugins[i].TemplateID = template.ID
		template.Plugins[i].CreatedAt = time.Now()
	}

	return s.db.Create(template).Error
}

// UpdateTemplate updates an existing template
func (s *TaskTemplateService) UpdateTemplate(template *model.TaskTemplate) error {
	template.UpdatedAt = time.Now()
	return s.db.Save(template).Error
}

// DeleteTemplate soft-deletes a template
func (s *TaskTemplateService) DeleteTemplate(id string) error {
	return s.db.Delete(&model.TaskTemplate{}, "id = ?", id).Error
}

func generateID() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))))
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

// getDefaultTemplates returns a simplified but structured template catalog.
func (s *TaskTemplateService) getDefaultTemplates(userID string) []model.TaskTemplate {
	return []model.TaskTemplate{
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "Git 拉取",
			Category:    "源码",
			SubCategory: "Git拉取",
			Description: "从 Git 仓库拉取代码，默认分支为 main，可在阶段配置中修改",
			PresetFields: mustMarshal(map[string]interface{}{
				"repoUrl":          "",
				"branch":           "main",
				"authType":         "none",
				"gitCredentialKey": "github",
				"gitUsernameField": "username",
				"gitTokenField":    "token",
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{}),
			Steps:            []model.TaskTemplateStep{},
			Plugins:          []model.TaskTemplatePlugin{},
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "Java 安全扫描",
			Category:    "代码扫描",
			SubCategory: "Java代码扫描",
			Description: "按 Java 项目常见流程执行安全扫描",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildCluster":      "云效默认构建集群",
				"buildNode":         "Linux/amd64",
				"buildEnvironment":  "container",
				"containerImage":    "build-steps/alinux3",
				"downloadSource":    "all",
				"taskOutputArtifact": true,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":     "DEFAULT",
				"timeoutMinutes": 240,
				"debugMode":      false,
				"deployTask":     false,
				"dockerDaemon":   false,
			}),
			Steps: []model.TaskTemplateStep{
				{ID: generateID(), StepOrder: 1, Name: "配置 MavenSettings 文件", Command: "echo use maven settings", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
				{ID: generateID(), StepOrder: 2, Name: "安装 Java", Command: "echo install java", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
				{ID: generateID(), StepOrder: 3, Name: "执行 Java 构建命令", Command: "mvn -B clean package -Dmaven.test.skip=true", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
				{ID: generateID(), StepOrder: 4, Name: "Java 安全扫描 Spotbugs", Command: "mvn com.github.spotbugs:spotbugs-maven-plugin:check", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
			},
			Plugins: []model.TaskTemplatePlugin{
				{ID: generateID(), PluginName: "SpotBugs", PluginConfig: mustMarshal(map[string]interface{}{"severity": "medium"}), CreatedAt: time.Now()},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "Python 代码扫描",
			Category:    "代码扫描",
			SubCategory: "Python代码扫描",
			Description: "使用 Bandit 扫描 Python 安全问题",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildCluster":      "云效默认构建集群",
				"buildNode":         "Linux/amd64",
				"buildEnvironment":  "container",
				"containerImage":    "python:3.11",
				"downloadSource":    "all",
				"taskOutputArtifact": false,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":      "DEFAULT",
				"timeoutMinutes": 180,
				"debugMode":      false,
				"deployTask":     false,
				"dockerDaemon":   false,
			}),
			Steps: []model.TaskTemplateStep{
				{ID: generateID(), StepOrder: 1, Name: "安装依赖", Command: "pip install -r requirements.txt", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
				{ID: generateID(), StepOrder: 2, Name: "执行 Bandit 扫描", Command: "bandit -r . -f txt", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
			},
			Plugins:   []model.TaskTemplatePlugin{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "单元测试",
			Category:    "测试",
			SubCategory: "通用测试",
			Description: "执行测试任务",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildEnvironment": "default",
				"timeoutMinutes":   120,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":      "DEFAULT",
				"timeoutMinutes": 120,
				"debugMode":      false,
				"deployTask":     false,
				"dockerDaemon":   false,
			}),
			Steps: []model.TaskTemplateStep{
				{ID: generateID(), StepOrder: 1, Name: "执行测试", Command: "echo run test command", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
			},
			Plugins:   []model.TaskTemplatePlugin{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "Java 构建",
			Category:    "构建",
			SubCategory: "Java构建",
			Description: "Maven 构建 Java 项目",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildEnvironment": "container",
				"containerImage":   "maven:3.9-eclipse-temurin-17",
				"timeoutMinutes":   240,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":      "DEFAULT",
				"timeoutMinutes": 240,
				"debugMode":      false,
				"deployTask":     false,
				"dockerDaemon":   false,
			}),
			Steps: []model.TaskTemplateStep{
				{ID: generateID(), StepOrder: 1, Name: "执行构建", Command: "mvn -B clean package -Dmaven.test.skip=true", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
			},
			Plugins:   []model.TaskTemplatePlugin{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "测试并构建",
			Category:    "测试构建",
			SubCategory: "Java测试构建",
			Description: "先测试再构建",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildEnvironment": "container",
				"containerImage":   "maven:3.9-eclipse-temurin-17",
				"timeoutMinutes":   300,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":      "DEFAULT",
				"timeoutMinutes": 300,
				"debugMode":      false,
				"deployTask":     false,
				"dockerDaemon":   false,
			}),
			Steps: []model.TaskTemplateStep{
				{ID: generateID(), StepOrder: 1, Name: "执行单元测试", Command: "mvn test", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
				{ID: generateID(), StepOrder: 2, Name: "执行构建", Command: "mvn -B clean package -Dmaven.test.skip=true", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
			},
			Plugins:   []model.TaskTemplatePlugin{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "镜像构建",
			Category:    "镜像构建",
			SubCategory: "Docker",
			Description: "构建Docker镜像",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildEnvironment": "container",
				"containerImage":   "docker:24",
				"timeoutMinutes":   300,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":      "DEFAULT",
				"timeoutMinutes": 300,
				"debugMode":      false,
				"deployTask":     false,
				"dockerDaemon":   true,
			}),
			Steps: []model.TaskTemplateStep{
				{ID: generateID(), StepOrder: 1, Name: "构建镜像", Command: "docker build -t ${IMAGE}:${TAG} .", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
				{ID: generateID(), StepOrder: 2, Name: "推送镜像", Command: "docker push ${IMAGE}:${TAG}", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
			},
			Plugins:   []model.TaskTemplatePlugin{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "应用部署",
			Category:    "部署",
			SubCategory: "Kubernetes部署",
			Description: "部署到 K8s 集群",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildEnvironment": "default",
				"timeoutMinutes":   240,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":      "DEFAULT",
				"timeoutMinutes": 240,
				"debugMode":      false,
				"deployTask":     true,
				"dockerDaemon":   false,
			}),
			Steps: []model.TaskTemplateStep{
				{ID: generateID(), StepOrder: 1, Name: "执行 kubectl 部署", Command: "kubectl apply -f k8s/", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
			},
			Plugins:   []model.TaskTemplatePlugin{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "工具任务",
			Category:    "工具",
			SubCategory: "通用工具",
			Description: "执行通用工具类命令",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildEnvironment": "default",
				"timeoutMinutes":   60,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":      "DEFAULT",
				"timeoutMinutes": 60,
				"debugMode":      false,
				"deployTask":     false,
				"dockerDaemon":   false,
			}),
			Steps: []model.TaskTemplateStep{
				{ID: generateID(), StepOrder: 1, Name: "执行工具命令", Command: "echo run tool command", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
			},
			Plugins:   []model.TaskTemplatePlugin{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "代码任务",
			Category:    "代码",
			SubCategory: "代码处理",
			Description: "执行代码格式化/检查等操作",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildEnvironment": "default",
				"timeoutMinutes":   90,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":      "DEFAULT",
				"timeoutMinutes": 90,
				"debugMode":      false,
				"deployTask":     false,
				"dockerDaemon":   false,
			}),
			Steps: []model.TaskTemplateStep{
				{ID: generateID(), StepOrder: 1, Name: "代码任务命令", Command: "echo run code task", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
			},
			Plugins:   []model.TaskTemplatePlugin{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "执行命令",
			Category:    "执行命令",
			SubCategory: "Shell命令",
			Description: "直接执行自定义命令",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildEnvironment": "default",
				"timeoutMinutes":   60,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":      "DEFAULT",
				"timeoutMinutes": 60,
				"debugMode":      false,
				"deployTask":     false,
				"dockerDaemon":   false,
			}),
			Steps: []model.TaskTemplateStep{
				{ID: generateID(), StepOrder: 1, Name: "执行命令", Command: "echo hello pipeline", Envs: mustMarshal(map[string]interface{}{}), CreatedAt: time.Now()},
			},
			Plugins:   []model.TaskTemplatePlugin{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          generateID(),
			UserID:      userID,
			Name:        "空模板",
			Category:    "空模板",
			SubCategory: "空模板",
			Description: "无预设步骤，完全自定义",
			PresetFields: mustMarshal(map[string]interface{}{
				"buildEnvironment": "default",
				"timeoutMinutes":   120,
			}),
			AdvancedSettings: mustMarshal(map[string]interface{}{
				"buildSpec":      "DEFAULT",
				"timeoutMinutes": 120,
				"debugMode":      false,
				"deployTask":     false,
				"dockerDaemon":   false,
			}),
			Steps:     []model.TaskTemplateStep{},
			Plugins:   []model.TaskTemplatePlugin{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}
