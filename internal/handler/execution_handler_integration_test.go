package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"devops-first/internal/database"
	"devops-first/internal/model"
	"devops-first/internal/service"
)

func setupExecutionTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	database.DB = db

	if err := model.AutoMigratePipelineConfigs(db); err != nil {
		t.Fatalf("migrate pipeline configs: %v", err)
	}
	if err := model.AutoMigrateExecutionBatch(db); err != nil {
		t.Fatalf("migrate execution batches: %v", err)
	}
	if err := model.AutoMigrateExecutionLog(db); err != nil {
		t.Fatalf("migrate execution logs: %v", err)
	}

	if err := db.Create(&model.PipelineConfig{
		UserID:          1,
		PipelineID:      "pipeline-1",
		Name:            "pipeline-1",
		BuildType:       "none",
		DeployType:      "docker",
		DockerImage:     "demo:latest",
		DockerContainer: "demo",
	}).Error; err != nil {
		t.Fatalf("seed pipeline config: %v", err)
	}

	queue := service.NewExecutionQueue(1, t.TempDir())
	es := service.NewExecutionService(queue)
	SetExecutionService(es)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})
	r.POST("/api/pipelines/:id/execute", HandleSubmitExecution)
	r.GET("/api/executions/:batch_id", HandleGetBatchStatus)
	r.GET("/api/executions/:batch_id/logs", HandleGetBatchLogs)
	r.GET("/api/pipelines/:id/executions", HandleGetBatchHistory)
	r.GET("/ws/execute/:batch_id", HandleExecutionWebSocket)

	return r
}

func TestExecutionSubmitAndQueryEndpoints(t *testing.T) {
	router := setupExecutionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/pipelines/pipeline-1/execute?system_id=sys-1&triggered_by=manual", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("submit status=%d body=%s", w.Code, w.Body.String())
	}

	var submitResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &submitResp); err != nil {
		t.Fatalf("unmarshal submit response: %v", err)
	}
	batchID, _ := submitResp["batch_id"].(string)
	if batchID == "" {
		t.Fatalf("empty batch_id in response: %s", w.Body.String())
	}

	waitForBatchTerminalStatus(t, router, batchID)

	statusReq := httptest.NewRequest(http.MethodGet, "/api/executions/"+batchID, nil)
	statusW := httptest.NewRecorder()
	router.ServeHTTP(statusW, statusReq)
	if statusW.Code != http.StatusOK {
		t.Fatalf("status endpoint code=%d body=%s", statusW.Code, statusW.Body.String())
	}

	logsReq := httptest.NewRequest(http.MethodGet, "/api/executions/"+batchID+"/logs?limit=200", nil)
	logsW := httptest.NewRecorder()
	router.ServeHTTP(logsW, logsReq)
	if logsW.Code != http.StatusOK {
		t.Fatalf("logs endpoint code=%d body=%s", logsW.Code, logsW.Body.String())
	}
	if !strings.Contains(logsW.Body.String(), "Start executing pipeline") {
		t.Fatalf("expected execution logs, got: %s", logsW.Body.String())
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/api/pipelines/pipeline-1/executions?limit=10", nil)
	historyW := httptest.NewRecorder()
	router.ServeHTTP(historyW, historyReq)
	if historyW.Code != http.StatusOK {
		t.Fatalf("history endpoint code=%d body=%s", historyW.Code, historyW.Body.String())
	}
	if !strings.Contains(historyW.Body.String(), batchID) {
		t.Fatalf("history missing batch id, body=%s", historyW.Body.String())
	}
}

func TestExecutionWebSocketStreamsLogs(t *testing.T) {
	router := setupExecutionTestRouter(t)

	submitReq := httptest.NewRequest(http.MethodPost, "/api/pipelines/pipeline-1/execute?system_id=sys-1", nil)
	submitW := httptest.NewRecorder()
	router.ServeHTTP(submitW, submitReq)
	if submitW.Code != http.StatusOK {
		t.Fatalf("submit status=%d body=%s", submitW.Code, submitW.Body.String())
	}

	var submitResp map[string]interface{}
	if err := json.Unmarshal(submitW.Body.Bytes(), &submitResp); err != nil {
		t.Fatalf("unmarshal submit response: %v", err)
	}
	batchID, _ := submitResp["batch_id"].(string)
	if batchID == "" {
		t.Fatalf("empty batch_id")
	}

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/execute/" + batchID
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	var foundConnected bool
	var foundLog bool

	for i := 0; i < 12; i++ {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read ws message: %v", err)
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			t.Fatalf("unmarshal ws payload: %v, raw=%s", err, string(msg))
		}

		msgType, _ := payload["type"].(string)
		if msgType == "connected" {
			foundConnected = true
		}
		if msgType == "log" {
			foundLog = true
			break
		}
	}

	if !foundConnected {
		t.Fatalf("did not receive connected event")
	}
	if !foundLog {
		t.Fatalf("did not receive log event from websocket")
	}
}

func waitForBatchTerminalStatus(t *testing.T, router *gin.Engine, batchID string) {
	t.Helper()

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet, "/api/executions/"+batchID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			body := w.Body.String()
			if strings.Contains(body, "\"status\":\"success\"") || strings.Contains(body, "\"status\":\"failed\"") || strings.Contains(body, "\"status\":\"cancelled\"") {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("batch did not reach terminal state: %s", batchID)
}
