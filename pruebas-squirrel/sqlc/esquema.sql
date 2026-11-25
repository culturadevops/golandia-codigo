CREATE TABLE productos (
    id        bigint PRIMARY KEY,
    nombre    text    NOT NULL,
    categoria text    NOT NULL,
    precio    numeric(10,2) NOT NULL,
    stock     int     NOT NULL
);
