CREATE TABLE waypoints_traits (
    waypoint TEXT NOT NULL,
    trait TEXT NOT NULL,
    
    PRIMARY KEY (waypoint, trait)
    
    FOREIGN KEY(waypoint) REFERENCES waypoints(symbol)
    FOREIGN KEY(trait) REFERENCES traits(symbol)
);

CREATE TABLE traits (
    symbol TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    
    PRIMARY KEY (symbol)
);
