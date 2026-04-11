ALTER TABLE approved_users DROP CONSTRAINT approved_users_pkey;
ALTER TABLE approved_users DROP COLUMN id;
ALTER TABLE approved_users ADD COLUMN id serial PRIMARY KEY;