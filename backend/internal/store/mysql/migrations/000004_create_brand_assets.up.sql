CREATE TABLE brand_assets (
  id          CHAR(36) PRIMARY KEY,
  kind        VARCHAR(20) NOT NULL,
  object_key  VARCHAR(512) NOT NULL,
  meta        JSON NOT NULL,
  version     INT NOT NULL,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  INDEX idx_brand_assets_kind_active (kind, is_active)
) ENGINE=InnoDB;
