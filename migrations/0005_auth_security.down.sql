-- Revert auth hardening + token freshness
ALTER TABLE refresh_tokens
  DROP COLUMN IF EXISTS replaced_by_hash;

ALTER TABLE users
  DROP COLUMN IF EXISTS locked_until,
  DROP COLUMN IF EXISTS failed_login_attempts,
  DROP COLUMN IF EXISTS token_version;
