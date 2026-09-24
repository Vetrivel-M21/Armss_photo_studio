CREATE TABLE renders (
  id          CHAR(36) PRIMARY KEY,
  project_id  CHAR(36) NOT NULL,
  kind        VARCHAR(20) NOT NULL,
  provider    VARCHAR(50),
  object_key  VARCHAR(512) NOT NULL,
  prompt      TEXT,
  cost_cents  INT,
  meta        JSON,
  created_at  DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  INDEX idx_renders_lookup (project_id, kind, created_at DESC)
) ENGINE=InnoDB;
