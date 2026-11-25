ALTER TABLE eventos ADD COLUMN cuando timestamptz NOT NULL DEFAULT now();
