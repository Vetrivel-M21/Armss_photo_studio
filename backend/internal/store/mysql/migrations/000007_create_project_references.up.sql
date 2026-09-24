CREATE TABLE project_references (
  project_id    CHAR(36) NOT NULL,
  object_key    VARCHAR(512) NOT NULL,
  style_profile JSON,
  PRIMARY KEY (project_id, object_key),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB;
