-- foreign keys must be disabled to drop tables with foreign keys
PRAGMA foreign_keys = OFF;

DROP TABLE IF EXISTS clans;
DROP TABLE IF EXISTS element_metadata;
DROP TABLE IF EXISTS elements;
DROP TABLE IF EXISTS elements_parents;
DROP TABLE IF EXISTS input;
DROP TABLE IF EXISTS input_lines;
DROP TABLE IF EXISTS message_log;
DROP TABLE IF EXISTS metadata;
DROP TABLE IF EXISTS report_queue;
DROP TABLE IF EXISTS report_queue_data;
DROP TABLE IF EXISTS report_sections;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS turns;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS users_roles;

-- foreign keys must be enabled with every database connection
PRAGMA foreign_keys = ON;

-- metadata stores information about this schema file
-- and the paths to files needed by the server.
CREATE TABLE metadata
(
    version        TEXT NOT NULL, -- version of this schema file
    input_path     TEXT NOT NULL, -- absolute path to the input directory
    output_path    TEXT NOT NULL, -- absolute path to the output directory
    public_path    TEXT NOT NULL, -- absolute path to the public directory
    templates_path TEXT NOT NULL  -- absolute path to the templates directory
);

INSERT INTO metadata (version, input_path, output_path, public_path, templates_path)
VALUES ('0.0.1', 'data/input', 'data/output', '', '');

-- input stores data for every input file uploaded
CREATE TABLE input
(
    id     INTEGER PRIMARY KEY,
    status TEXT      NOT NULL DEFAULT 'pending',         -- status of the input file
    path   TEXT      NOT NULL,                           -- absolute path to the input file
    name   TEXT      NOT NULL,                           -- file name, extracted from input path
    cksum  TEXT      NOT NULL,                           -- SHA-256 checksum of the file
    crdttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was created
    updttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was last updated
    UNIQUE (cksum)
);

CREATE TABLE input_lines
(
    iid     INTEGER NOT NULL, -- input id
    sect_no INTEGER NOT NULL, -- original input section
    line_no INTEGER NOT NULL, -- original input line
    line    TEXT    NOT NULL,
    PRIMARY KEY (iid, sect_no, line_no),
    FOREIGN KEY (iid) REFERENCES input (id) ON DELETE CASCADE
);


-- users stores information about the end-users (the players)
CREATE TABLE users
(
    uid             TEXT NOT NULL, -- uuid
    username        TEXT NOT NULL, -- forced to lowercase
    email           TEXT NOT NULL, -- forced to lowercase
    hashed_password TEXT NOT NULL, -- hashed and hex-encoded
    PRIMARY KEY (uid),
    UNIQUE (username),
    UNIQUE (email)
);

-- roles defines role names for the policy agent
CREATE TABLE roles
(
    rlid TEXT NOT NULL, -- forced to lowercase
    PRIMARY KEY (rlid)
);

-- users_roles stores the roles currently assigned to a user
CREATE TABLE users_roles
(
    uid   TEXT NOT NULL,                -- user id
    rlid  TEXT NOT NULL,                -- role id
    value TEXT NOT NULL DEFAULT 'true', -- any value, defaults to true
    PRIMARY KEY (uid, rlid),
    FOREIGN KEY (uid) REFERENCES users (uid) ON DELETE CASCADE,
    FOREIGN KEY (rlid) REFERENCES roles (rlid) ON DELETE CASCADE

);

-- sessions stores all of the active sessions.
-- it is expected that a scheduler runs a process to remove expired sessions.
CREATE TABLE sessions
(
    sid        TEXT      NOT NULL, -- uuid
    uid        TEXT      NOT NULL, -- user id
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (sid),
    FOREIGN KEY (uid) REFERENCES users (uid) ON DELETE CASCADE
);

-- turns stores turn data from all the reports uploaded
CREATE TABLE turns
(
    tid    TEXT      NOT NULL,                           -- formatted as 0899-12
    turn   TEXT      NOT NULL,                           -- formatted as 899-12
    year   INTEGER   NOT NULL,
    month  INTEGER   NOT NULL,
    crdttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was created
    PRIMARY KEY (tid),
    UNIQUE (turn),
    UNIQUE (year, month)
);

