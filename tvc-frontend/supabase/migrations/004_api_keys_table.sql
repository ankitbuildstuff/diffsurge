-- ============================================================================
-- TVC — API Keys Table
-- Run this in Supabase SQL Editor or via `supabase db push`
-- ============================================================================

-- ── API Keys Table ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS api_keys (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
  name            VARCHAR(100) NOT NULL,
  key_prefix      VARCHAR(20) NOT NULL,
  key_hash        TEXT NOT NULL,
  last_used_at    TIMESTAMPTZ,
  expires_at      TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by      UUID NOT NULL REFERENCES auth.users(id)
);

-- ── Indexes ────────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_api_keys_organization_id ON api_keys(organization_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_project_id ON api_keys(project_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON api_keys(key_prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_created_by ON api_keys(created_by);

-- ── Comments ───────────────────────────────────────────────────────────────

COMMENT ON TABLE api_keys IS 'API keys for programmatic access to TVC services';
COMMENT ON COLUMN api_keys.key_prefix IS 'First ~10 chars of key for display (e.g., tvc_live_abc...)';
COMMENT ON COLUMN api_keys.key_hash IS 'bcrypt hash of the full key';
COMMENT ON COLUMN api_keys.last_used_at IS 'Timestamp of last successful authentication';
COMMENT ON COLUMN api_keys.expires_at IS 'Optional expiration date for the key';
