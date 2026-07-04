package monitoring

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"carbon-scribe/project-portal/project-portal-backend/internal/monitoring/ingestion"
)

// PostgresRepository implements Repository using a *sql.DB connection.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository constructs a PostgresRepository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Save inserts a SatelliteReading into the satellite_readings table.
func (r *PostgresRepository) Save(ctx context.Context, reading *ingestion.SatelliteReading) error {
	metaJSON, err := json.Marshal(reading.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	var bboxJSON []byte
	if reading.BoundingBox != nil {
		bboxJSON, err = json.Marshal(reading.BoundingBox)
		if err != nil {
			return fmt.Errorf("marshal bounding_box: %w", err)
		}
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO satellite_readings
			(id, project_id, source, data_type, ndvi_mean, ndvi_min, ndvi_max,
			 biomass_tons, imagery_url, bounding_box, metadata, captured_at, ingested_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		reading.ID, reading.ProjectID, reading.Source, reading.DataType,
		reading.NDVIMean, reading.NDVIMin, reading.NDVIMax,
		reading.BiomassTons, reading.ImageryURL, bboxJSON, metaJSON,
		reading.CapturedAt, reading.IngestedAt,
	)
	return err
}

// ListByProject returns the most recent satellite readings for a project.
func (r *PostgresRepository) ListByProject(ctx context.Context, projectID string, limit int) ([]ingestion.SatelliteReading, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, source, data_type, ndvi_mean, ndvi_min, ndvi_max,
		       biomass_tons, imagery_url, bounding_box, metadata, captured_at, ingested_at
		FROM satellite_readings
		WHERE project_id = $1
		ORDER BY captured_at DESC
		LIMIT $2`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ingestion.SatelliteReading
	for rows.Next() {
		var sr ingestion.SatelliteReading
		var bboxJSON, metaJSON []byte
		if err := rows.Scan(
			&sr.ID, &sr.ProjectID, &sr.Source, &sr.DataType,
			&sr.NDVIMean, &sr.NDVIMin, &sr.NDVIMax,
			&sr.BiomassTons, &sr.ImageryURL, &bboxJSON, &metaJSON,
			&sr.CapturedAt, &sr.IngestedAt,
		); err != nil {
			return nil, err
		}
		if len(bboxJSON) > 0 {
			sr.BoundingBox = &ingestion.BoundingBox{}
			_ = json.Unmarshal(bboxJSON, sr.BoundingBox)
		}
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &sr.Metadata)
		}
		results = append(results, sr)
	}
	return results, rows.Err()
}

// SaveWebhookReading inserts a WebhookReading into the webhook_readings table.
func (r *PostgresRepository) SaveWebhookReading(ctx context.Context, reading *ingestion.WebhookReading) error {
	metaJSON, err := json.Marshal(reading.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	var locationJSON []byte
	if reading.Location != nil {
		locationJSON, err = json.Marshal(reading.Location)
		if err != nil {
			return fmt.Errorf("marshal location: %w", err)
		}
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO webhook_readings
			(id, project_id, source, source_type, metric_name, metric_value, unit,
			 location, metadata, webhook_id, captured_at, ingested_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		reading.ID, reading.ProjectID, reading.Source, reading.SourceType,
		reading.MetricName, reading.MetricValue, reading.Unit,
		locationJSON, metaJSON, reading.WebhookID,
		reading.CapturedAt, reading.IngestedAt,
	)
	return err
}

// GetWebhookReadingByID retrieves a webhook reading by its webhook ID (for deduplication).
func (r *PostgresRepository) GetWebhookReadingByID(ctx context.Context, webhookID string) (*ingestion.WebhookReading, error) {
	var reading ingestion.WebhookReading
	var locationJSON, metaJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, source, source_type, metric_name, metric_value, unit,
		       location, metadata, webhook_id, captured_at, ingested_at
		FROM webhook_readings
		WHERE webhook_id = $1`, webhookID).Scan(
		&reading.ID, &reading.ProjectID, &reading.Source, &reading.SourceType,
		&reading.MetricName, &reading.MetricValue, &reading.Unit,
		&locationJSON, &metaJSON, &reading.WebhookID,
		&reading.CapturedAt, &reading.IngestedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if len(locationJSON) > 0 {
		reading.Location = &ingestion.Location{}
		_ = json.Unmarshal(locationJSON, reading.Location)
	}
	if len(metaJSON) > 0 {
		_ = json.Unmarshal(metaJSON, &reading.Metadata)
	}
	return &reading, nil
}

// ListWebhookReadings returns the most recent webhook readings for a project.
func (r *PostgresRepository) ListWebhookReadings(ctx context.Context, projectID string, limit int) ([]ingestion.WebhookReading, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, source, source_type, metric_name, metric_value, unit,
		       location, metadata, webhook_id, captured_at, ingested_at
		FROM webhook_readings
		WHERE project_id = $1
		ORDER BY captured_at DESC
		LIMIT $2`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanWebhookReadings(rows)
}

// ListWebhookReadingsByMetric returns readings filtered by metric name.
func (r *PostgresRepository) ListWebhookReadingsByMetric(ctx context.Context, projectID, metricName string, limit int) ([]ingestion.WebhookReading, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, source, source_type, metric_name, metric_value, unit,
		       location, metadata, webhook_id, captured_at, ingested_at
		FROM webhook_readings
		WHERE project_id = $1 AND metric_name = $2
		ORDER BY captured_at DESC
		LIMIT $3`, projectID, metricName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanWebhookReadings(rows)
}

// ListWebhookReadingsBySource returns readings filtered by source.
func (r *PostgresRepository) ListWebhookReadingsBySource(ctx context.Context, projectID, source string, limit int) ([]ingestion.WebhookReading, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, source, source_type, metric_name, metric_value, unit,
		       location, metadata, webhook_id, captured_at, ingested_at
		FROM webhook_readings
		WHERE project_id = $1 AND source = $2
		ORDER BY captured_at DESC
		LIMIT $3`, projectID, source, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanWebhookReadings(rows)
}

// GetWebhookReadingsByTimeRange returns readings within a time range.
func (r *PostgresRepository) GetWebhookReadingsByTimeRange(ctx context.Context, projectID string, start, end time.Time) ([]ingestion.WebhookReading, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, source, source_type, metric_name, metric_value, unit,
		       location, metadata, webhook_id, captured_at, ingested_at
		FROM webhook_readings
		WHERE project_id = $1 AND captured_at BETWEEN $2 AND $3
		ORDER BY captured_at ASC`, projectID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanWebhookReadings(rows)
}

// scanWebhookReadings is a helper to scan webhook readings from rows.
func scanWebhookReadings(rows *sql.Rows) ([]ingestion.WebhookReading, error) {
	var results []ingestion.WebhookReading
	for rows.Next() {
		var wr ingestion.WebhookReading
		var locationJSON, metaJSON []byte
		if err := rows.Scan(
			&wr.ID, &wr.ProjectID, &wr.Source, &wr.SourceType,
			&wr.MetricName, &wr.MetricValue, &wr.Unit,
			&locationJSON, &metaJSON, &wr.WebhookID,
			&wr.CapturedAt, &wr.IngestedAt,
		); err != nil {
			return nil, err
		}
		if len(locationJSON) > 0 {
			wr.Location = &ingestion.Location{}
			_ = json.Unmarshal(locationJSON, wr.Location)
		}
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &wr.Metadata)
		}
		results = append(results, wr)
	}
	return results, rows.Err()
}
