CREATE TABLE IF NOT EXISTS user_limits (
  id                        CHAR(36) PRIMARY KEY,
  user_id                   CHAR(36) UNIQUE NULL,
  max_api_calls_per_hour    INT NOT NULL DEFAULT 30,
  max_renders_per_hour      INT NOT NULL DEFAULT 6,
  monthly_spend_cap_cents   INT NOT NULL DEFAULT 1500,
  is_unlimited              BOOLEAN NOT NULL DEFAULT FALSE,
  created_at                DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at                DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS api_rate_records (
  id          BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id     CHAR(36) NOT NULL,
  action_type VARCHAR(30) NOT NULL,
  created_at  DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  INDEX idx_user_window (user_id, action_type, created_at)
) ENGINE=InnoDB;
