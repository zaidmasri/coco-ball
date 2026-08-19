-- +migrate Up

-- SQLite can't ALTER a column-level UNIQUE constraint away, so the table is
-- rebuilt without it. The blanket UNIQUE on email must go: once deleted_at
-- exists, a soft-deleted row keeps occupying its email forever, and a
-- global UNIQUE would permanently block that email from being reused by a
-- new signup. A partial unique index scoped to active rows replaces it,
-- preserving uniqueness among live accounts while freeing deleted ones up.
CREATE TABLE users_new (
	id TEXT PRIMARY KEY,
	email TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	first_name TEXT NOT NULL DEFAULT '',
	last_name TEXT NOT NULL DEFAULT '',
	deleted_at INTEGER
);

INSERT INTO users_new (id, email, created_at, first_name, last_name)
SELECT id, email, created_at, first_name, last_name FROM users;

DROP TABLE users;
ALTER TABLE users_new RENAME TO users;

CREATE UNIQUE INDEX idx_users_email_active ON users(email) WHERE deleted_at IS NULL;

-- +migrate Down

DROP INDEX idx_users_email_active;

CREATE TABLE users_old (
	id TEXT PRIMARY KEY,
	email TEXT UNIQUE NOT NULL,
	created_at INTEGER NOT NULL,
	first_name TEXT NOT NULL DEFAULT '',
	last_name TEXT NOT NULL DEFAULT ''
);

INSERT INTO users_old (id, email, created_at, first_name, last_name)
SELECT id, email, created_at, first_name, last_name FROM users WHERE deleted_at IS NULL;

DROP TABLE users;
ALTER TABLE users_old RENAME TO users;
