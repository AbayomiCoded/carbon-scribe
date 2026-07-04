package monitoring

import (
	"context"

	"carbon-scribe/project-portal/project-portal-backend/internal/monitoring/ingestion"
)

// Service orchestrates satellite data ingestion and retrieval.
type Service struct {
	pipeline        *ingestion.SatellitePipeline
	webhookPipeline *ingestion.WebhookPipeline
	iotPipeline     *ingestion.IoTPipeline
	repo            Repository
}

// NewService constructs a monitoring Service.
func NewService(repo Repository) *Service {
	return &Service{
		pipeline:        ingestion.NewSatellitePipeline(repo),
		webhookPipeline: ingestion.NewWebhookPipeline(repo),
		iotPipeline:     ingestion.NewIoTPipeline(repo),
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

// IngestIoT validates and persists an IoT reading.
func (s *Service) IngestIoT(ctx context.Context, req IngestIoTRequest) (*IoTReading, error) {
	ir := ingestion.IoTRequest{
		ProjectID:      req.ProjectID,
		SensorID:       req.SensorID,
		SensorType:     req.SensorType,
		Value:          req.Value,
		Unit:           req.Unit,
		Location:       req.Location,
		Metadata:       req.Metadata,
		CapturedAt:     req.CapturedAt,
		DeviceID:       req.DeviceID,
		BatteryLevel:   req.BatteryLevel,
		SignalStrength: req.SignalStrength,
	}
	return s.iotPipeline.Ingest(ctx, ir)
}

// ListIoTReadings returns the most recent IoT readings for a project.
func (s *Service) ListIoTReadings(ctx context.Context, projectID string, limit int) ([]IoTReading, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetIoTReadingsByProject(ctx, projectID, limit)
}

// ListIoTReadingsBySensor returns readings filtered by sensor ID.
func (s *Service) ListIoTReadingsBySensor(ctx context.Context, projectID, sensorID string, limit int) ([]IoTReading, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetIoTReadingsBySensor(ctx, projectID, sensorID, limit)
}

// ListIoTReadingsByType returns readings filtered by sensor type.
func (s *Service) ListIoTReadingsByType(ctx context.Context, projectID, sensorType string, limit int) ([]IoTReading, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetIoTReadingsByType(ctx, projectID, sensorType, limit)
}

// MetricService methods
func (s *Service) SaveMetric(ctx context.Context, metric *SystemMetric) error {
	return s.repo.SaveMetric(ctx, metric)
}

func (s *Service) SaveMetricsBatch(ctx context.Context, metrics []SystemMetric) error {
	return s.repo.SaveMetricsBatch(ctx, metrics)
}

func (s *Service) QueryMetrics(ctx context.Context, req MetricQueryRequest) ([]SystemMetric, error) {
	return s.repo.QueryMetrics(ctx, req)
}

func (s *Service) GetMetricAggregation(ctx context.Context, req AggregationRequest) (*MetricAggregationResult, error) {
	return s.repo.GetMetricAggregation(ctx, req)
}

func (s *Service) GetMetricRate(ctx context.Context, req RateRequest) (*RateResult, error) {
	return s.repo.GetMetricRate(ctx, req)
}

func (s *Service) GetLatestMetric(ctx context.Context, metricName, service string) (*SystemMetric, error) {
	return s.repo.GetLatestMetric(ctx, metricName, service)
}

func (s *Service) GetMetricLabels(ctx context.Context, metricName string) ([]map[string]string, error) {
	return s.repo.GetMetricLabels(ctx, metricName)
}

func (s *Service) CleanupOldMetrics(ctx context.Context, retentionDays int) (int64, error) {
	return s.repo.CleanupOldMetrics(ctx, retentionDays)
}
