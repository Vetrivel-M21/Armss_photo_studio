CREATE TABLE idempotency_keys (
  `key`             VARCHAR(255) PRIMARY KEY,
  project_id        CHAR(36) NOT NULL,
  endpoint          VARCHAR(50) NOT NULL,
  response_snapshot JSON NOT NULL,
  created_at        DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB;
