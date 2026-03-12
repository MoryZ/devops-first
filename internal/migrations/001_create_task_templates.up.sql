-- Create task_templates table
CREATE TABLE IF NOT EXISTS task_templates (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL COMMENT 'Template name, e.g., Java构建',
    category VARCHAR(100) NOT NULL COMMENT 'Category, e.g., 代码扫描, 构建, 部署',
    sub_category VARCHAR(100) COMMENT 'Sub-category, e.g., Java代码扫描',
    description TEXT,
    preset_fields JSON COMMENT 'Preset fields like buildCluster, buildNode, timeout, etc.',
    advanced_settings JSON COMMENT 'Advanced settings like debugMode, deployTask, dockerDaemon',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_user_id (user_id),
    KEY idx_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create task_template_steps table
CREATE TABLE IF NOT EXISTS task_template_steps (
    id VARCHAR(36) PRIMARY KEY,
    template_id VARCHAR(36) NOT NULL,
    step_order INT NOT NULL COMMENT 'Step order in template',
    name VARCHAR(255) NOT NULL COMMENT 'Step name, e.g., 配置MavenSettings文件',
    command TEXT COMMENT 'Command to execute',
    shell_specified BOOLEAN DEFAULT FALSE COMMENT 'Whether to run in shell',
    envs JSON COMMENT 'Environment variables',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES task_templates(id) ON DELETE CASCADE,
    KEY idx_template_id (template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create task_template_plugins table (for plugin extensions)
CREATE TABLE IF NOT EXISTS task_template_plugins (
    id VARCHAR(36) PRIMARY KEY,
    template_id VARCHAR(36) NOT NULL,
    plugin_name VARCHAR(255) NOT NULL COMMENT 'Plugin name, e.g., SonarQube',
    plugin_config JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES task_templates(id) ON DELETE CASCADE,
    KEY idx_template_id (template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
