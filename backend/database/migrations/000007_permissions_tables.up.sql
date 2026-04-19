CREATE TABLE roles (
                       id          SERIAL PRIMARY KEY,
                       name        TEXT UNIQUE NOT NULL,
                       description TEXT,
                       permissions JSONB NOT NULL DEFAULT '{}',
                       updated_at  TIMESTAMP,
                       updated_by  BYTEA REFERENCES users (id) ON DELETE SET NULL
);

INSERT INTO roles (name, description, permissions) VALUES
                                                       ('superuser',   'Full system access',         '{"can_read": true, "can_write": true, "can_delete": true}'),
                                                       ('manager',     'Manages staff and documents', '{"can_read": true, "can_write": true, "can_delete": false}'),
                                                       ('user',        'Standard document access',    '{"can_read": true, "can_write": true, "can_delete": false}'),
                                                       ('contributor', 'Can submit documents',        '{"can_read": true, "can_write": true, "can_delete": false}'),
                                                       ('viewer',      'Read-only access',            '{"can_read": true, "can_write": false, "can_delete": false}');

ALTER TABLE users
    ADD COLUMN role_id           INTEGER REFERENCES roles (id) ON DELETE SET NULL,
    ADD COLUMN role_assigned_at  TIMESTAMP,
    ADD COLUMN role_assigned_by  BYTEA REFERENCES users (id) ON DELETE SET NULL;