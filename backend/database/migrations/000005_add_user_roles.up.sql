ALTER TABLE users ADD COLUMN role text NOT NULL DEFAULT 'user';
ALTER TABLE approved_users RENAME COLUMN permission TO role;