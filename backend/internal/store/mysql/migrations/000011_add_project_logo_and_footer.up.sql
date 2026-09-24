ALTER TABLE projects
  ADD COLUMN logo_object_key VARCHAR(512) NULL AFTER template_id,
  ADD COLUMN footer_text     VARCHAR(255) NULL AFTER logo_object_key;
