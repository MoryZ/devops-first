package service

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"devops-first/internal/database"
	"devops-first/internal/model"
)

// ExecutionService handles pipeline execution requests
type ExecutionService struct {
	queue *ExecutionQueue
}

type CommitRecord struct {
	CodeVersion string `json:"code_version"`
	Content     string `json:"content"`
	Author      string `json:"author"`
	CommittedAt string `json:"committed_at"`
}

// NewExecutionService creates a new execution service
func NewExecutionService(queue *ExecutionQueue) *ExecutionService {
	return &ExecutionService{
		queue: queue,
	}
}

// ExecuteRequest represents a pipeline execution request
type ExecuteRequest struct {
	SystemID    string `json:"system_id"`
	PipelineID  string `json:"pipeline_id"`
	TriggeredBy string `json:"triggered_by"` // manual, webhook, schedule
	StartNodeID string `json:"start_node_id"`
	UserID      uint   `json:"user_id"`
}

// ExecuteResponse represents the response after submitting execution
type ExecuteResponse struct {
	BatchID string `json:"batch_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// SubmitExecution submits a new pipeline execution
func (es *ExecutionService) SubmitExecution(req *ExecuteRequest) (*ExecuteResponse, error) {
	db := database.GetDB()

	// Get pipeline config
	var pipeline model.PipelineConfig
	if err := db.Where("pipeline_id = ? AND user_id = ?", req.PipelineID, req.UserID).First(&pipeline).Error; err != nil {
		// If config not found, create a default one
		pipeline = model.PipelineConfig{
			UserID:        req.UserID,
			PipelineID:    req.PipelineID,
			Name:          "Default Pipeline",
			RepositoryType: "git",
			AutoMerge:     true,
			AutoTag:       true,
			DisplayOrder:  0,
			RepoURL:       "",
			Branch:        "main",
			GitUsername:   "",
			GitToken:      "",
			ProjectPath:   "/tmp/testproject",
			BuildType:     "maven",
			MavenCommand:  "mvn clean package -DskipTests",
			DeployType:    "docker",
			DockerImage:   "openjdk:11-jre",
			DockerContainer: "test-container",
			DockerRunArgs: "-d -p 8080:8080",
			MainStagesJSON: `[{"name":"checkout","type":"checkout"},{"name":"build","type":"build"},{"name":"deploy","type":"deploy"}]`,
			EnvStagesJSON:  `[]`,
		}
		if err := db.Create(&pipeline).Error; err != nil {
			return nil, fmt.Errorf("failed to create default pipeline config: %v", err)
		}
	}

	// Get batch number
	var batchCount int64
	db.Model(&model.ExecutionBatch{}).
		Where("pipeline_id = ? AND user_id = ?", req.PipelineID, req.UserID).
		Count(&batchCount)

	batchNumber := int(batchCount) + 1

	// Create batch record
	batch := &model.ExecutionBatch{
		ID:           es.queue.GenerateBatchID(),
		UserID:       req.UserID,
		SystemID:     req.SystemID,
		PipelineID:   req.PipelineID,
		PipelineName: pipeline.Name,
		BatchNumber:  batchNumber,
		Status:       "pending",
		TriggeredBy:  req.TriggeredBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := db.Create(batch).Error; err != nil {
		return nil, fmt.Errorf("failed to create batch record: %v", err)
	}

	// Submit to queue
	if err := es.queue.SubmitTask(batch, &pipeline, req.UserID, req.SystemID, req.StartNodeID); err != nil {
		// Update batch status to failed
		batch.Status = "failed"
		batch.ErrorMessage = err.Error()
		db.Model(batch).Updates(batch)
		return nil, fmt.Errorf("failed to submit task to queue: %v", err)
	}

	return &ExecuteResponse{
		BatchID: batch.ID,
		Status:  "pending",
		Message: "Execution submitted to queue",
	}, nil
}

// GetBatchStatus returns the current status of a batch
func (es *ExecutionService) GetBatchStatus(batchID string) (*model.ExecutionBatch, error) {
	db := database.GetDB()
	var batch model.ExecutionBatch
	if err := db.Where("id = ?", batchID).First(&batch).Error; err != nil {
		return nil, fmt.Errorf("batch not found: %v", err)
	}
	return &batch, nil
}

// ListBatchesForPipeline returns execution history for a pipeline
func (es *ExecutionService) ListBatchesForPipeline(userID uint, pipelineID string, limit int) ([]*model.ExecutionBatch, error) {
	var batches []*model.ExecutionBatch
	db := database.GetDB()

	if limit == 0 {
		limit = 50 // Default limit
	}

	if err := db.Where("user_id = ? AND pipeline_id = ?", userID, pipelineID).
		Order("created_at DESC").
		Limit(limit).
		Find(&batches).Error; err != nil {
		return nil, fmt.Errorf("failed to list batches: %v", err)
	}

	return batches, nil
}

// GetBatchLogs returns logs for a batch
func (es *ExecutionService) GetBatchLogs(batchID string, limit int) ([]*model.ExecutionLog, error) {
	var logs []*model.ExecutionLog
	db := database.GetDB()

	if limit == 0 {
		limit = 1000 // Default limit
	}

	if err := db.Where("batch_id = ?", batchID).
		Order("created_at ASC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get batch logs: %v", err)
	}

	return logs, nil
}

// GetBatchCommits returns commit history for the batch's work tree.
func (es *ExecutionService) GetBatchCommits(batchID string, limit int) ([]CommitRecord, error) {
	if limit <= 0 {
		limit = 20
	}

	batch, err := es.GetBatchStatus(batchID)
	if err != nil {
		return nil, err
	}

	// Fallback if workdir is unavailable.
	fallback := func() []CommitRecord {
		if strings.TrimSpace(batch.CommitID) == "" {
			return []CommitRecord{}
		}
		return []CommitRecord{{
			CodeVersion: batch.CommitID,
			Content:     "-",
			Author:      "-",
			CommittedAt: "-",
		}}
	}

	workDir := strings.TrimSpace(batch.WorkDir)
	if workDir == "" {
		return fallback(), nil
	}

	format := "%H%x1f%an%x1f%ad%x1f%s"
	cmd := exec.Command("git", "-C", workDir, "log", "--date=format:%Y-%m-%d %H:%M:%S", "--pretty=format:"+format, "-n", strconv.Itoa(limit))
	out, err := cmd.Output()
	if err != nil {
		return fallback(), nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	res := make([]CommitRecord, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\x1f")
		rec := CommitRecord{}
		if len(parts) > 0 {
			rec.CodeVersion = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			rec.Author = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			rec.CommittedAt = strings.TrimSpace(parts[2])
		}
		if len(parts) > 3 {
			rec.Content = strings.TrimSpace(parts[3])
		}
		res = append(res, rec)
	}

	if len(res) == 0 {
		return fallback(), nil
	}
	return res, nil
}

// CancelBatch cancels a pending or running batch
func (es *ExecutionService) CancelBatch(batchID string) error {
	db := database.GetDB()
	batch, err := es.GetBatchStatus(batchID)
	if err != nil {
		return err
	}

	if batch.Status == "success" || batch.Status == "failed" {
		return fmt.Errorf("cannot cancel a completed batch with status: %s", batch.Status)
	}

	// Stop in-flight process immediately if running.
	es.queue.CancelBatch(batchID)

	batch.Status = "cancelled"
	now := time.Now()
	batch.CompletedAt = &now
	batch.UpdatedAt = time.Now()
	return db.Model(batch).Updates(batch).Error
}

// RerunFromNode creates a new batch and starts execution from a specific BPM node.
func (es *ExecutionService) RerunFromNode(userID uint, batchID string, nodeID string) (*ExecuteResponse, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("node_id is required")
	}
	original, err := es.GetBatchStatus(batchID)
	if err != nil {
		return nil, err
	}
	if original.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	return es.SubmitExecution(&ExecuteRequest{
		SystemID:    original.SystemID,
		PipelineID:  original.PipelineID,
		TriggeredBy: "rerun-node",
		StartNodeID: nodeID,
		UserID:      userID,
	})
}

// CleanupOldLogs removes execution logs older than retention period
func (es *ExecutionService) CleanupOldLogs(retentionDays int) error {
	db := database.GetDB()
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Delete logs older than retention period
	if err := db.Where("created_at < ?", cutoffDate).Delete(&model.ExecutionLog{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old logs: %v", err)
	}

	// Optionally also delete associated batch records
	if err := db.Where("created_at < ? AND status IN ?", cutoffDate, []string{"success", "failed"}).Delete(&model.ExecutionBatch{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old batches: %v", err)
	}

	return nil
}

// GetQueueStats returns current queue statistics
func (es *ExecutionService) GetQueueStats() map[string]interface{} {
	return es.queue.GetQueueStats()
}

// RegisterLogCallback registers a live log callback for a batch.
func (es *ExecutionService) RegisterLogCallback(batchID string, callback func(line string, level string)) {
	es.queue.RegisterLogCallback(batchID, callback)
}

// UnregisterLogCallback removes the live log callback for a batch.
func (es *ExecutionService) UnregisterLogCallback(batchID string) {
	es.queue.UnregisterLogCallback(batchID)
}
