package features

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Get(ctx context.Context, entityID int64) (Feature, error) {
	var feature Feature
	err := s.pool.QueryRow(ctx, `
		SELECT entity_id, click_rate, popularity_score, bot_score, updated_at
		FROM features WHERE entity_id = $1`, entityID).
		Scan(&feature.EntityID, &feature.ClickRate, &feature.PopularityScore, &feature.BotScore, &feature.UpdatedAt)
	return feature, err
}
