ALTER TABLE projects
    ADD COLUMN aspect_ratio    VARCHAR(10)  NOT NULL DEFAULT '3:4'        AFTER footer_text,
    ADD COLUMN logo_position   VARCHAR(30)  NOT NULL DEFAULT 'top-center'  AFTER aspect_ratio,
    ADD COLUMN logo_scale      VARCHAR(20)  NOT NULL DEFAULT 'standard'    AFTER logo_position,
    ADD COLUMN slogan_position VARCHAR(30)  NOT NULL DEFAULT 'below-logo'  AFTER logo_scale,
    ADD COLUMN slogan_scale    VARCHAR(20)  NOT NULL DEFAULT 'standard'    AFTER slogan_position,
    ADD COLUMN footer_visible  TINYINT(1)   NOT NULL DEFAULT 1             AFTER slogan_scale;
