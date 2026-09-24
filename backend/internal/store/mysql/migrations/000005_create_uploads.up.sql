CREATE TABLE uploads (
  id            CHAR(36) PRIMARY KEY,
  object_key    VARCHAR(512) NOT NULL,
  content_type  VARCHAR(100) NOT NULL,
  size_bytes    INT NOT NULL,
  created_by    CHAR(36),
  created_at    DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB;
