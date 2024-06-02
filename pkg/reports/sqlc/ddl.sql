DROP TABLE IF EXISTS report_sections;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS unit_metadata;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS clans;
DROP TABLE IF EXISTS users;

-- createtime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
-- updatetime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

CREATE TABLE users
(
    uid             TEXT NOT NULL, -- uuid
    username        TEXT NOT NULL, -- forced to lowercase
    email           TEXT NOT NULL, -- forced to lowercase
    hashed_password TEXT,
    PRIMARY KEY (uid),
    UNIQUE (username),
    UNIQUE (email)
);

INSERT INTO users (uid, username, email)
VALUES ('ad141219-8db3-4544-9513-67715b08f0c2', 'ottomap', 'ottomap@example.com');

CREATE TABLE clans
(
    clan_id           TEXT NOT NULL, -- formatted as 0138
    controlled_by_uid TEXT NOT NULL, -- user that controls the clan
    PRIMARY KEY (clan_id)
);

INSERT INTO clans (clan_id, controlled_by_uid)
VALUES ('0138', 'ad141219-8db3-4544-9513-67715b08f0c2');

-- kind can be Courier, Element, Fleet, Garrison, or Tribe
CREATE TABLE units
(
    unit_id   TEXT NOT NULL, -- formatted as 1138e2
    kind      TEXT NOT NULL, -- see above for possible values
    clan_id   TEXT NOT NULL, -- formatted as 0138
    parent_id TEXT NOT NULL, -- parent unit, eg 0138
    PRIMARY KEY (unit_id)
);

-- action_cd can be
--   AAM -- Absorbed  After  Movement
--   ABM -- Absorbed  Before Movement
--   CAM -- Created   After  Movement
--   CBM -- Created   Before Movement
--   DAM -- Destroyed After  Movement
--   DBM -- Destroyed Before Movement
-- hex is the hex the unit was in when the action occurred. it is either
-- the "origin" hex or the hex the unit was in when it was destroyed.
CREATE TABLE unit_metadata
(
    unit_id   TEXT NOT NULL, -- formatted as 0138e2
    turn_id   TEXT NOT NULL, -- turn id, eg 899-12
    action_cd TEXT NOT NULL, -- see above for possible values
    hex       TEXT NOT NULL, -- formatted as GG CCRR or "N/A"
    PRIMARY KEY (unit_id, turn_id)
);

CREATE TABLE reports
(
    rpt_id  TEXT    NOT NULL, -- formatted as 899-12.0138
    turn    TEXT    NOT NULL, -- formatted as 899-12
    year    INTEGER NOT NULL,
    month   INTEGER NOT NULL,
    clan_id TEXT    NOT NULL, -- formatted as 0138
    PRIMARY KEY (rpt_id),
    UNIQUE (turn, year, month, clan_id)
);

CREATE TABLE report_sections
(
    rpt_id  TEXT NOT NULL, -- formatted as 899-12.0138
    unit_id TEXT NOT NULL, -- formatted as 0138e2
    lines   TEXT NOT NULL, -- all the lines in the report section
    PRIMARY KEY (rpt_id, unit_id)
);

