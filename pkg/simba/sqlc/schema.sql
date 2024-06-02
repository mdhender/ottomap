PRAGMA foreign_keys = ON;

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;

-- createtime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
-- updatetime TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

CREATE TABLE users
(
    uid             TEXT NOT NULL, -- uuid
    username        TEXT NOT NULL, -- forced to lowercase
    email           TEXT NOT NULL, -- forced to lowercase
    hashed_password TEXT NOT NULL, -- hashed and hex-encoded
    clan            TEXT NOT NULL, -- clan controlled by user
    PRIMARY KEY (uid),
    UNIQUE (username),
    UNIQUE (email)
);

CREATE TABLE roles
(
    rid TEXT NOT NULL, -- forced to lowercase
    PRIMARY KEY (rid)
);

CREATE TABLE users_roles
(
    uid   TEXT NOT NULL,                -- users.uid
    rid   TEXT NOT NULL,                -- roles.rid
    value TEXT NOT NULL DEFAULT 'true', -- any value, defaults to true
    PRIMARY KEY (uid, rid),
    FOREIGN KEY (uid) REFERENCES users (uid) ON DELETE CASCADE,
    FOREIGN KEY (rid) REFERENCES roles (rid) ON DELETE CASCADE

);

CREATE TABLE sessions
(
    sid        TEXT      NOT NULL, -- uuid
    uid        TEXT      NOT NULL, -- users.uid
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (sid),
    FOREIGN KEY (uid) REFERENCES users (uid) ON DELETE CASCADE
);

INSERT INTO users (uid, username, email, hashed_password, clan)
VALUES ('ad141219-8db3-4544-9513-67715b08f0c2',
        'ottomap',
        'ottomap@example.com',
        '$$',
        '0991'
       );

INSERT INTO sessions (sid, uid, expires_at)
VALUES ('ad141219-8db3-4544-9513-67715b08f0c2',
        'ad141219-8db3-4544-9513-67715b08f0c2',
        DATETIME('now', '+14 days'));

INSERT INTO roles (rid)
VALUES ('anonymous');
INSERT INTO roles (rid)
VALUES ('administrator');
INSERT INTO roles (rid)
VALUES ('operator');
INSERT INTO roles (rid)
VALUES ('service');
INSERT INTO roles (rid)
VALUES ('user');
INSERT INTO roles (rid)
VALUES ('clans');
INSERT INTO roles (rid)
VALUES ('authenticated');

INSERT INTO users_roles (uid, rid)
VALUES ('ad141219-8db3-4544-9513-67715b08f0c2', 'user');
INSERT INTO users_roles (uid, rid, value)
VALUES ('ad141219-8db3-4544-9513-67715b08f0c2', 'clans', '0991');