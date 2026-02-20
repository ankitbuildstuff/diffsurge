-- Organizations
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Projects
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (organization_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_projects_org ON projects(organization_id);

-- Environments
CREATE TABLE IF NOT EXISTS environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    base_url VARCHAR(500) NOT NULL,
    is_source BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (project_id, name)
);

CREATE INDEX IF NOT EXISTS idx_environments_project ON environments(project_id);

-- Traffic Logs (partitioned)
CREATE TABLE IF NOT EXISTS traffic_logs (
    id UUID DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    environment_id UUID NOT NULL,
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    query_params JSONB,
    request_headers JSONB,
    request_body JSONB,
    status_code INTEGER NOT NULL,
    response_headers JSONB,
    response_body JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    latency_ms INTEGER,
    ip_address INET,
    user_agent TEXT,
    pii_redacted BOOLEAN DEFAULT false,
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE INDEX IF NOT EXISTS idx_traffic_logs_project ON traffic_logs(project_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_traffic_logs_path ON traffic_logs(project_id, path, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_traffic_logs_status ON traffic_logs(project_id, status_code, timestamp DESC);

-- Replay Sessions
CREATE TABLE IF NOT EXISTS replay_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source_environment_id UUID NOT NULL REFERENCES environments(id),
    target_environment_id UUID NOT NULL REFERENCES environments(id),
    name VARCHAR(255),
    description TEXT,
    traffic_filter JSONB,
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    sample_size INTEGER,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_requests INTEGER DEFAULT 0,
    successful_requests INTEGER DEFAULT 0,
    failed_requests INTEGER DEFAULT 0,
    mismatched_responses INTEGER DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_replay_sessions_project ON replay_sessions(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_replay_sessions_status ON replay_sessions(status);

-- Replay Results
CREATE TABLE IF NOT EXISTS replay_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    replay_session_id UUID NOT NULL REFERENCES replay_sessions(id) ON DELETE CASCADE,
    original_traffic_log_id UUID NOT NULL,
    target_status_code INTEGER,
    target_response_body JSONB,
    target_latency_ms INTEGER,
    status_match BOOLEAN,
    body_match BOOLEAN,
    diff_report JSONB,
    severity VARCHAR(50),
    error_message TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_replay_results_session ON replay_results(replay_session_id);
CREATE INDEX IF NOT EXISTS idx_replay_results_severity ON replay_results(replay_session_id, severity);

-- Schema Versions
CREATE TABLE IF NOT EXISTS schema_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    version VARCHAR(100) NOT NULL,
    schema_type VARCHAR(50) NOT NULL,
    schema_content JSONB NOT NULL,
    git_commit VARCHAR(100),
    git_branch VARCHAR(100),
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (project_id, version)
);

CREATE INDEX IF NOT EXISTS idx_schema_versions_project ON schema_versions(project_id, created_at DESC);

-- Schema Diffs
CREATE TABLE IF NOT EXISTS schema_diffs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    from_version_id UUID REFERENCES schema_versions(id),
    to_version_id UUID REFERENCES schema_versions(id),
    diff_report JSONB NOT NULL,
    has_breaking_changes BOOLEAN DEFAULT false,
    breaking_changes JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_schema_diffs_project ON schema_diffs(project_id, created_at DESC);
