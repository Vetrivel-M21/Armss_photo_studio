ALTER TABLE agent_settings
  ADD COLUMN anthropic_api_key VARCHAR(255) NULL,
  ADD COLUMN gemini_api_key VARCHAR(255) NULL,
  ADD COLUMN openai_api_key VARCHAR(255) NULL;
