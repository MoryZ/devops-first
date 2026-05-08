package service

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"devops-first/internal/database"
	"devops-first/internal/model"
)

var errNoExecutableBPMNodes = errors.New("no executable bpm nodes")
var errBatchCancelled = errors.New("batch cancelled")

// ExecutionTask represents a single execution task
type ExecutionTask struct {
	Batch    *model.ExecutionBatch
	Pipeline *model.PipelineConfig
	UserID   uint
	SystemID string
	StartNodeID string
}

// ExecutionQueue manages concurrent pipeline executions
type ExecutionQueue struct {
	maxWorkers   int
	taskChan     chan *ExecutionTask
	mu           sync.Mutex
	baseWorkDir  string
	logCallbacks map[string]func(line string, level string)
	logMu        sync.RWMutex
	cancelMu     sync.Mutex
	cancelled    map[string]bool
	runningCmd   map[string]*exec.Cmd
}

// NewExecutionQueue creates a new execution queue with configured max workers
func NewExecutionQueue(maxWorkers int, baseWorkDir string) *ExecutionQueue {
	eq := &ExecutionQueue{
		maxWorkers:   maxWorkers,
		taskChan:     make(chan *ExecutionTask, maxWorkers*2),
		baseWorkDir:  baseWorkDir,
		logCallbacks: make(map[string]func(line string, level string)),
		cancelled:    make(map[string]bool),
		runningCmd:   make(map[string]*exec.Cmd),
	}
	if err := os.MkdirAll(baseWorkDir, 0755); err != nil {
		log.Printf("warning: failed to create base work dir: %v", err)
	}
	for i := 0; i < maxWorkers; i++ {
		go eq.worker()
	}
	return eq
}

// GenerateBatchID generates a unique batch ID
func (eq *ExecutionQueue) GenerateBatchID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("batch_%d_%s", time.Now().Unix(), hex.EncodeToString(b[:8]))
}

// SubmitTask submits a task to the queue
func (eq *ExecutionQueue) SubmitTask(batch *model.ExecutionBatch, pipeline *model.PipelineConfig, userID uint, systemID string, startNodeID string) error {
	task := &ExecutionTask{
		Batch:    batch,
		Pipeline: pipeline,
		UserID:   userID,
		SystemID: systemID,
		StartNodeID: startNodeID,
	}
	select {
	case eq.taskChan <- task:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("task queue is full, max concurrent executions: %d", eq.maxWorkers)
	}
}

// RegisterLogCallback registers a callback for log output (used for WebSocket streaming)
func (eq *ExecutionQueue) RegisterLogCallback(batchID string, callback func(line string, level string)) {
	eq.logMu.Lock()
	defer eq.logMu.Unlock()
	eq.logCallbacks[batchID] = callback
}

// UnregisterLogCallback removes the log callback
func (eq *ExecutionQueue) UnregisterLogCallback(batchID string) {
	eq.logMu.Lock()
	defer eq.logMu.Unlock()
	delete(eq.logCallbacks, batchID)
}

