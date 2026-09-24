ALTER TABLE agent_settings
  ADD COLUMN default_image_provider VARCHAR(64) NOT NULL DEFAULT 'flux',
  ADD COLUMN flux_api_key VARCHAR(255) NULL,
  ADD COLUMN ideogram_api_key VARCHAR(255) NULL;
