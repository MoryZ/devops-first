package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"devops-first/internal/database"
	"devops-first/internal/handler"
	"devops-first/internal/middleware"
	"devops-first/internal/model"
	"devops-first/internal/service"
)

// Config mocks a minimal app configuration layer.
type Config struct {
	HTTPAddr   string
	GitPath    string
	MavenPath  string
	DockerPath string
	GinMode    string
	Proxies    []string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func LoadConfig() Config {
	return Config{
		HTTPAddr:   getEnv("HTTP_ADDR", ":8081"),
		GitPath:    getEnv("GIT_PATH", "git"),
		MavenPath:  getEnv("MVN_PATH", "mvn"),
		DockerPath: getEnv("DOCKER_PATH", "docker"),
		GinMode:    getEnv("GIN_MODE", gin.DebugMode),
		Proxies:    splitCSV(getEnv("TRUSTED_PROXIES", "127.0.0.1,::1")),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "devops_first"),
	}
}

func main() {
	loadEnvFiles()

	cfg := LoadConfig()
	gin.SetMode(cfg.GinMode)

	// Initialize database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	if err := database.InitDB(dsn); err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	// Auto migrate tables
	if err := model.AutoMigrateUsers(database.GetDB()); err != nil {
		log.Fatalf("auto migrate users failed: %v", err)
	}
	if err := model.AutoMigratePipelineConfigs(database.GetDB()); err != nil {
		log.Fatalf("auto migrate pipeline configs failed: %v", err)
	}
	if err := model.AutoMigrateSystem(database.GetDB()); err != nil {
		log.Fatalf("auto migrate system failed: %v", err)
	}
	if err := model.AutoMigrateExecutionBatch(database.GetDB()); err != nil {
		log.Fatalf("auto migrate execution batch failed: %v", err)
	}
	if err := model.AutoMigrateExecutionLog(database.GetDB()); err != nil {
		log.Fatalf("auto migrate execution log failed: %v", err)
	}
	if err := model.AutoMigrateBPMDefinitions(database.GetDB()); err != nil {
		log.Fatalf("auto migrate bpm definitions failed: %v", err)
	}
	if err := model.AutoMigrateTaskTemplates(database.GetDB()); err != nil {
		log.Fatalf("auto migrate task templates failed: %v", err)
	}
	if err := model.AutoMigrateGlobalVariables(database.GetDB()); err != nil {
		log.Fatalf("auto migrate global variables failed: %v", err)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	// Configure CORS to allow frontend requests with Authorization header
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))
	if err := router.SetTrustedProxies(cfg.Proxies); err != nil {
		log.Fatalf("set trusted proxies failed: %v", err)
	}

	deploySvc := service.NewDeploymentService(service.Config{
		GitPath:    cfg.GitPath,
		MavenPath:  cfg.MavenPath,
		DockerPath: cfg.DockerPath,
	})
	deployHandler := handler.NewDeployHandler(deploySvc)

	// Initialize execution queue (max 10 concurrent executions, /tmp/devops-exec base directory)
	executionQueue := service.NewExecutionQueue(10, "/tmp/devops-exec")
	executionService := service.NewExecutionService(executionQueue)
	handler.SetExecutionService(executionService)

	// Initialize task template service
	taskTemplateService := service.NewTaskTemplateService(database.GetDB())
	taskTemplateHandler := handler.NewTaskTemplateHandler(taskTemplateService)

	// Health check (public)
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", handler.HandleLogin)
		authGroup.POST("/register", handler.HandleRegister)
	}

	// Public API routes (no JWT required)
	publicGroup := router.Group("")
	{
		// Task template routes (read-only, public)
		publicGroup.GET("/api/task-templates", taskTemplateHandler.GetAllTemplates)
		publicGroup.GET("/api/task-templates/category", taskTemplateHandler.GetTemplatesByCategory)
		publicGroup.GET("/api/task-templates/:id", taskTemplateHandler.GetTemplateByID)
	}

	// Protected routes (require JWT)
	protectedGroup := router.Group("")
	protectedGroup.Use(middleware.JWTMiddleware())
	{
		protectedGroup.GET("/ws/deploy", deployHandler.HandleDeployWS)
		protectedGroup.GET("/api/pipelines", handler.HandleListPipelineConfigs)
		protectedGroup.GET("/api/pipelines/:id/config", handler.HandleGetPipelineConfig)
		protectedGroup.PUT("/api/pipelines/config", handler.HandleUpsertPipelineConfig)
		protectedGroup.PUT("/api/pipelines/:id/release-unit", handler.HandleUpsertPipelineReleaseUnitBinding)

		// Pipeline execution routes
		protectedGroup.POST("/api/pipelines/:id/execute", handler.HandleSubmitExecution)
		protectedGroup.GET("/api/pipelines/:id/executions", handler.HandleGetBatchHistory)
		protectedGroup.GET("/api/executions/:batch_id", handler.HandleGetBatchStatus)
		protectedGroup.GET("/api/executions/:batch_id/logs", handler.HandleGetBatchLogs)
		protectedGroup.GET("/api/executions/:batch_id/commits", handler.HandleGetBatchCommits)
		protectedGroup.POST("/api/executions/:batch_id/cancel", handler.HandleCancelBatch)
		protectedGroup.POST("/api/executions/:batch_id/rerun-node", handler.HandleRerunNode)
		protectedGroup.GET("/ws/execute/:batch_id", handler.HandleExecutionWebSocket)
		protectedGroup.GET("/api/queue/stats", handler.HandleQueueStats)

		// System and Plan management routes
		protectedGroup.GET("/api/systems", handler.HandleListSystems)
		protectedGroup.POST("/api/systems", handler.HandleCreateSystem)
		protectedGroup.GET("/api/systems/:system_id", handler.HandleGetSystem)
		protectedGroup.GET("/api/systems/:system_id/plans", handler.HandleListPlans)
		protectedGroup.POST("/api/systems/:system_id/plans", handler.HandleCreatePlan)
		protectedGroup.GET("/api/systems/:system_id/pipelines", handler.HandleListPipelinesForSystem)
		protectedGroup.POST("/api/systems/:system_id/pipelines", handler.HandleCreatePipeline)
		protectedGroup.GET("/api/plans/:plan_id/pipelines", handler.HandleListPipelinesForPlan)
		protectedGroup.GET("/api/pipelines/:id", handler.HandleGetPipeline)
		protectedGroup.GET("/api/pipelines/:id/bpm", handler.HandleGetBPMDefinition)
		protectedGroup.PUT("/api/pipelines/:id/bpm", handler.HandleUpsertBPMDefinition)
		protectedGroup.GET("/api/global-vars", handler.HandleListGlobalVariables)
		protectedGroup.PUT("/api/global-vars", handler.HandleUpsertGlobalVariable)
		protectedGroup.DELETE("/api/global-vars/:id", handler.HandleDeleteGlobalVariable)

		// Task template routes (write/delete, protected)
		protectedGroup.POST("/api/task-templates/init", taskTemplateHandler.InitializeTemplates)
		protectedGroup.POST("/api/task-templates", taskTemplateHandler.CreateTemplate)
		protectedGroup.DELETE("/api/task-templates/:id", taskTemplateHandler.DeleteTemplate)
	}

	log.Printf("server listening on %s", cfg.HTTPAddr)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func loadEnvFiles() {
	loaded := make([]string, 0, 3)
	for _, file := range []string{".env", ".env.local", ".env.example"} {
		if _, err := os.Stat(file); err == nil {
			if err := godotenv.Load(file); err != nil {
				log.Printf("warning: failed to load %s: %v", file, err)
				continue
			}
			loaded = append(loaded, file)
		}
	}

	if len(loaded) > 0 {
		log.Printf("loaded env files: %s", strings.Join(loaded, ", "))
	}
}