// CancelBatch marks a batch as cancelled and kills the currently running command if any.
func (eq *ExecutionQueue) CancelBatch(batchID string) {
	eq.cancelMu.Lock()
	eq.cancelled[batchID] = true
	cmd := eq.runningCmd[batchID]
	eq.cancelMu.Unlock()

	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func (eq *ExecutionQueue) clearCancel(batchID string) {
	eq.cancelMu.Lock()
	delete(eq.cancelled, batchID)
	eq.cancelMu.Unlock()
}

func (eq *ExecutionQueue) isCancelled(batchID string) bool {
	eq.cancelMu.Lock()
	defer eq.cancelMu.Unlock()
	return eq.cancelled[batchID]
}

func (eq *ExecutionQueue) setRunningCmd(batchID string, cmd *exec.Cmd) {
	eq.cancelMu.Lock()
	defer eq.cancelMu.Unlock()
	if cmd == nil {
		delete(eq.runningCmd, batchID)
		return
	}
	eq.runningCmd[batchID] = cmd
}

func (eq *ExecutionQueue) worker() {
	for task := range eq.taskChan {
		eq.executeTask(task)
	}
}

// setNodeStatus updates a single stage's status in batch.StagesStatusJSON and persists to DB.
func (eq *ExecutionQueue) setNodeStatus(batch *model.ExecutionBatch, stageName, status string) {
	current := make(map[string]string)
	if batch.StagesStatusJSON != "" {
		_ = json.Unmarshal([]byte(batch.StagesStatusJSON), &current)
	}
	current[stageName] = status
	b, _ := json.Marshal(current)
	batch.StagesStatusJSON = string(b)
	database.GetDB().Model(batch).Update("stages_status_json", batch.StagesStatusJSON)
}

// initNodeStatuses initializes all given stage names to "pending" in a single DB write.
func (eq *ExecutionQueue) initNodeStatuses(batch *model.ExecutionBatch, names []string) {
	statuses := make(map[string]string, len(names))
	for _, n := range names {
		statuses[n] = "pending"
	}
	b, _ := json.Marshal(statuses)
	batch.StagesStatusJSON = string(b)
	database.GetDB().Model(batch).Update("stages_status_json", batch.StagesStatusJSON)
}

// executeTask runs a full pipeline: source -> build -> deploy
func (eq *ExecutionQueue) executeTask(task *ExecutionTask) {
	batch := task.Batch
	eq.clearCancel(batch.ID)
	db := database.GetDB()
	workDir := eq.createWorkDir(batch)

	batch.Status = "running"
	batch.StartedAt = toPtr(time.Now())
	batch.WorkDir = workDir
	db.Model(batch).Updates(batch)

	eq.logLine(batch.ID, "info", fmt.Sprintf("Start executing pipeline: %s", batch.PipelineName))
	eq.logLine(batch.ID, "info", fmt.Sprintf("Work directory: %s", workDir))

	startTime := time.Now()
	var err error

	err = eq.executeByBPM(task, workDir)
	if errors.Is(err, errNoExecutableBPMNodes) {
		eq.logLine(batch.ID, "warn", "BPM 无可执行节点，回退到默认 source->build->deploy")
		err = eq.executeLegacyStages(task, workDir)
	}

	batch.CompletedAt = toPtr(time.Now())
	batch.TotalDuration = time.Since(startTime).Milliseconds()

	if err != nil {
		if errors.Is(err, errBatchCancelled) || eq.isCancelled(batch.ID) {
			batch.Status = "cancelled"
			batch.ErrorMessage = "batch cancelled"
			eq.logLine(batch.ID, "warn", "Execution cancelled")
		} else {
			batch.Status = "failed"
			batch.ErrorMessage = err.Error()
			eq.logLine(batch.ID, "error", fmt.Sprintf("Execution failed: %s", err.Error()))
		}
	} else {
		batch.Status = "success"
		eq.logLine(batch.ID, "info", fmt.Sprintf("Execution completed successfully in %.2fs", time.Since(startTime).Seconds()))
	}
	db.Model(batch).Updates(batch)
}

func (eq *ExecutionQueue) stageSource(batch *model.ExecutionBatch, pipeline *model.PipelineConfig, workDir string) error {
	eq.logLine(batch.ID, "info", "=== Stage: source ===")

	if pipeline.RepoURL == "" && pipeline.ProjectPath == "" {
		eq.logLine(batch.ID, "warn", "No repository URL configured, skipping source stage")
		return nil
	}
	if pipeline.RepoURL == "" && pipeline.ProjectPath != "" {
		eq.logLine(batch.ID, "info", "Using existing project_path without git pull")
		return nil
	}

	branch := pipeline.Branch
	if branch == "" {
		branch = "main"
	}

	cloneURL := withGitAuth(pipeline.RepoURL, pipeline.GitUsername, pipeline.GitToken)

	gitDir := filepath.Join(workDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		eq.logLine(batch.ID, "info", fmt.Sprintf("Repository exists, git pull origin %s", branch))
		if err := eq.runCmd(batch.ID, "source", workDir, "git", "pull", "origin", branch); err != nil {
			return err
		}
	} else {
		eq.logLine(batch.ID, "info", fmt.Sprintf("git clone [repo-hidden] (branch: %s)", branch))
		if err := eq.runCmd(batch.ID, "source", workDir, "git", "clone", "--branch", branch, "--depth", "1", cloneURL, "."); err != nil {
			return err
		}
	}

	// Capture commit ID and persist to DB
	if out, err := exec.Command("git", "-C", workDir, "rev-parse", "HEAD").Output(); err == nil {
		commitID := strings.TrimSpace(string(out))
		batch.CommitID = commitID
		database.GetDB().Model(batch).Update("commit_id", commitID)
		eq.logLine(batch.ID, "info", fmt.Sprintf("Commit: %s", commitID[:min(7, len(commitID))]))
	}
	return nil
}

func withGitAuth(repoURL, username, token string) string {
	if repoURL == "" || token == "" {
		return repoURL
	}
	u, err := url.Parse(repoURL)
	if err != nil {
		return repoURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return repoURL
	}
	if username == "" {
		username = "oauth2"
	}
	u.User = url.UserPassword(username, token)
	return u.String()
}

func (eq *ExecutionQueue) stageBuild(batch *model.ExecutionBatch, pipeline *model.PipelineConfig, workDir string) error {
	eq.logLine(batch.ID, "info", "=== Stage: build ===")

	buildDir := workDir
	if pipeline.ProjectPath != "" {
		if filepath.IsAbs(pipeline.ProjectPath) {
			buildDir = pipeline.ProjectPath
		} else {
			buildDir = filepath.Join(workDir, pipeline.ProjectPath)
		}
	}

	switch pipeline.BuildType {
	case "maven":
		cmd := pipeline.MavenCommand
		if cmd == "" {
			cmd = "mvn clean package -DskipTests"
		}
		eq.logLine(batch.ID, "info", fmt.Sprintf("Maven command: %s", cmd))
		return eq.runShell(batch.ID, "build", buildDir, cmd)
	case "gradle":
		cmd := pipeline.GradleCommand
		if cmd == "" {
			cmd = "./gradlew clean bootJar"
		}
		eq.logLine(batch.ID, "info", fmt.Sprintf("Gradle command: %s", cmd))
		return eq.runShell(batch.ID, "build", buildDir, cmd)
	case "npm":
		npmCmd := pipeline.NPMCommand
		if npmCmd == "" {
			npmCmd = "npm run build"
		}
		eq.logLine(batch.ID, "info", "Running npm install")
		if err := eq.runCmd(batch.ID, "build", buildDir, "npm", "install"); err != nil {
			return err
		}
		eq.logLine(batch.ID, "info", fmt.Sprintf("NPM command: %s", npmCmd))
		return eq.runShell(batch.ID, "build", buildDir, npmCmd)
	case "none":
		eq.logLine(batch.ID, "info", "Build skipped (build_type=none)")
		return nil
	default:
		return fmt.Errorf("unsupported build type: %s", pipeline.BuildType)
	}
}

func (eq *ExecutionQueue) stageDeploy(batch *model.ExecutionBatch, pipeline *model.PipelineConfig, workDir string) error {
	eq.logLine(batch.ID, "info", "=== Stage: deploy ===")

	switch pipeline.DeployType {
	case "docker":
		return eq.deployDocker(batch, pipeline, workDir)
	case "jar":
		return eq.deployJar(batch, workDir)
	case "war":
		eq.logLine(batch.ID, "warn", "WAR deploy is not implemented yet")
		return nil
	default:
		return fmt.Errorf("unsupported deploy type: %s", pipeline.DeployType)
	}
}

func (eq *ExecutionQueue) deployDocker(batch *model.ExecutionBatch, pipeline *model.PipelineConfig, workDir string) error {
	image := pipeline.DockerImage
	container := pipeline.DockerContainer
	if image == "" {
		return fmt.Errorf("docker_image is required for docker deploy")
	}
	if container == "" {
		return fmt.Errorf("docker_container is required for docker deploy")
	}

	eq.logLine(batch.ID, "info", fmt.Sprintf("docker build -t %s .", image))
	if err := eq.runCmd(batch.ID, "deploy", workDir, "docker", "build", "-t", image, "."); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	eq.logLine(batch.ID, "info", fmt.Sprintf("stopping old container: %s", container))
	_ = eq.runCmd(batch.ID, "deploy", workDir, "docker", "stop", container)
	_ = eq.runCmd(batch.ID, "deploy", workDir, "docker", "rm", container)

	runArgs := []string{"run", "--name", container}
	if pipeline.DockerRunArgs != "" {
		runArgs = append(runArgs, strings.Fields(pipeline.DockerRunArgs)...)
	}
	runArgs = append(runArgs, image)
	eq.logLine(batch.ID, "info", fmt.Sprintf("docker %s", strings.Join(runArgs, " ")))
	return eq.runCmd(batch.ID, "deploy", workDir, "docker", runArgs...)
}

func (eq *ExecutionQueue) deployJar(batch *model.ExecutionBatch, workDir string) error {
	eq.logLine(batch.ID, "info", "Searching jar artifact (*.jar)")
	jarPath, err := findFile(workDir, "*.jar")
	if err != nil {
		return err
	}
	eq.logLine(batch.ID, "info", fmt.Sprintf("java -jar %s", jarPath))
	return eq.runCmd(batch.ID, "deploy", workDir, "java", "-jar", jarPath)
}

func (eq *ExecutionQueue) runCmd(batchID, stage, dir, name string, args ...string) error {
	if eq.isCancelled(batchID) {
		return errBatchCancelled
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s failed: %w", name, err)
	}
	eq.setRunningCmd(batchID, cmd)
	defer eq.setRunningCmd(batchID, nil)

	var wg sync.WaitGroup
	stream := func(scanner *bufio.Scanner, level string) {
		defer wg.Done()
		for scanner.Scan() {
			eq.logLineStage(batchID, level, stage, scanner.Text())
		}
	}
	wg.Add(2)
	go stream(bufio.NewScanner(stdout), "info")
	go stream(bufio.NewScanner(stderr), "warn")
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if eq.isCancelled(batchID) {
			return errBatchCancelled
		}
		return fmt.Errorf("command failed [%s]: %w", name, err)
	}
	if eq.isCancelled(batchID) {
		return errBatchCancelled
	}
	return nil
}

