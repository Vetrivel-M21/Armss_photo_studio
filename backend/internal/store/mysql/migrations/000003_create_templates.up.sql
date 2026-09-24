CREATE TABLE templates (
  id            CHAR(36) PRIMARY KEY,
  title         VARCHAR(255) NOT NULL,
  object_key    VARCHAR(512) NOT NULL,
  thumb_key     VARCHAR(512),
  occasion_tags JSON NOT NULL,
  season        VARCHAR(50),
  style_profile JSON,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_by    CHAR(36),
  created_at    DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
  INDEX idx_templates_active_season (is_active, season)
) ENGINE=InnoDB;
