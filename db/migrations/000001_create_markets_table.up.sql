CREATE TABLE waypoints (
    symbol TEXT NOT NULL,
    type TEXT NOT NULL,
    x TEXT NOT NULL,
    y TEXT NOT NULL,

    PRIMARY KEY (symbol)
);

CREATE TABLE goods (
    symbol TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,

    PRIMARY KEY (symbol)
);

CREATE TABLE markets (
    waypoint TEXT NOT NULL,
    good TEXT NOT NULL,
    type TEXT NOT NULL,
    volume INTEGER,
    activity INTEGER,
    bid INTEGER,
    ask INTEGER,

    PRIMARY KEY (waypoint, good, type),

    FOREIGN KEY(waypoint) REFERENCES waypoints(symbol),
    FOREIGN KEY(good) REFERENCES goods(symbol)
);

