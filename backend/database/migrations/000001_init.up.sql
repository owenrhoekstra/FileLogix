create table users
(
    id    bytea not null
        primary key,
    email text  not null
        unique
);

alter table users
    owner to owenrhoekstra;

create table credentials
(
    id               serial
        primary key,
    user_id          bytea                 not null
        references users
            on delete cascade,
    credential_id    bytea                 not null
        unique,
    public_key       bytea                 not null,
    attestation_type text                  not null,
    transports       text[],
    sign_count       bigint  default 0     not null,
    backup_eligible  boolean default false not null,
    backup_state     boolean default false not null
);

alter table credentials
    owner to owenrhoekstra;

create table webauthn_sessions
(
    id         uuid      not null
        primary key,
    email      text      not null,
    challenge  bytea     not null,
    data       jsonb     not null,
    expires_at timestamp not null
);

alter table webauthn_sessions
    owner to owenrhoekstra;

create table sessions
(
    token      text      not null
        primary key,
    user_id    bytea
        references users
            on delete cascade,
    expires_at timestamp not null
);

alter table sessions
    owner to owenrhoekstra;

create index idx_sessions_user_id
    on sessions (user_id);

