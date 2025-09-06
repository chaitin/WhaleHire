-- Add UUID default generators for all tables
-- Users table
ALTER TABLE users
ALTER COLUMN id
SET DEFAULT gen_random_uuid();
-- User identities table
ALTER TABLE user_identities
ALTER COLUMN id
SET DEFAULT gen_random_uuid();
ALTER TABLE user_identities
ALTER COLUMN user_id
SET DEFAULT gen_random_uuid();
-- User login histories table
ALTER TABLE user_login_histories
ALTER COLUMN id
SET DEFAULT gen_random_uuid();
ALTER TABLE user_login_histories
ALTER COLUMN user_id
SET DEFAULT gen_random_uuid();
-- Admin roles table
ALTER TABLE admin_roles
ALTER COLUMN id
SET DEFAULT gen_random_uuid();
ALTER TABLE admin_roles
ALTER COLUMN admin_id
SET DEFAULT gen_random_uuid();
-- Admin login histories table
ALTER TABLE admin_login_histories
ALTER COLUMN id
SET DEFAULT gen_random_uuid();
ALTER TABLE admin_login_histories
ALTER COLUMN admin_id
SET DEFAULT gen_random_uuid();
-- Admins table (handled in previous migration)
-- ALTER TABLE admins ALTER COLUMN id SET DEFAULT gen_random_uuid();