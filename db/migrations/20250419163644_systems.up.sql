CREATE TABLE systems (
    symbol TEXT NOT NULL,
    constellation TEXT NOT NULL,
    name TEXT NOT NULL,
    sector_symbol TEXT NOT NULL,
    type TEXT NOT NULL,
    x INT NOT NULL,
    y INT NOT NULL,

    PRIMARY KEY (symbol)
);