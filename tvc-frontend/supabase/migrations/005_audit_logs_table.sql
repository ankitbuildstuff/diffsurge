-- ============================================================================
-- TVC — Audit Logs Table
-- Run this in Supabase SQL Editor or via `supabase db push`
-- ============================================================================

-- ── Audit Action Enum ──────────────────────────────────────────────────────

DO $$ BEGIN
  CREATE TYPE audit_action AS ENUM (
    'create',
    'update',
    'delete',
    'invite',
    'remove',
    'login',
    'logout',
    'access'
  );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ── Audit Logs Table ───────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS audit_logs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id         UUID REFERENCES auth.users(id) ON DELETE SET NULL,
  action          audit_action NOT NULL,
  resource_type   VARCHAR(50) NOT NULL,
  resource_id     UUID,
  details         JSONB,
  ip_address      INET,
  user_agent      TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── Indexes ────────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_audit_logs_organization_id ON audit_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Composite index for common query pattern (org + time)
CREATE INDEX IF NOT EXISTS idx_audit_logs_org_created ON audit_logs(organization_id, created_at DESC);

-- ── Comments ───────────────────────────────────────────────────────────────

COMMENT ON TABLE audit_logs IS 'Audit trail for all critical actions in the system';
COMMENT ON COLUMN audit_logs.action IS 'Type of action performed';
COMMENT ON COLUMN audit_logs.resource_type IS 'Type of resource affected (e.g., project, environment, member)';
COMMENT ON COLUMN audit_logs.resource_id IS 'UUID of the affected resource';
COMMENT ON COLUMN audit_logs.details IS 'Additional context about the action (changes, metadata, etc.)';
COMMENT ON COLUMN audit_logs.ip_address IS 'IP address of the user who performed the action';
COMMENT ON COLUMN audit_logs.user_agent IS 'Browser/client user agent string';
