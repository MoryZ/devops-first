CREATE TABLE IF NOT EXISTS execution_batches (
  id VARCHAR(64) PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  system_id VARCHAR(128) NOT NULL,
  pipeline_id VARCHAR(128) NOT NULL,
  pipeline_name VARCHAR(255) NOT NULL,
  batch_number INT NOT NULL,
  status VARCHAR(32) NOT NULL,
  triggered_by VARCHAR(32) NOT NULL,
  work_dir VARCHAR(1024) DEFAULT '',
  started_at DATETIME NULL,
  completed_at DATETIME NULL,
  total_duration BIGINT NOT NULL DEFAULT 0,
  error_message TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  deleted_at DATETIME NULL,
  INDEX idx_user_batch (user_id),
  INDEX idx_system_batch (system_id),
  INDEX idx_pipeline_batch (pipeline_id),
  INDEX idx_status (status),
  INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS execution_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  batch_id VARCHAR(64) NOT NULL,
  stage VARCHAR(64) NOT NULL,
  log_line TEXT NOT NULL,
  log_level VARCHAR(16) NOT NULL,
  line_no INT NOT NULL,
  created_at DATETIME NOT NULL,
  INDEX idx_batch_log (batch_id),
  INDEX idx_stage (stage)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