func (eq *ExecutionQueue) runShell(batchID, stage, dir, shellCmd string) error {
	return eq.runCmd(batchID, stage, dir, "/bin/sh", "-c", shellCmd)
}

func (eq *ExecutionQueue) logLineStage(batchID, level, stage, line string) {
	db := database.GetDB()
	log.Printf("[%s][%s] %s: %s", batchID, stage, level, line)

	db.Create(&model.ExecutionLog{
		BatchID:  batchID,
		Stage:    stage,
		LogLine:  line,
		LogLevel: level,
		LineNo:   0,
	})

	eq.logMu.RLock()
	callback, ok := eq.logCallbacks[batchID]
	eq.logMu.RUnlock()
	if ok && callback != nil {
		callback(line, level)
	}
}

func (eq *ExecutionQueue) logLine(batchID, level, line string) {
	eq.logLineStage(batchID, level, "system", line)
}

func (eq *ExecutionQueue) createWorkDir(batch *model.ExecutionBatch) string {
	path := filepath.Join(eq.baseWorkDir, batch.SystemID, batch.PipelineID, batch.ID)
	if err := os.MkdirAll(path, 0755); err != nil {
		eq.logLine(batch.ID, "warn", fmt.Sprintf("Failed to create work directory: %v", err))
	}
	return path
}

