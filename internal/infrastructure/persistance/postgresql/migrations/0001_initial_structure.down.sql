-- Drop index first
DROP INDEX IF EXISTS idx_session_user_id;

-- Drop child table first (because it references users)
DROP TABLE IF EXISTS users_jwt_sessions;

-- Drop parent table
DROP TABLE IF EXISTS users;
