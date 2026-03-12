package handler

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"devops-first/internal/service"
)

type DeployHandler struct {
	svc      *service.DeploymentService
	upgrader websocket.Upgrader
}

func NewDeployHandler(svc *service.DeploymentService) *DeployHandler {
	return &DeployHandler{
		svc: svc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *DeployHandler) HandleDeployWS(c *gin.Context) {
	projectPath := c.Query("project_path")
	if projectPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_path query parameter is required"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "websocket upgrade failed", "details": err.Error()})
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	var writeMu sync.Mutex
	send := func(line string) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteJSON(gin.H{
			"type":      "log",
			"message":   line,
			"timestamp": time.Now().Format(time.RFC3339Nano),
		})
	}

	err = h.svc.Deploy(ctx, projectPath, send)
	if err != nil {
		if errors.Is(err, service.ErrDeploymentInProgress) {
			_ = send("[error] deployment already in progress for this project")
			return
		}
		if !errors.Is(err, context.Canceled) {
			_ = send("[error] deployment failed: " + err.Error())
		}
		return
	}

	_ = send("[done] websocket session complete")
}
