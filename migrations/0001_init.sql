-- Desired state: which models exist and how many replicas each should have.
CREATE TABLE models (
  id               BIGSERIAL PRIMARY KEY,
  name             TEXT UNIQUE NOT NULL,
  image            TEXT NOT NULL,
  hf_repo          TEXT NOT NULL,
  desired_replicas INT  NOT NULL DEFAULT 0 CHECK (desired_replicas >= 0),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Durable queue. Used from Tuesday.
CREATE TABLE jobs (
  id         BIGSERIAL PRIMARY KEY,
  model_id   BIGINT NOT NULL REFERENCES models(id),
  status     TEXT NOT NULL DEFAULT 'queued'
             CHECK (status IN ('queued', 'running', 'done', 'failed')),
  payload    JSONB NOT NULL,
  result     JSONB,
  attempts   INT NOT NULL DEFAULT 0,
  locked_by  TEXT,
  locked_at  TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Keeps the "claim the oldest queued job" query fast as done jobs pile up.
CREATE INDEX jobs_queued_idx ON jobs (created_at) WHERE status = 'queued';
