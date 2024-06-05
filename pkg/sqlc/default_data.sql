-- foreign keys must be enabled with every database connection
PRAGMA foreign_keys = ON;

INSERT INTO roles (rlid)
VALUES ('anonymous');
INSERT INTO roles (rlid)
VALUES ('administrator');
INSERT INTO roles (rlid)
VALUES ('operator');
INSERT INTO roles (rlid)
VALUES ('service');
INSERT INTO roles (rlid)
VALUES ('user');
INSERT INTO roles (rlid)
VALUES ('clans');
INSERT INTO roles (rlid)
VALUES ('authenticated');

INSERT INTO users (uid, username, email, hashed_password)
VALUES ('ad141219-8db3-4544-9513-67715b08f0c2',
        'ottomap',
        'ottomap@example.com',
        '$$');


INSERT INTO users_roles (uid, rlid)
VALUES ('ad141219-8db3-4544-9513-67715b08f0c2', 'user');
INSERT INTO users_roles (uid, rlid, value)
VALUES ('ad141219-8db3-4544-9513-67715b08f0c2', 'clans', '0138');

INSERT INTO clans (cid, uid)
VALUES ('0138', 'ad141219-8db3-4544-9513-67715b08f0c2');