func findFile(root, pattern string) (string, error) {
	var found string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || found != "" {
			return err
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			found = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("no file matching %s found under %s", pattern, root)
	}
	return found, nil
}

func toPtr(t time.Time) *time.Time {
	return &t
}

type bpmNode struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Action   string `json:"action"`
	TaskType string `json:"taskType"`
	Command  string `json:"command"`
	Steps    []bpmTaskStep `json:"steps"`
	PresetFields map[string]interface{} `json:"presetFields"`
	Order    int `json:"order"`
	RunMode  string `json:"runMode"`
	TriggerMode string `json:"triggerMode"`
	ParallelGroup string `json:"parallelGroup"`
}

type bpmTaskStep struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type bpmEdge struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition"`
}

type bpmGraph struct {
	Nodes []bpmNode `json:"nodes"`
	Edges []bpmEdge `json:"edges"`
}

func (eq *ExecutionQueue) executeByBPM(task *ExecutionTask, workDir string) error {
	bpmSvc := NewBPMDefinitionService()
	resp, err := bpmSvc.Get(task.UserID, task.Pipeline.PipelineID)
	if err != nil {
		return fmt.Errorf("load bpm definition failed: %w", err)
	}

	graph, err := parseBPMGraph(resp.Definition)
	if err != nil {
		return fmt.Errorf("parse bpm definition failed: %w", err)
	}

	return eq.executeBPMGraph(task, graph, workDir, task.StartNodeID)
}

