ALTER TABLE tasks ADD COLUMN IF NOT EXISTS interval_type TEXT NOT NULL DEFAULT 'none';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS interval_days INT;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS day_of_month INT;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS is_even BOOLEAN;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS specific_dates JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS next_run_at TIMESTAMPTZ;

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_interval_type') THEN
		ALTER TABLE tasks
		ADD CONSTRAINT chk_interval_type
		CHECK (interval_type IN ('none', 'daily', 'monthly', 'parity', 'dates'));
	END IF;
END
$$;

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_interval_days') THEN
		ALTER TABLE tasks
		ADD CONSTRAINT chk_interval_days
		CHECK (interval_days IS NULL OR interval_days >= 1);
	END IF;
END
$$;

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_day_of_month') THEN
		ALTER TABLE tasks
		ADD CONSTRAINT chk_day_of_month
		CHECK (day_of_month IS NULL OR (day_of_month >= 1 AND day_of_month <= 30));
	END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_tasks_next_run_at ON tasks (next_run_at);
