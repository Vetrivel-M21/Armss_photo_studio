CREATE TABLE stage_outputs (
  id          CHAR(36) PRIMARY KEY,
  project_id  CHAR(36) NOT NULL,
  stage       VARCHAR(30) NOT NULL,
  payload     JSON NOT NULL,
  model       VARCHAR(100),
  tokens_in   INT,
  tokens_out  INT,
  superseded  BOOLEAN NOT NULL DEFAULT FALSE,
  created_at  DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
  INDEX idx_stage_outputs_lookup (project_id, stage, created_at DESC)
) ENGINE=InnoDB;
