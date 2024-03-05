CREATE TABLE waypoints (
    symbol TEXT NOT NULL,
    type TEXT NOT NULL,
    x TEXT NOT NULL,
    y TEXT NOT NULL,
    is_market boolean NOT NULL,
    is_shipyard boolean NOT NULL,
    PRIMARY KEY (symbol)
);

CREATE TABLE markets (
    waypoint TEXT NOT NULL,
    good TEXT NOT NULL,
    type TEXT NOT NULL,
    PRIMARY KEY (waypoint, good, type)
);
