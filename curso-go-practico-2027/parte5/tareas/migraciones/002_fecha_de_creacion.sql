ALTER TABLE tareas ADD COLUMN creada timestamptz NOT NULL DEFAULT now();
