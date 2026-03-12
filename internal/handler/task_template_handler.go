package handler

import (
	"encoding/json"
	"net/http"

	"devops-first/internal/model"
	"devops-first/internal/service"
	"github.com/gin-gonic/gin"
)

type TaskTemplateHandler struct {
	templateService *service.TaskTemplateService
}

func NewTaskTemplateHandler(templateService *service.TaskTemplateService) *TaskTemplateHandler {
	return &TaskTemplateHandler{
		templateService: templateService,
	}
}

// InitializeTemplates initializes default templates for a user
func (h *TaskTemplateHandler) InitializeTemplates(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "missing user_id"})
		return
	}

	if err := h.templateService.InitializeDefaultTemplates(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "templates initialized"})
}

// GetAllTemplates retrieves all templates for the current user (or public templates if not authenticated)
func (h *TaskTemplateHandler) GetAllTemplates(c *gin.Context) {
	userID := c.GetString("user_id")
	
	// If not authenticated, use "public" userID for global templates
	if userID == "" {
		userID = "public"
		// Always ensure public templates are initialized
		_ = h.templateService.InitializeDefaultTemplates(userID)
	}

	templates, err := h.templateService.GetAllTemplates(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": templates,
	})
}

// GetTemplatesByCategory retrieves templates filtered by category (or public templates if not authenticated)
func (h *TaskTemplateHandler) GetTemplatesByCategory(c *gin.Context) {
	userID := c.GetString("user_id")

	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "category is required"})
		return
	}

	// If not authenticated, use "public" userID for global templates
	if userID == "" {
		userID = "public"
		// Always ensure public templates are initialized
		_ = h.templateService.InitializeDefaultTemplates(userID)
	}

	templates, err := h.templateService.GetTemplatesByCategory(userID, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": templates,
	})
}

// GetTemplateByID retrieves a template by ID
func (h *TaskTemplateHandler) GetTemplateByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "template id is required"})
		return
	}

	template, err := h.templateService.GetTemplateByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": template,
	})
}

// CreateTemplate creates a new template
func (h *TaskTemplateHandler) CreateTemplate(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "missing user_id"})
		return
	}

	var req struct {
		Name             string                 `json:"name"`
		Category         string                 `json:"category"`
		SubCategory      string                 `json:"sub_category"`
		Description      string                 `json:"description"`
		PresetFields     map[string]interface{} `json:"preset_fields"`
		AdvancedSettings map[string]interface{} `json:"advanced_settings"`
		Steps            []struct {
			Name           string                 `json:"name"`
			Command        string                 `json:"command"`
			ShellSpecified bool                   `json:"shell_specified"`
			Envs           map[string]interface{} `json:"envs"`
		} `json:"steps"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	// Helper to convert map to json.RawMessage
	toRawMessage := func(v map[string]interface{}) json.RawMessage {
		data, _ := json.Marshal(v)
		return data
	}

	template := &model.TaskTemplate{
		UserID:           userID,
		Name:             req.Name,
		Category:         req.Category,
		SubCategory:      req.SubCategory,
		Description:      req.Description,
		PresetFields:     toRawMessage(req.PresetFields),
		AdvancedSettings: toRawMessage(req.AdvancedSettings),
		Steps:            []model.TaskTemplateStep{},
	}

	for i, step := range req.Steps {
		template.Steps = append(template.Steps, model.TaskTemplateStep{
			StepOrder:      i + 1,
			Name:           step.Name,
			Command:        step.Command,
			ShellSpecified: step.ShellSpecified,
			Envs:           toRawMessage(step.Envs),
		})
	}

	if err := h.templateService.CreateTemplate(template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": template,
	})
}

// DeleteTemplate deletes a template
func (h *TaskTemplateHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "template id is required"})
		return
	}

	if err := h.templateService.DeleteTemplate(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "template deleted"})
}