-- clans stores data about each clan in the game
CREATE TABLE clans
(
    cid TEXT NOT NULL, -- clan id, formatted as 0138
    uid TEXT NOT NULL, -- uuid
    PRIMARY KEY (cid),
    UNIQUE (uid),
    FOREIGN KEY (uid) REFERENCES users (uid) ON DELETE CASCADE
);


-- elements stores data about all the elements (units) from all the reports uploaded
-- kind can be Courier, Element, Fleet, Garrison, or Tribe
CREATE TABLE elements
(
    cid    TEXT      NOT NULL,                           -- clan id
    eid    TEXT      NOT NULL,                           -- element id, formatted as 1138e2
    kind   TEXT      NOT NULL,                           -- see above for possible values
    crdttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was created
    PRIMARY KEY (eid),
    FOREIGN KEY (cid) REFERENCES clans (cid) ON DELETE CASCADE
);

-- elements_parents is used to store the parent of an element
CREATE TABLE elements_parents
(
    eid        TEXT NOT NULL, -- element id
    parent_eid TEXT NOT NULL, -- parent element id
    PRIMARY KEY (eid, parent_eid),
    FOREIGN KEY (eid) REFERENCES elements (eid) ON DELETE CASCADE,
    FOREIGN KEY (parent_eid) REFERENCES elements (eid) ON DELETE CASCADE
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
CREATE TABLE element_metadata
(
    eid       TEXT      NOT NULL,                           -- element id
    tid       TEXT      NOT NULL,                           -- turn id
    action_cd TEXT      NOT NULL,                           -- see above for possible values
    hex       TEXT      NOT NULL,                           -- formatted as GG CCRR or "N/A"
    crdttm    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was created
    PRIMARY KEY (eid, tid),
    FOREIGN KEY (eid) REFERENCES elements (eid) ON DELETE CASCADE,
    FOREIGN KEY (tid) REFERENCES turns (tid) ON DELETE CASCADE
);

-- reports stores data from all of the reports uploaded
CREATE TABLE reports
(
    tid    TEXT      NOT NULL,                           -- turn id
    cid    TEXT      NOT NULL,                           -- clan id
    rid    TEXT      NOT NULL,                           -- formatted as 0899-12.0138
    crdttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was created
    PRIMARY KEY (rid),
    FOREIGN KEY (tid) REFERENCES turns (tid) ON DELETE CASCADE,
    FOREIGN KEY (cid) REFERENCES clans (cid) ON DELETE CASCADE,
    UNIQUE (tid, cid)
);

-- report_sections stores the lines from each section in a report
CREATE TABLE report_sections
(
    rid    TEXT      NOT NULL,                           -- report id
    eid    TEXT      NOT NULL,                           -- element id
    lines  TEXT      NOT NULL,                           -- all the lines in the report section
    crdttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was created
    PRIMARY KEY (rid, eid),
    FOREIGN KEY (rid) REFERENCES reports (rid) ON DELETE CASCADE,
    FOREIGN KEY (eid) REFERENCES elements (eid) ON DELETE CASCADE
);

-- report_queue stores data needed to process the turn reports
CREATE TABLE report_queue
(
    qid    TEXT      NOT NULL,                           -- unique id, generated by file uploader
    cid    TEXT      NOT NULL,                           -- id of the clan uploading the report
    status TEXT      NOT NULL,                           -- current status of the request
    crdttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was created
    updttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was last updated
    PRIMARY KEY (qid),
    FOREIGN KEY (cid) REFERENCES clans (cid) ON DELETE CASCADE
);

-- report_queue_data stores the unprocessed report data
CREATE TABLE report_queue_data
(
    qid    TEXT      NOT NULL,
    name   TEXT      NOT NULL,                           -- local file name; unsafe text, should whitelist
    cksum  TEXT      NOT NULL,                           -- generated by the file upload function
    lines  TEXT      NOT NULL,
    crdttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was created
    PRIMARY KEY (qid),
    UNIQUE (cksum),
    FOREIGN KEY (qid) REFERENCES report_queue (qid) ON DELETE CASCADE
);

CREATE TABLE log_messages
(
    id      INTEGER PRIMARY KEY,                         -- unique id for the log message
    arg_1   TEXT      NOT NULL,
    arg_2   TEXT      NOT NULL,
    arg_3   TEXT      NOT NULL,
    message TEXT      NOT NULL,
    crdttm  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP -- when the row was created
);
