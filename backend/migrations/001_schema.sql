CREATE TABLE IF NOT EXISTS features (
    entity_id BIGINT PRIMARY KEY,
    click_rate DOUBLE PRECISION NOT NULL,
    popularity_score DOUBLE PRECISION NOT NULL,
    bot_score DOUBLE PRECISION NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO features (entity_id, click_rate, popularity_score, bot_score)
SELECT id,
       round((random() * 0.25)::numeric, 6),
       round((random() * 100)::numeric, 6),
       round((random())::numeric, 6)
FROM generate_series(1, 1000000) AS ids(id)
ON CONFLICT (entity_id) DO NOTHING;
