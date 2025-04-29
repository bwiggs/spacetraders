CREATE TABLE agents (
    symbol TEXT NOT NULL,
    headquarters TEXT NOT NULL REFERENCES systems(symbol),
    credits BIGINT NOT NULL,
    faction TEXT NOT NULL,
    shipCount INT NOT NULL,

    json jsonb NOT NULL,

    PRIMARY KEY (symbol)
);