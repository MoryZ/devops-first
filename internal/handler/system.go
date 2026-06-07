package handler

import (
	"net/http"

	"devops-first/internal/service"

	"github.com/gin-gonic/gin"
)

// HandleListSystems - GET /api/systems
func HandleListSystems(c *gin.Context) {
	userIDNum, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := parsePositiveInt(p); err == nil {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := parsePositiveInt(ps); err == nil && parsed <= 200 {
			pageSize = parsed
		}
	}
	keyword := c.Query("keyword")

	total, systems, err := service.ListSystemsPaginated(userIDNum, page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":      systems,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// HandleCreateSystem - POST /api/systems
func HandleCreateSystem(c *gin.Context) {
	userIDNum, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	system, err := service.CreateSystem(userIDNum, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, system)
}

// HandleGetSystem - GET /api/systems/:system_id
func HandleGetSystem(c *gin.Context) {
	userIDNum, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	systemID := c.Param("system_id")

	system, err := service.GetSystem(userIDNum, systemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, system)
}

// HandleUpdateSystem - PUT /api/systems/:system_id
func HandleUpdateSystem(c *gin.Context) {
	userIDNum, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	systemID := c.Param("system_id")

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	system, err := service.UpdateSystem(userIDNum, systemID, req.Name, req.Description, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, system)
}

// HandleDeleteSystem - DELETE /api/systems/:system_id
func HandleDeleteSystem(c *gin.Context) {
	userIDNum, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	systemID := c.Param("system_id")

	if err := service.DeleteSystem(userIDNum, systemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "system deleted"})
}

// HandleListPlans - GET /api/systems/:system_id/plans
func HandleListPlans(c *gin.Context) {
	systemID := c.Param("system_id")

	plans, err := service.ListPlans(systemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": plans})
}

// HandleCreatePlan - POST /api/systems/:system_id/plans
func HandleCreatePlan(c *gin.Context) {
	systemID := c.Param("system_id")

	var req struct {
		Version     string `json:"version" binding:"required"`
		Status      string `json:"status"`
		PlannedDate string `json:"planned_date"`
		Description string `json:"description"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status == "" {
		req.Status = "planning"
	}

	plan, err := service.CreatePlan(systemID, req.Version, req.Status, req.PlannedDate, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// HandleListPipelinesForPlan - GET /api/plans/:plan_id/pipelines
func HandleListPipelinesForPlan(c *gin.Context) {
	planID := c.Param("plan_id")

	pipelines, err := service.ListPipelinesForPlan(planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": pipelines})
}

// HandleListPipelinesForSystem - GET /api/systems/:system_id/pipelines
func HandleListPipelinesForSystem(c *gin.Context) {
	systemID := c.Param("system_id")

	pipelines, err := service.ListPipelinesForSystem(systemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": pipelines})
}

// HandleCreatePipeline - POST /api/systems/:system_id/pipelines (or /api/plans/:plan_id/pipelines)
func HandleCreatePipeline(c *gin.Context) {
	systemID := c.Param("system_id")
	if systemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing system_id in path"})
		return
	}

	var req struct {
		PlanID      string `json:"plan_id"`
		Name        string `json:"name" binding:"required"`
		AppType     string `json:"app_type" binding:"required"` // java, node, sql
		Description string `json:"description"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pipeline, err := service.CreatePipeline(systemID, req.PlanID, req.Name, req.AppType, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pipeline)
}

// HandleGetPipeline - GET /api/pipelines/:pipeline_id
func HandleGetPipeline(c *gin.Context) {
	pipelineID := c.Param("id")
	if pipelineID == "" {
		pipelineID = c.Param("pipeline_id")
	}

	pipeline, err := service.GetPipeline(pipelineID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pipeline)
}
