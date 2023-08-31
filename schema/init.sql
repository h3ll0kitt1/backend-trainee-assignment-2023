CREATE TABLE IF NOT EXISTS users (
    id integer PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS segments (
    slug varchar(255) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users_segments (
    user_id         integer references users (id) on delete cascade      not null,
    segment_slug    varchar(255) references segments (slug) on delete cascade   not null,
    expires_at      timestamp,
    UNIQUE (user_id, segment_slug)
);

CREATE TABLE IF NOT EXISTS segments_history (
    user_id       int              not null,
    segment_slug  varchar(255)     not null,
    action        boolean          not null,
    action_time   timestamp        not null
);