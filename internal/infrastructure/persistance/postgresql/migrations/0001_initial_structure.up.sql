CREATE TABLE IF NOT EXISTS users (
    id           char(26) PRIMARY KEY,
    email        varchar(256) NOT NULL UNIQUE,
    password     char(256) NOT NULL,
    first_name   varchar(256) NOT NULL,
    last_name    varchar(256) NOT NULL,
    nickname     varchar(256) NOT NULL UNIQUE,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_verification_sessions (
    token_hash   char(64) PRIMARY KEY,
    id           char(26) NOT NULL,
    email        varchar(256) NOT NULL UNIQUE,
    password     char(256) NOT NULL,
    first_name   varchar(256) NOT NULL,
    last_name    varchar(256) NOT NULL,
    nickname     varchar(256) NOT NULL,
);

CREATE TABLE IF NOT EXISTS user_password_change_sessions (
    id  char(26) PRIMARY KEY,
    user_id char(26) NOT NULL,
    expires_at TIMESTAMP NOT NULL,

    constraint fk_password_change_session_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);


CREATE TYPE IF NOT EXISTS deviceType AS ENUM(
    'Mobile',
    'Tablet',
    'Desktop'
);

CREATE TABLE IF NOT EXISTS user_jwt_sessions (
    user_id   char(26) NOT NULL 
    refresh_token_id char(26) PRIMARY KEY NOT NULL 
    device_type 
    device_platform varchar(30)
    is_revoked boolean NOT NULL
    expires_at TIMESTAMP NOT NULL

    constraint fk_jwt_session_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_session_user_id ON user_jwt_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_session_expiry ON user_jwt_sessions(expires_at) WHERE is_revoked = false;

CREATE TABLE IF NOT EXISTS users_outbox (
    id char(26) PRIMARY KEY,
    event varchar(256) NOT NULL,
    payload JSONB NOT NULL,
    retry_count smallint NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
)