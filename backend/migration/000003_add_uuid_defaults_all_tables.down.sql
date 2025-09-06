-- Remove UUID default generators for all tables

-- Users table
ALTER TABLE users ALTER COLUMN id DROP DEFAULT;

-- User identities table
ALTER TABLE user_identities ALTER COLUMN id DROP DEFAULT;
ALTER TABLE user_identities ALTER COLUMN user_id DROP DEFAULT;

-- User login histories table
ALTER TABLE user_login_histories ALTER COLUMN id DROP DEFAULT;
ALTER TABLE user_login_histories ALTER COLUMN user_id DROP DEFAULT;

-- Admin roles table
ALTER TABLE admin_roles ALTER COLUMN id DROP DEFAULT;
ALTER TABLE admin_roles ALTER COLUMN admin_id DROP DEFAULT;

-- Admin login histories table
ALTER TABLE admin_login_histories ALTER COLUMN id DROP DEFAULT;
ALTER TABLE admin_login_histories ALTER COLUMN admin_id DROP DEFAULT;

-- Admins table (handled in previous migration)
-- ALTER TABLE admins ALTER COLUMN id DROP DEFAULT;