func (eq *ExecutionQueue) executeBPMGraph(task *ExecutionTask, g *bpmGraph, workDir string, startNodeID string) error {
	if g == nil || len(g.Nodes) == 0 {
		return errNoExecutableBPMNodes
	}

	if len(g.Edges) == 0 {
		return eq.executeBPMByOrder(task, g.Nodes, workDir)
	}

	batch := task.Batch
	pipeline := task.Pipeline

	nodeByID := make(map[string]bpmNode, len(g.Nodes))
	inDegree := make(map[string]int, len(g.Nodes))
	out := make(map[string][]bpmEdge, len(g.Nodes))

	for _, n := range g.Nodes {
		nodeByID[n.ID] = n
		if _, ok := inDegree[n.ID]; !ok {
			inDegree[n.ID] = 0
		}
	}
	for _, e := range g.Edges {
		if _, ok := nodeByID[e.From]; !ok {
			continue
		}
		if _, ok := nodeByID[e.To]; !ok {
			continue
		}
		out[e.From] = append(out[e.From], e)
		inDegree[e.To]++
	}

	queue := make([]string, 0)
	if strings.TrimSpace(startNodeID) != "" {
		if _, ok := nodeByID[startNodeID]; ok {
			queue = append(queue, startNodeID)
			eq.logLine(batch.ID, "info", fmt.Sprintf("[BPM] 从指定节点重跑: %s", startNodeID))
		}
	}
	if len(queue) == 0 {
		for _, n := range g.Nodes {
			if strings.EqualFold(n.Type, "start") {
				queue = append(queue, n.ID)
			}
		}
	}
	if len(queue) == 0 {
		for _, n := range g.Nodes {
			if inDegree[n.ID] == 0 {
				queue = append(queue, n.ID)
			}
		}
	}
	if len(queue) == 0 {
		return fmt.Errorf("graph has no entry node")
	}

	visited := make(map[string]bool, len(g.Nodes))
	nodeStatus := make(map[string]string, len(g.Nodes))
	executedCount := 0

	// Initialize all non-start/end nodes to "pending"
	pendingNames := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		nt := strings.ToLower(strings.TrimSpace(n.Type))
		if nt != "start" && nt != "end" {
			pendingNames = append(pendingNames, n.Name)
		}
	}
	if len(pendingNames) > 0 {
		eq.initNodeStatuses(batch, pendingNames)
	}

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if visited[id] {
			continue
		}
		node, ok := nodeByID[id]
		if !ok {
			continue
		}
		visited[id] = true

		nodeType := strings.ToLower(strings.TrimSpace(node.Type))
		if nodeType != "start" && nodeType != "end" {
			executedCount++
			eq.logLine(batch.ID, "info", fmt.Sprintf("[BPM] 执行节点: %s (%s)", node.Name, node.Type))
			eq.setNodeStatus(batch, node.Name, "running")
			if err := eq.executeBPMNode(task, workDir, node); err != nil {
				nodeStatus[node.ID] = "failed"
				eq.setNodeStatus(batch, node.Name, "failed")
				return fmt.Errorf("node %s failed: %w", node.ID, err)
			}
			nodeStatus[node.ID] = "success"
			eq.setNodeStatus(batch, node.Name, "success")
		} else {
			nodeStatus[node.ID] = "success"
		}

		nextEdges := out[id]
		if len(nextEdges) == 0 {
			continue
		}

		if nodeType == "gateway" {
			selected, selReason := selectGatewayEdge(nextEdges, pipeline, nodeStatus)
			if selected != nil {
				eq.logLine(batch.ID, "info", fmt.Sprintf("[BPM] 网关选路: %s -> %s (%s)", id, selected.To, selReason))
				queue = append(queue, selected.To)
			}
			continue
		}

		for _, e := range nextEdges {
			queue = append(queue, e.To)
		}
	}

	if executedCount == 0 {
		return errNoExecutableBPMNodes
	}
	return nil
}

func selectGatewayEdge(edges []bpmEdge, pipeline *model.PipelineConfig, nodeStatus map[string]string) (*bpmEdge, string) {
	for i := range edges {
		e := &edges[i]
		cond := strings.TrimSpace(e.Condition)
		if cond == "" {
			continue
		}
		if matchesCondition(cond, pipeline, nodeStatus) {
			return e, fmt.Sprintf("condition matched: %s", cond)
		}
	}
	for i := range edges {
		e := &edges[i]
		if strings.TrimSpace(e.Condition) == "" {
			return e, "default branch"
		}
	}
	if len(edges) > 0 {
		return &edges[0], "fallback first branch"
	}
	return nil, "no branch"
}

func matchesCondition(condition string, pipeline *model.PipelineConfig, nodeStatus map[string]string) bool {
	cond := strings.TrimSpace(condition)
	if cond == "" {
		return true
	}
	if strings.Contains(cond, "&&") {
		parts := strings.Split(cond, "&&")
		for _, p := range parts {
			if !matchesCondition(strings.TrimSpace(p), pipeline, nodeStatus) {
				return false
			}
		}
		return true
	}
	if strings.Contains(cond, "==") {
		parts := strings.SplitN(cond, "==", 2)
		left := strings.TrimSpace(parts[0])
		right := trimQuotes(strings.TrimSpace(parts[1]))
		return readConditionValue(left, pipeline, nodeStatus) == right
	}
	if strings.Contains(cond, "!=") {
		parts := strings.SplitN(cond, "!=", 2)
		left := strings.TrimSpace(parts[0])
		right := trimQuotes(strings.TrimSpace(parts[1]))
		return readConditionValue(left, pipeline, nodeStatus) != right
	}
	return false
}

