package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"devops-first/internal/service"
)

func getUserID(c *gin.Context) (uint, bool) {
	raw, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	userID, ok := raw.(uint)
	return userID, ok
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
