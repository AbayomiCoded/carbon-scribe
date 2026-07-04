package monitoring

import (
	"context"

	"carbon-scribe/project-portal/project-portal-backend/internal/monitoring/ingestion"
)

// Service orchestrates satellite data ingestion and retrieval.
type Service struct {
	pipeline        *ingestion.SatellitePipeline
	webhookPipeline *ingestion.WebhookPipeline
	repo            Repository
}

// NewService constructs a monitoring Service.
func NewService(repo Repository) *Service {
	return &Service{
		pipeline:        ingestion.NewSatellitePipeline(repo),
		webhookPipeline: ingestion.NewWebhookPipeline(repo),
		repo:            repo,
	}
}

// IngestSatellite validates and persists a satellite reading.
func (s *Service) IngestSatellite(ctx context.Context, req IngestSatelliteRequest) (*SatelliteReading, error) {
	ir := ingestion.IngestRequest{
		ProjectID:   req.ProjectID,
		Source:      req.Source,
		DataType:    req.DataType,
		NDVIMean:    req.NDVIMean,
		NDVIMin:     req.NDVIMin,
		NDVIMax:     req.NDVIMax,
		BiomassTons: req.BiomassTons,
		ImageryURL:  req.ImageryURL,
		BoundingBox: req.BoundingBox,
		Metadata:    req.Metadata,
		CapturedAt:  req.CapturedAt,
	}
	return s.pipeline.Ingest(ctx, ir)
}

// IngestWebhook validates and persists a webhook reading.
func (s *Service) IngestWebhook(ctx context.Context, req IngestWebhookRequest) (*WebhookReading, error) {
	wr := ingestion.WebhookRequest{
		ProjectID:   req.ProjectID,
		Source:      req.Source,
		SourceType:  req.SourceType,
		MetricName:  req.MetricName,
		MetricValue: req.MetricValue,
		Unit:        req.Unit,
		Location:    req.Location,
		Metadata:    req.Metadata,
		CapturedAt:  req.CapturedAt,
		WebhookID:   req.WebhookID,
	}
	return s.webhookPipeline.Ingest(ctx, wr)
}

// ListReadings returns the most recent satellite readings for a project.
func (s *Service) ListReadings(ctx context.Context, projectID string, limit int) ([]SatelliteReading, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListByProject(ctx, projectID, limit)
}

// ListWebhookReadings returns the most recent webhook readings for a project.
func (s *Service) ListWebhookReadings(ctx context.Context, projectID string, limit int) ([]WebhookReading, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListWebhookReadings(ctx, projectID, limit)
}

// ListWebhookReadingsByMetric returns readings filtered by metric name.
func (s *Service) ListWebhookReadingsByMetric(ctx context.Context, projectID, metricName string, limit int) ([]WebhookReading, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListWebhookReadingsByMetric(ctx, projectID, metricName, limit)
}

// ListWebhookReadingsBySource returns readings filtered by source.
func (s *Service) ListWebhookReadingsBySource(ctx context.Context, projectID, source string, limit int) ([]WebhookReading, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListWebhookReadingsBySource(ctx, projectID, source, limit)
}
