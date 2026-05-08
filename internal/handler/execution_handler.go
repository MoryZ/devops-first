package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"devops-first/internal/service"
)

var executionService *service.ExecutionService
var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// SetExecutionService sets the execution service instance
func SetExecutionService(es *service.ExecutionService) {
	executionService = es
}

// HandleSubmitExecution handles POST /api/pipelines/:id/execute
func HandleSubmitExecution(c *gin.Context) {
	pipelineID := c.Param("id")
	systemID := c.Query("system_id")

	if pipelineID == "" || systemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline_id and system_id are required"})
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type"})
		return
	}

	triggeredBy := c.Query("triggered_by")
	if triggeredBy == "" {
		triggeredBy = "manual"
	}
	startNodeID := c.Query("start_node_id")

	req := &service.ExecuteRequest{
		SystemID:    systemID,
		PipelineID:  pipelineID,
		TriggeredBy: triggeredBy,
		StartNodeID: startNodeID,
		UserID:      userID,
	}

	resp, err := executionService.SubmitExecution(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HandleGetBatchStatus handles GET /api/executions/:batch_id
func HandleGetBatchStatus(c *gin.Context) {
	batchID := c.Param("batch_id")

	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id is required"})
		return
	}

	batch, err := executionService.GetBatchStatus(batchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, batch)
}

// HandleGetBatchLogs handles GET /api/executions/:batch_id/logs
func HandleGetBatchLogs(c *gin.Context) {
	batchID := c.Param("batch_id")
	limitStr := c.Query("limit")

	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id is required"})
		return
	}

	limit := 1000
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := executionService.GetBatchLogs(batchID, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs, "count": len(logs)})
}

// HandleGetBatchCommits handles GET /api/executions/:batch_id/commits
func HandleGetBatchCommits(c *gin.Context) {
	batchID := c.Param("batch_id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id is required"})
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	items, err := executionService.GetBatchCommits(batchID, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

// HandleGetBatchHistory handles GET /api/pipelines/:id/executions
func HandleGetBatchHistory(c *gin.Context) {
	pipelineID := c.Param("id")
	limitStr := c.Query("limit")

	if pipelineID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline_id is required"})
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type"})
		return
	}

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	batches, err := executionService.ListBatchesForPipeline(userID, pipelineID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"batches": batches, "count": len(batches)})
}

// HandleCancelBatch handles POST /api/executions/:batch_id/cancel
func HandleCancelBatch(c *gin.Context) {
	batchID := c.Param("batch_id")

	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id is required"})
		return
	}

	if err := executionService.CancelBatch(batchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "batch cancelled"})
}

// HandleRerunNode handles POST /api/executions/:batch_id/rerun-node
func HandleRerunNode(c *gin.Context) {
	batchID := c.Param("batch_id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id is required"})
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type"})
		return
	}

	var req struct {
		NodeID string `json:"node_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := executionService.RerunFromNode(userID, batchID, req.NodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// HandleExecutionWebSocket handles WS /ws/execute/:batch_id
func HandleExecutionWebSocket(c *gin.Context) {
	batchID := c.Param("batch_id")

	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id is required"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer ws.Close()
	defer executionService.UnregisterLogCallback(batchID)

	// Send initial message
	if err := ws.WriteJSON(gin.H{
		"type":     "connected",
		"batch_id": batchID,
	}); err != nil {
		return
	}

	// Replay recent logs so newly connected clients can catch up.
	if logs, err := executionService.GetBatchLogs(batchID, 200); err == nil {
		for _, l := range logs {
			if err := ws.WriteJSON(gin.H{
				"type":       "log",
				"batch_id":   batchID,
				"level":      l.LogLevel,
				"line":       l.LogLine,
				"created_at": l.CreatedAt,
			}); err != nil {
				return
			}
		}
	}

	executionService.RegisterLogCallback(batchID, func(line string, level string) {
		_ = ws.WriteJSON(gin.H{
			"type":     "log",
			"batch_id": batchID,
			"level":    level,
			"line":     line,
		})
	})

	// Keep connection alive and listen for messages
	for {
		var msg map[string]interface{}
		if err := ws.ReadJSON(&msg); err != nil {
			break
		}
	}
}

// HandleQueueStats handles GET /api/queue/stats
func HandleQueueStats(c *gin.Context) {
	stats := executionService.GetQueueStats()
	c.JSON(http.StatusOK, stats)
}
