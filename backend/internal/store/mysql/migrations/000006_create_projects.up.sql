CREATE TABLE projects (
  id           CHAR(36) PRIMARY KEY,
  user_id      CHAR(36) NOT NULL,
  title        VARCHAR(255),
  input_text   TEXT NOT NULL,
  template_id  CHAR(36),
  state        VARCHAR(30) NOT NULL DEFAULT 'draft',
  created_at   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (template_id) REFERENCES templates(id),
  CHECK (state IN ('draft','briefing','concepts_ready','developing','render_ready','composited','approved')),
  INDEX idx_projects_user (user_id, created_at DESC)
) ENGINE=InnoDB;
