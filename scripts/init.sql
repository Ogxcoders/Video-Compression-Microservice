CREATE TABLE IF NOT EXISTS jobs (
    id SERIAL PRIMARY KEY,
    job_id VARCHAR(255) UNIQUE NOT NULL,
    post_id INTEGER NOT NULL,
    user_id INTEGER,
    compression_type VARCHAR(50) NOT NULL,
    
    video_file_url TEXT,
    video_quality VARCHAR(50),
    video_hls_enabled BOOLEAN DEFAULT FALSE,
    video_hls_variants TEXT[],
    
    image_file_url TEXT,
    image_quality VARCHAR(50),
    image_variants TEXT[],
    
    priority INTEGER DEFAULT 5,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    video_status VARCHAR(50),
    image_status VARCHAR(50),
    
    video_result JSONB,
    image_result JSONB,
    error_message TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    scheduled_time TIMESTAMP,
    
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    
    processing_time INTEGER,
    video_processing_time INTEGER,
    image_processing_time INTEGER
);

CREATE INDEX idx_jobs_job_id ON jobs(job_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_created_at ON jobs(created_at);
CREATE INDEX idx_jobs_post_id ON jobs(post_id);
CREATE INDEX idx_jobs_scheduled_time ON jobs(scheduled_time);

CREATE TABLE IF NOT EXISTS queue_stats (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    total_jobs INTEGER DEFAULT 0,
    completed_jobs INTEGER DEFAULT 0,
    failed_jobs INTEGER DEFAULT 0,
    avg_processing_time INTEGER DEFAULT 0,
    video_jobs INTEGER DEFAULT 0,
    image_jobs INTEGER DEFAULT 0,
    combined_jobs INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_queue_stats_date ON queue_stats(date);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_jobs_updated_at BEFORE UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_queue_stats_updated_at BEFORE UPDATE ON queue_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
