CREATE TABLE sessions (
  id          VARCHAR(64) PRIMARY KEY,
  user_id     CHAR(36) NOT NULL,
  expires_at  DATETIME(6) NOT NULL,
  created_at  DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_sessions_expires_at (expires_at)
) ENGINE=InnoDB;
