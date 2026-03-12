package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"devops-first/internal/service"
)

// HandleGetBPMDefinition handles GET /api/pipelines/:id/bpm
func HandleGetBPMDefinition(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	pipelineID := c.Param("id")
	svc := service.NewBPMDefinitionService()
	item, err := svc.Get(userID, pipelineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// HandleUpsertBPMDefinition handles PUT /api/pipelines/:id/bpm
func HandleUpsertBPMDefinition(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	pipelineID := c.Param("id")
	var req service.BPMDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.PipelineID == "" {
		req.PipelineID = pipelineID
	}
	if req.PipelineID != pipelineID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline id mismatch"})
		return
	}

	svc := service.NewBPMDefinitionService()
	if err := svc.Upsert(userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
