CREATE TABLE IF NOT EXISTS F(
    FKEY INT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS D(
    DKEY INT PRIMARY KEY,
    BREF INT
);

CREATE TABLE IF NOT EXISTS E(
    EKEY INT PRIMARY KEY,
    DREF INT,
    FREF INT,
    FOREIGN KEY (DREF) REFERENCES D,
    FOREIGN KEY (FREF) REFERENCES F
);

CREATE TABLE IF NOT EXISTS C(
    CKEY INT PRIMARY KEY,
    EREF INT,
    FOREIGN KEY (EREF) REFERENCES E
);

CREATE TABLE IF NOT EXISTS B(
    BKEY INT PRIMARY KEY,
    CREF INT,
    FOREIGN KEY (CREF) REFERENCES C
);

CREATE TABLE IF NOT EXISTS A(
    AKEY INT PRIMARY KEY,
    BREF INT,
    FOREIGN KEY (BREF) REFERENCES B
)