func readConditionValue(left string, pipeline *model.PipelineConfig, nodeStatus map[string]string) string {
	key := strings.ToLower(strings.TrimSpace(left))
	if strings.HasPrefix(key, "status(") && strings.HasSuffix(key, ")") {
		nodeID := strings.TrimSuffix(strings.TrimPrefix(key, "status("), ")")
		return nodeStatus[nodeID]
	}
	if strings.HasPrefix(key, "node.") && strings.HasSuffix(key, ".status") {
		nodeID := strings.TrimSuffix(strings.TrimPrefix(key, "node."), ".status")
		return nodeStatus[nodeID]
	}
	return readPipelineField(left, pipeline)
}

func trimQuotes(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "\"")
	v = strings.TrimSuffix(v, "\"")
	v = strings.TrimPrefix(v, "'")
	v = strings.TrimSuffix(v, "'")
	return v
}

func readPipelineField(field string, pipeline *model.PipelineConfig) string {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "build_type", "buildtype":
		return pipeline.BuildType
	case "deploy_type", "deploytype":
		return pipeline.DeployType
	case "repository_type", "repositorytype":
		return pipeline.RepositoryType
	case "branch":
		return pipeline.Branch
	default:
		return ""
	}
}

func parseBPMGraph(definition map[string]interface{}) (*bpmGraph, error) {
	body, err := json.Marshal(definition)
	if err != nil {
		return nil, err
	}
	var g bpmGraph
	if err := json.Unmarshal(body, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func topoSortBPM(g *bpmGraph) ([]bpmNode, error) {
	if g == nil || len(g.Nodes) == 0 {
		return nil, errNoExecutableBPMNodes
	}

	nodeByID := make(map[string]bpmNode, len(g.Nodes))
	inDegree := make(map[string]int, len(g.Nodes))
	out := make(map[string][]string, len(g.Nodes))

	for _, n := range g.Nodes {
		nodeByID[n.ID] = n
		if _, ok := inDegree[n.ID]; !ok {
			inDegree[n.ID] = 0
		}
	}

	for _, e := range g.Edges {
		if _, ok := nodeByID[e.From]; !ok {
			continue
		}
		if _, ok := nodeByID[e.To]; !ok {
			continue
		}
		out[e.From] = append(out[e.From], e.To)
		inDegree[e.To]++
	}

	queue := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		if inDegree[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}

	if len(queue) == 0 {
		return nil, fmt.Errorf("graph has no start node (all nodes have incoming edges)")
	}

	result := make([]bpmNode, 0, len(g.Nodes))
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		result = append(result, nodeByID[id])
		for _, next := range out[id] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(result) != len(g.Nodes) {
		return nil, fmt.Errorf("graph has cycle or disconnected invalid edges")
	}
	return result, nil
}

func (eq *ExecutionQueue) executeBPMNode(task *ExecutionTask, workDir string, node bpmNode) error {
	batch := task.Batch
	pipeline := task.Pipeline
	nodeType := strings.ToLower(strings.TrimSpace(node.Type))

	if strings.EqualFold(strings.TrimSpace(node.TriggerMode), "manual") {
		eq.logLine(batch.ID, "info", fmt.Sprintf("[BPM] 节点 %s 为手动触发，当前执行跳过", node.Name))
		return nil
	}

	switch nodeType {
	case "task":
		action := resolveTaskAction(node)
		switch action {
		case "source":
			return eq.stageSource(batch, eq.pipelineForSourceNode(task.UserID, pipeline, node), workDir)
		case "build":
			if pipeline.BuildType == "none" {
				eq.logLine(batch.ID, "info", "[BPM] build 节点跳过（build_type=none）")
				return nil
			}
			return eq.stageBuild(batch, pipeline, workDir)
		case "deploy":
			return eq.stageDeploy(batch, pipeline, workDir)
		default:
			return eq.executeNodeSteps(batch.ID, workDir, node, action)
		}
	case "approval":
		eq.logLine(batch.ID, "info", "[BPM] 审批节点（当前默认自动通过）")
		return nil
	case "agent":
		eq.logLine(batch.ID, "info", "[BPM] 智能体节点（当前为占位执行）")
		return nil
	case "gateway":
		eq.logLine(batch.ID, "info", "[BPM] 网关节点（当前按拓扑顺序继续）")
		return nil
	default:
		eq.logLine(batch.ID, "warn", fmt.Sprintf("[BPM] 未支持节点类型: %s，已跳过", node.Type))
		return nil
	}
}

func resolveTaskAction(node bpmNode) string {
	if node.Action != "" {
		return strings.ToLower(strings.TrimSpace(node.Action))
	}
	if node.TaskType != "" {
		return strings.ToLower(strings.TrimSpace(node.TaskType))
	}
	name := strings.ToLower(strings.TrimSpace(node.Name))
	if strings.Contains(name, "source") || strings.Contains(name, "git") || strings.Contains(name, "源码") {
		return "source"
	}
	if strings.Contains(name, "build") || strings.Contains(name, "maven") || strings.Contains(name, "gradle") || strings.Contains(name, "构建") {
		return "build"
	}
	if strings.Contains(name, "deploy") || strings.Contains(name, "docker") || strings.Contains(name, "部署") {
		return "deploy"
	}
	if strings.Contains(name, "测试") || strings.Contains(name, "scan") || strings.Contains(name, "扫描") || strings.Contains(name, "命令") || strings.Contains(name, "tool") || strings.Contains(name, "工具") {
		return "custom"
	}
	return ""
}

func (eq *ExecutionQueue) pipelineForSourceNode(userID uint, pipeline *model.PipelineConfig, node bpmNode) *model.PipelineConfig {
	if pipeline == nil || len(node.PresetFields) == 0 {
		return pipeline
	}

	merged := *pipeline
	if repoURL := stringField(node.PresetFields, "repoUrl", "repo_url"); repoURL != "" {
		merged.RepoURL = repoURL
	}
	if branch := stringField(node.PresetFields, "branch"); branch != "" {
		merged.Branch = branch
	}

	authType := strings.ToLower(strings.TrimSpace(stringField(node.PresetFields, "authType", "auth_type")))
	username, token := eq.resolveSourceCredentials(userID, node.PresetFields)
	if authType == "none" {
		merged.GitToken = ""
		merged.GitUsername = ""
	} else if authType == "token" {
		merged.GitToken = token
		if username != "" {
			merged.GitUsername = username
		}
		if merged.GitUsername == "" {
			merged.GitUsername = "git"
		}
	}

	return &merged
}

func stringField(values map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		default:
			text := strings.TrimSpace(fmt.Sprint(typed))
			if text != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}

func (eq *ExecutionQueue) resolveSourceCredentials(userID uint, values map[string]interface{}) (string, string) {
	svc := NewGlobalVariableService()
	namespaceKey := stringField(values, "gitCredentialKey", "git_credential_key")
	usernameField := stringField(values, "gitUsernameField", "git_username_field")
	tokenField := stringField(values, "gitTokenField", "git_token_field")

	username, err := svc.ResolveValue(userID, namespaceKey, usernameField)
	if err != nil && namespaceKey != "" && usernameField != "" {
		log.Printf("warning: resolve username %s.%s failed: %v", namespaceKey, usernameField, err)
	}
	token, err := svc.ResolveValue(userID, namespaceKey, tokenField)
	if err != nil && namespaceKey != "" && tokenField != "" {
		log.Printf("warning: resolve token %s.%s failed: %v", namespaceKey, tokenField, err)
	}
	return strings.TrimSpace(username), strings.TrimSpace(token)
}

func (eq *ExecutionQueue) executeNodeSteps(batchID string, workDir string, node bpmNode, action string) error {
	if strings.TrimSpace(node.Command) != "" {
		eq.logLine(batchID, "info", fmt.Sprintf("[BPM] 执行节点命令: %s", node.Command))
		return eq.runShell(batchID, "task", workDir, node.Command)
	}

	ran := false
	for _, step := range node.Steps {
		cmd := strings.TrimSpace(step.Command)
		if cmd == "" {
			continue
		}
		ran = true
		stepName := strings.TrimSpace(step.Name)
		if stepName == "" {
			stepName = "未命名步骤"
		}
		eq.logLine(batchID, "info", fmt.Sprintf("[BPM] 执行步骤: %s", stepName))
		if err := eq.runShell(batchID, "task", workDir, cmd); err != nil {
			return fmt.Errorf("step %s failed: %w", stepName, err)
		}
	}

	if ran {
		return nil
	}

	eq.logLine(batchID, "warn", fmt.Sprintf("[BPM] task action=%s 未匹配内置动作，且无可执行命令，按 no-op 处理", action))
	return nil
}

func (eq *ExecutionQueue) executeBPMByOrder(task *ExecutionTask, nodes []bpmNode, workDir string) error {
	batch := task.Batch

	items := make([]bpmNode, 0, len(nodes))
	for _, n := range nodes {
		t := strings.ToLower(strings.TrimSpace(n.Type))
		if t == "start" || t == "end" {
			continue
		}
		items = append(items, n)
	}
	if len(items) == 0 {
		return errNoExecutableBPMNodes
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Order == items[j].Order {
			return items[i].ID < items[j].ID
		}
		if items[i].Order == 0 {
			return false
		}
		if items[j].Order == 0 {
			return true
		}
		return items[i].Order < items[j].Order
	})

	// Initialize all nodes to "pending"
	pendingNames := make([]string, 0, len(items))
	for _, n := range items {
		pendingNames = append(pendingNames, n.Name)
	}
	eq.initNodeStatuses(batch, pendingNames)

	for i := 0; i < len(items); i++ {
		node := items[i]
		runMode := strings.ToLower(strings.TrimSpace(node.RunMode))
		if runMode != "parallel" {
			eq.logLine(batch.ID, "info", fmt.Sprintf("[BPM] 串行执行节点: %s", node.Name))
			eq.setNodeStatus(batch, node.Name, "running")
			if err := eq.executeBPMNode(task, workDir, node); err != nil {
				eq.setNodeStatus(batch, node.Name, "failed")
				return fmt.Errorf("node %s failed: %w", node.ID, err)
			}
			eq.setNodeStatus(batch, node.Name, "success")
			continue
		}

		groupKey := strings.TrimSpace(node.ParallelGroup)
		group := []bpmNode{node}
		for j := i + 1; j < len(items); j++ {
			if strings.ToLower(strings.TrimSpace(items[j].RunMode)) != "parallel" {
				break
			}
			nextGroupKey := strings.TrimSpace(items[j].ParallelGroup)
			if groupKey != "" && nextGroupKey != groupKey {
				break
			}
			if groupKey == "" && nextGroupKey != "" {
				break
			}
			group = append(group, items[j])
			i = j
		}

		if groupKey == "" {
			eq.logLine(batch.ID, "info", fmt.Sprintf("[BPM] 并行执行 %d 个节点（默认组）", len(group)))
		} else {
			eq.logLine(batch.ID, "info", fmt.Sprintf("[BPM] 并行执行组 %s，共 %d 个节点", groupKey, len(group)))
		}
		for _, n := range group {
			eq.setNodeStatus(batch, n.Name, "running")
		}
		errCh := make(chan error, len(group))
		var wg sync.WaitGroup
		for _, n := range group {
			stageNode := n
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := eq.executeBPMNode(task, workDir, stageNode); err != nil {
					eq.setNodeStatus(batch, stageNode.Name, "failed")
					errCh <- fmt.Errorf("node %s failed: %w", stageNode.ID, err)
				} else {
					eq.setNodeStatus(batch, stageNode.Name, "success")
				}
			}()
		}
		wg.Wait()
		close(errCh)
		if err, ok := <-errCh; ok {
			return err
		}
	}

	return nil
}

