CREATE TABLE eventos (
    id       bigserial PRIMARY KEY,
    servidor text NOT NULL,
    nivel    text NOT NULL,
    mensaje  text NOT NULL
);
