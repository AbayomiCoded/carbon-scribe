package monitoring

import (
	"context"
	"time"

	"carbon-scribe/project-portal/project-portal-backend/internal/monitoring/ingestion"
)

// Repository defines the persistence contract for monitoring data.
type Repository interface {
	// Satellite methods
	Save(ctx context.Context, reading *ingestion.SatelliteReading) error
	ListByProject(ctx context.Context, projectID string, limit int) ([]ingestion.SatelliteReading, error)

	// Webhook methods
	SaveWebhookReading(ctx context.Context, reading *ingestion.WebhookReading) error
	GetWebhookReadingByID(ctx context.Context, webhookID string) (*ingestion.WebhookReading, error)
	ListWebhookReadings(ctx context.Context, projectID string, limit int) ([]ingestion.WebhookReading, error)
	ListWebhookReadingsByMetric(ctx context.Context, projectID, metricName string, limit int) ([]ingestion.WebhookReading, error)
	ListWebhookReadingsBySource(ctx context.Context, projectID, source string, limit int) ([]ingestion.WebhookReading, error)
	GetWebhookReadingsByTimeRange(ctx context.Context, projectID string, start, end time.Time) ([]ingestion.WebhookReading, error)
}
