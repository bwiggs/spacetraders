CREATE TABLE IF NOT EXISTS ships (
    type TEXT NOT NULL,
    name TEXT,
    description TEXT,
    
    PRIMARY KEY (type)
);

CREATE TABLE IF NOT EXISTS shipyards (
    waypoint TEXT NOT NULL,
    ship TEXT NOT NULL,
    supply TEXT,
    bid DECIMAL,
    
    PRIMARY KEY (waypoint, ship)
    
    FOREIGN KEY(waypoint) REFERENCES waypoints(symbol)
    FOREIGN KEY(ship) REFERENCES ships(type)
);


