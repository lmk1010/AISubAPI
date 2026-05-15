-- Add optional 5m/1h cache creation prices for channel pricing.
-- Existing cache_write_price remains the generic fallback for both TTLs.

ALTER TABLE channel_model_pricing
    ADD COLUMN IF NOT EXISTS cache_write_5m_price NUMERIC(20,12),
    ADD COLUMN IF NOT EXISTS cache_write_1h_price NUMERIC(20,12);

ALTER TABLE channel_pricing_intervals
    ADD COLUMN IF NOT EXISTS cache_write_5m_price NUMERIC(20,12),
    ADD COLUMN IF NOT EXISTS cache_write_1h_price NUMERIC(20,12);

ALTER TABLE channel_account_stats_model_pricing
    ADD COLUMN IF NOT EXISTS cache_write_5m_price NUMERIC(20,12),
    ADD COLUMN IF NOT EXISTS cache_write_1h_price NUMERIC(20,12);

ALTER TABLE channel_account_stats_pricing_intervals
    ADD COLUMN IF NOT EXISTS cache_write_5m_price NUMERIC(20,12),
    ADD COLUMN IF NOT EXISTS cache_write_1h_price NUMERIC(20,12);

COMMENT ON COLUMN channel_model_pricing.cache_write_5m_price IS '5分钟缓存写入每 token 价格，NULL 表示使用 cache_write_price 或默认价';
COMMENT ON COLUMN channel_model_pricing.cache_write_1h_price IS '1小时缓存写入每 token 价格，NULL 表示使用 cache_write_price 或默认价';
COMMENT ON COLUMN channel_pricing_intervals.cache_write_5m_price IS '区间内 5分钟缓存写入每 token 价格，NULL 表示使用 cache_write_price 或默认价';
COMMENT ON COLUMN channel_pricing_intervals.cache_write_1h_price IS '区间内 1小时缓存写入每 token 价格，NULL 表示使用 cache_write_price 或默认价';
COMMENT ON COLUMN channel_account_stats_model_pricing.cache_write_5m_price IS '账号统计定价：5分钟缓存写入每 token 价格';
COMMENT ON COLUMN channel_account_stats_model_pricing.cache_write_1h_price IS '账号统计定价：1小时缓存写入每 token 价格';
COMMENT ON COLUMN channel_account_stats_pricing_intervals.cache_write_5m_price IS '账号统计区间定价：5分钟缓存写入每 token 价格';
COMMENT ON COLUMN channel_account_stats_pricing_intervals.cache_write_1h_price IS '账号统计区间定价：1小时缓存写入每 token 价格';