func (eq *ExecutionQueue) executeLegacyStages(task *ExecutionTask, workDir string) error {
	batch := task.Batch
	// Initialize all legacy stages to "pending"
	legacyNames := []string{"代码检出", "编译构建", "部署上线"}
	eq.initNodeStatuses(batch, legacyNames)

	eq.setNodeStatus(batch, "代码检出", "running")
	if err := eq.stageSource(batch, task.Pipeline, workDir); err != nil {
		eq.setNodeStatus(batch, "代码检出", "failed")
		return err
	}
	eq.setNodeStatus(batch, "代码检出", "success")

	if task.Pipeline.BuildType != "none" {
		eq.setNodeStatus(batch, "编译构建", "running")
		if err := eq.stageBuild(batch, task.Pipeline, workDir); err != nil {
			eq.setNodeStatus(batch, "编译构建", "failed")
			return err
		}
		eq.setNodeStatus(batch, "编译构建", "success")
	}

	eq.setNodeStatus(batch, "部署上线", "running")
	if err := eq.stageDeploy(batch, task.Pipeline, workDir); err != nil {
		eq.setNodeStatus(batch, "部署上线", "failed")
		return err
	}
	eq.setNodeStatus(batch, "部署上线", "success")
	return nil
}

// GetQueueStats returns current queue statistics
func (eq *ExecutionQueue) GetQueueStats() map[string]interface{} {
	return map[string]interface{}{
		"max_workers":      eq.maxWorkers,
		"queued_tasks":     len(eq.taskChan),
		"active_callbacks": len(eq.logCallbacks),
	}
}
