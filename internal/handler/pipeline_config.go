package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"devops-first/internal/service"
)

type PipelineReleaseUnitBindingRequest struct {
	ReleaseUnitID string `json:"release_unit_id"`
}

func getUserID(c *gin.Context) (uint, bool) {
	raw, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	userID, ok := raw.(uint)
	return userID, ok
}

func parsePositiveInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// HandleListPipelineConfigs returns all pipeline configs for current user.
func HandleListPipelineConfigs(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	svc := service.NewPipelineConfigService()
	items, err := svc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// HandleUpsertPipelineConfig creates or updates one pipeline config.
func HandleUpsertPipelineConfig(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req service.PipelineConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := service.NewPipelineConfigService()
	if err := svc.Upsert(userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HandleGetPipelineConfig returns one pipeline config by pipeline ID.
func HandleGetPipelineConfig(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	pipelineID := c.Param("id")
	if pipelineID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline id is required"})
		return
	}

	svc := service.NewPipelineConfigService()
	item, err := svc.Get(userID, pipelineID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// HandleUpsertPipelineReleaseUnitBinding creates or updates only release unit binding for a pipeline.
func HandleUpsertPipelineReleaseUnitBinding(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	pipelineID := c.Param("id")
	if pipelineID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline id is required"})
		return
	}

	var req PipelineReleaseUnitBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := service.NewPipelineConfigService()
	if err := svc.UpsertReleaseUnitBinding(userID, pipelineID, req.ReleaseUnitID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
