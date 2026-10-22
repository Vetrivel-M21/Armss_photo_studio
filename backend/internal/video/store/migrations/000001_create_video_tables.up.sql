CREATE TABLE video_projects (
  id           CHAR(36)     PRIMARY KEY,
  user_id      CHAR(36)     NOT NULL,
  title        VARCHAR(255) NOT NULL DEFAULT 'Untitled Video',
  spec         JSON         NOT NULL,
  spec_version INT          NOT NULL DEFAULT 1,
  revision     INT          NOT NULL DEFAULT 1,
  created_at   DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at   DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_video_projects_user_updated (user_id, updated_at DESC)
) ENGINE=InnoDB;

CREATE TABLE video_jobs (
  id           CHAR(36)     PRIMARY KEY,
  project_id   CHAR(36)     NULL,
  user_id      CHAR(36)     NOT NULL,
  spec         JSON         NOT NULL,
  resolution   VARCHAR(10)  NOT NULL DEFAULT '1080p',
  status       VARCHAR(20)  NOT NULL DEFAULT 'queued',
  progress     TINYINT      NOT NULL DEFAULT 0,
  output_key   VARCHAR(512) NULL,
  error        TEXT         NULL,
  started_at   DATETIME(6)  NULL,
  finished_at  DATETIME(6)  NULL,
  created_at   DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at   DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  FOREIGN KEY (project_id) REFERENCES video_projects(id) ON DELETE SET NULL,
  FOREIGN KEY (user_id)    REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_video_jobs_user_created (user_id, created_at DESC),
  INDEX idx_video_jobs_status       (status, updated_at)
) ENGINE=InnoDB;

CREATE TABLE video_audio_presets (
  id           CHAR(36)     PRIMARY KEY,
  title        VARCHAR(255) NOT NULL,
  artist       VARCHAR(255) NULL,
  s3_key       VARCHAR(512) NOT NULL,
  duration     FLOAT        NOT NULL,
  license      VARCHAR(100) NOT NULL,
  source_url   VARCHAR(1024) NOT NULL,
  attribution  TEXT         NULL,
  created_at   DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB;
