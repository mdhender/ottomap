--  Copyright (c) 2024 Michael D Henderson. All rights reserved.

-- foreign keys must be disabled to drop tables with foreign keys
PRAGMA foreign_keys = OFF;

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

-- foreign keys must be enabled with every database connection
PRAGMA foreign_keys = ON;

-- crdttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was created
-- updttm TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was last updated
-- FOREIGN KEY (iid) REFERENCES input (id) ON DELETE CASCADE

CREATE TABLE users
(
    uid             INTEGER PRIMARY KEY,
    handle          TEXT      NOT NULL,
    hashed_password TEXT      NOT NULL,                           -- bcrypt hash of the password
    clan            TEXT      NOT NULL,
    magic_key       TEXT      NOT NULL,
    path            TEXT      NOT NULL,
    crdttm          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- when the row was created
    updttm          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP  -- when the row was last updated
);

CREATE UNIQUE INDEX IF NOT EXISTS users_clan_index ON users (clan);
CREATE UNIQUE INDEX IF NOT EXISTS users_handle_index ON users (handle);
CREATE UNIQUE INDEX IF NOT EXISTS users_magic_key_index ON users (magic_key);

CREATE TABLE sessions
(
    sid          TEXT      NOT NULL,
    uid          INTEGER   NOT NULL,
    expires_dttm TIMESTAMP NOT NULL, -- when the session will expire
    PRIMARY KEY (sid),
    FOREIGN KEY (uid) REFERENCES users (uid) ON DELETE CASCADE
);

