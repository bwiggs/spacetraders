CREATE TABLE fleet (
    symbol TEXT NOT NULL,
    data jsonb NOT NULL,

    PRIMARY KEY (symbol)
);