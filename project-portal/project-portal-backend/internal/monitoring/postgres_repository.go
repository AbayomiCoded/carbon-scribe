package monitoring

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// PostgresRepository implements Repository using a *sql.DB connection.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository constructs a PostgresRepository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ... existing satellite, webhook, IoT, and metric methods ...

// ============================================================================
// Health Check Methods
// ============================================================================

// SaveHealthCheck inserts a ServiceHealthCheck into the service_health_checks table.
func (r *PostgresRepository) SaveHealthCheck(ctx context.Context, check *ServiceHealthCheck) error {
	detailsJSON, err := json.Marshal(check.Details)
	if err != nil {
		return fmt.Errorf("marshal details: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO service_health_checks
			(id, service_name, check_type, status, latency_ms, error, details, checked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		check.ID, check.ServiceName, check.CheckType, check.Status,
		check.LatencyMs, check.Error, detailsJSON, check.CheckedAt, check.CreatedAt,
	)
	return err
}

// SaveHealthCheckResult inserts a HealthCheckResult into the health_check_results table.
func (r *PostgresRepository) SaveHealthCheckResult(ctx context.Context, result *HealthCheckResult) error {
	detailsJSON, err := json.Marshal(result.Details)
	if err != nil {
		return fmt.Errorf("marshal details: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO health_check_results
			(id, check_id, service_name, status, latency_ms, error, details, checked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		result.ID, result.CheckID, result.ServiceName, result.Status,
		result.LatencyMs, result.Error, detailsJSON, result.CheckedAt, result.CreatedAt,
	)
	return err
}

// GetHealthCheck retrieves a health check by ID.
func (r *PostgresRepository) GetHealthCheck(ctx context.Context, checkID string) (*ServiceHealthCheck, error) {
	var check ServiceHealthCheck
	var detailsJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, service_name, check_type, status, latency_ms, error, details, checked_at, created_at
		FROM service_health_checks
		WHERE id = $1`, checkID).Scan(
		&check.ID, &check.ServiceName, &check.CheckType, &check.Status,
		&check.LatencyMs, &check.Error, &detailsJSON, &check.CheckedAt, &check.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get health check: %w", err)
	}

	if len(detailsJSON) > 0 {
		_ = json.Unmarshal(detailsJSON, &check.Details)
	}
	return &check, nil
}

// GetHealthCheckResults retrieves health check results for a check ID.
func (r *PostgresRepository) GetHealthCheckResults(ctx context.Context, checkID string, limit int) ([]HealthCheckResult, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, check_id, service_name, status, latency_ms, error, details, checked_at, created_at
		FROM health_check_results
		WHERE check_id = $1
		ORDER BY checked_at DESC
		LIMIT $2`, checkID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHealthCheckResults(rows)
}

// GetHealthChecksByService retrieves all health checks for a service.
func (r *PostgresRepository) GetHealthChecksByService(ctx context.Context, serviceName string) ([]ServiceHealthCheck, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, service_name, check_type, status, latency_ms, error, details, checked_at, created_at
		FROM service_health_checks
		WHERE service_name = $1`, serviceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []ServiceHealthCheck
	for rows.Next() {
		var check ServiceHealthCheck
		var detailsJSON []byte
		if err := rows.Scan(&check.ID, &check.ServiceName, &check.CheckType, &check.Status,
			&check.LatencyMs, &check.Error, &detailsJSON, &check.CheckedAt, &check.CreatedAt); err != nil {
			return nil, err
		}
		if len(detailsJSON) > 0 {
			_ = json.Unmarshal(detailsJSON, &check.Details)
		}
		checks = append(checks, check)
	}
	return checks, rows.Err()
}

// GetHealthCheckResultsByTimeRange retrieves health check results within a time range.
func (r *PostgresRepository) GetHealthCheckResultsByTimeRange(ctx context.Context, serviceName string, start, end time.Time) ([]HealthCheckResult, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, check_id, service_name, status, latency_ms, error, details, checked_at, created_at
		FROM health_check_results
		WHERE service_name = $1 AND checked_at BETWEEN $2 AND $3
		ORDER BY checked_at ASC`, serviceName, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHealthCheckResults(rows)
}

// GetLatestHealthCheckResult retrieves the latest health check result for a service.
func (r *PostgresRepository) GetLatestHealthCheckResult(ctx context.Context, serviceName string) (*HealthCheckResult, error) {
	var result HealthCheckResult
	var detailsJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, check_id, service_name, status, latency_ms, error, details, checked_at, created_at
		FROM health_check_results
		WHERE service_name = $1
		ORDER BY checked_at DESC
		LIMIT 1`, serviceName).Scan(
		&result.ID, &result.CheckID, &result.ServiceName, &result.Status,
		&result.LatencyMs, &result.Error, &detailsJSON, &result.CheckedAt, &result.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest health check: %w", err)
	}

	if len(detailsJSON) > 0 {
		_ = json.Unmarshal(detailsJSON, &result.Details)
	}
	return &result, nil
}

// scanHealthCheckResults is a helper to scan health check results from rows.
func scanHealthCheckResults(rows *sql.Rows) ([]HealthCheckResult, error) {
	var results []HealthCheckResult
	for rows.Next() {
		var r HealthCheckResult
		var detailsJSON []byte
		if err := rows.Scan(&r.ID, &r.CheckID, &r.ServiceName, &r.Status,
			&r.LatencyMs, &r.Error, &detailsJSON, &r.CheckedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		if len(detailsJSON) > 0 {
			_ = json.Unmarshal(detailsJSON, &r.Details)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// ============================================================================
// Alert Methods
// ============================================================================

// SaveAlert inserts a SystemAlert into the system_alerts table.
func (r *PostgresRepository) SaveAlert(ctx context.Context, alert *SystemAlert) error {
	detailsJSON, err := json.Marshal(alert.Details)
	if err != nil {
		return fmt.Errorf("marshal details: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO system_alerts
			(id, service_name, severity, status, title, message, details,
			 triggered_by, resolved_by, acknowledged_by, acknowledged_at, resolved_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		alert.ID, alert.ServiceName, alert.Severity, alert.Status,
		alert.Title, alert.Message, detailsJSON,
		alert.TriggeredBy, alert.ResolvedBy, alert.AcknowledgedBy,
		alert.AcknowledgedAt, alert.ResolvedAt, alert.CreatedAt, alert.UpdatedAt,
	)
	return err
}

// GetAlert retrieves an alert by ID.
func (r *PostgresRepository) GetAlert(ctx context.Context, alertID string) (*SystemAlert, error) {
	var alert SystemAlert
	var detailsJSON []byte
	var acknowledgedAt, resolvedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, service_name, severity, status, title, message, details,
		       triggered_by, resolved_by, acknowledged_by, acknowledged_at, resolved_at, created_at, updated_at
		FROM system_alerts
		WHERE id = $1`, alertID).Scan(
		&alert.ID, &alert.ServiceName, &alert.Severity, &alert.Status,
		&alert.Title, &alert.Message, &detailsJSON,
		&alert.TriggeredBy, &alert.ResolvedBy, &alert.AcknowledgedBy,
		&acknowledgedAt, &resolvedAt, &alert.CreatedAt, &alert.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get alert: %w", err)
	}

	if acknowledgedAt.Valid {
		alert.AcknowledgedAt = &acknowledgedAt.Time
	}
	if resolvedAt.Valid {
		alert.ResolvedAt = &resolvedAt.Time
	}
	if len(detailsJSON) > 0 {
		_ = json.Unmarshal(detailsJSON, &alert.Details)
	}
	return &alert, nil
}

// ListAlerts retrieves alerts with optional filters.
func (r *PostgresRepository) ListAlerts(ctx context.Context, status AlertStatus, severity AlertSeverity, limit int) ([]SystemAlert, error) {
	query := `SELECT id, service_name, severity, status, title, message, details,
	          triggered_by, resolved_by, acknowledged_by, acknowledged_at, resolved_at, created_at, updated_at
	          FROM system_alerts WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	if severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, severity)
		argIndex++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// ListAlertsByService retrieves alerts for a specific service.
func (r *PostgresRepository) ListAlertsByService(ctx context.Context, serviceName string, limit int) ([]SystemAlert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, service_name, severity, status, title, message, details,
		       triggered_by, resolved_by, acknowledged_by, acknowledged_at, resolved_at, created_at, updated_at
		FROM system_alerts
		WHERE service_name = $1
		ORDER BY created_at DESC
		LIMIT $2`, serviceName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// UpdateAlert updates an alert.
func (r *PostgresRepository) UpdateAlert(ctx context.Context, alertID string, req UpdateAlertRequest) (*SystemAlert, error) {
	// First, get the current alert
	alert, err := r.GetAlert(ctx, alertID)
	if err != nil || alert == nil {
		return nil, err
	}

	// Update fields if provided
	if req.Status != "" {
		alert.Status = req.Status
	}
	if req.ResolvedBy != "" {
		alert.ResolvedBy = req.ResolvedBy
	}
	if req.Details != nil {
		alert.Details = req.Details
	}
	alert.UpdatedAt = time.Now()

	detailsJSON, err := json.Marshal(alert.Details)
	if err != nil {
		return nil, fmt.Errorf("marshal details: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE system_alerts
		SET status = $1, resolved_by = $2, details = $3, updated_at = $4,
		    resolved_at = CASE WHEN $1 = 'resolved' THEN NOW() ELSE resolved_at END
		WHERE id = $5`,
		alert.Status, alert.ResolvedBy, detailsJSON, alert.UpdatedAt, alertID)
	if err != nil {
		return nil, fmt.Errorf("update alert: %w", err)
	}

	return alert, nil
}

// AcknowledgeAlert acknowledges an alert.
func (r *PostgresRepository) AcknowledgeAlert(ctx context.Context, alertID, acknowledgedBy string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE system_alerts
		SET status = $1, acknowledged_by = $2, acknowledged_at = $3, updated_at = $3
		WHERE id = $4`,
		AlertStatusAcknowledged, acknowledgedBy, now, alertID)
	return err
}

// ResolveAlert resolves an alert.
func (r *PostgresRepository) ResolveAlert(ctx context.Context, alertID, resolvedBy string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE system_alerts
		SET status = $1, resolved_by = $2, resolved_at = $3, updated_at = $3
		WHERE id = $4`,
		AlertStatusResolved, resolvedBy, now, alertID)
	return err
}

// GetActiveAlerts retrieves all active alerts.
func (r *PostgresRepository) GetActiveAlerts(ctx context.Context) ([]SystemAlert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, service_name, severity, status, title, message, details,
		       triggered_by, resolved_by, acknowledged_by, acknowledged_at, resolved_at, created_at, updated_at
		FROM system_alerts
		WHERE status IN ('active', 'acknowledged')
		ORDER BY severity DESC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// GetAlertCountsBySeverity returns alert counts grouped by severity.
func (r *PostgresRepository) GetAlertCountsBySeverity(ctx context.Context) (map[AlertSeverity]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT severity, COUNT(*) as count
		FROM system_alerts
		WHERE status IN ('active', 'acknowledged')
		GROUP BY severity`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[AlertSeverity]int)
	for rows.Next() {
		var severity AlertSeverity
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			return nil, err
		}
		counts[severity] = count
	}
	return counts, rows.Err()
}

// scanAlerts is a helper to scan alerts from rows.
func scanAlerts(rows *sql.Rows) ([]SystemAlert, error) {
	var alerts []SystemAlert
	for rows.Next() {
		var a SystemAlert
		var detailsJSON []byte
		var acknowledgedAt, resolvedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.ServiceName, &a.Severity, &a.Status,
			&a.Title, &a.Message, &detailsJSON,
			&a.TriggeredBy, &a.ResolvedBy, &a.AcknowledgedBy,
			&acknowledgedAt, &resolvedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		if acknowledgedAt.Valid {
			a.AcknowledgedAt = &acknowledgedAt.Time
		}
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		if len(detailsJSON) > 0 {
			_ = json.Unmarshal(detailsJSON, &a.Details)
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// ============================================================================
// Service Dependency Methods
// ============================================================================

// SaveServiceDependency inserts a ServiceDependency.
func (r *PostgresRepository) SaveServiceDependency(ctx context.Context, dependency *ServiceDependency) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO service_dependencies
			(id, service_name, depends_on, dependency_type, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		dependency.ID, dependency.ServiceName, dependency.DependsOn,
		dependency.DependencyType, dependency.Description, dependency.IsActive,
		dependency.CreatedAt, dependency.UpdatedAt)
	return err
}

// GetServiceDependency retrieves a service dependency by ID.
func (r *PostgresRepository) GetServiceDependency(ctx context.Context, dependencyID string) (*ServiceDependency, error) {
	var dep ServiceDependency
	err := r.db.QueryRowContext(ctx, `
		SELECT id, service_name, depends_on, dependency_type, description, is_active, created_at, updated_at
		FROM service_dependencies
		WHERE id = $1`, dependencyID).Scan(
		&dep.ID, &dep.ServiceName, &dep.DependsOn, &dep.DependencyType,
		&dep.Description, &dep.IsActive, &dep.CreatedAt, &dep.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get dependency: %w", err)
	}
	return &dep, nil
}

// ListServiceDependencies retrieves all dependencies for a service.
func (r *PostgresRepository) ListServiceDependencies(ctx context.Context, serviceName string) ([]ServiceDependency, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, service_name, depends_on, dependency_type, description, is_active, created_at, updated_at
		FROM service_dependencies
		WHERE service_name = $1
		ORDER BY dependency_type, created_at`, serviceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []ServiceDependency
	for rows.Next() {
		var d ServiceDependency
		if err := rows.Scan(&d.ID, &d.ServiceName, &d.DependsOn, &d.DependencyType,
			&d.Description, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

// GetDependenciesByService retrieves services that depend on a given service.
func (r *PostgresRepository) GetDependenciesByService(ctx context.Context, serviceName string) ([]ServiceDependency, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, service_name, depends_on, dependency_type, description, is_active, created_at, updated_at
		FROM service_dependencies
		WHERE depends_on = $1
		ORDER BY dependency_type, created_at`, serviceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []ServiceDependency
	for rows.Next() {
		var d ServiceDependency
		if err := rows.Scan(&d.ID, &d.ServiceName, &d.DependsOn, &d.DependencyType,
			&d.Description, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

// DeleteServiceDependency deletes a service dependency.
func (r *PostgresRepository) DeleteServiceDependency(ctx context.Context, dependencyID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM service_dependencies WHERE id = $1`, dependencyID)
	return err
}

// UpdateServiceDependency updates a service dependency's active status.
func (r *PostgresRepository) UpdateServiceDependency(ctx context.Context, dependencyID string, isActive bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE service_dependencies
		SET is_active = $1, updated_at = NOW()
		WHERE id = $2`, isActive, dependencyID)
	return err
}

// ============================================================================
// Status Snapshot Methods
// ============================================================================

// SaveStatusSnapshot inserts a SystemStatusSnapshot.
func (r *PostgresRepository) SaveStatusSnapshot(ctx context.Context, snapshot *SystemStatusSnapshot) error {
	servicesJSON, err := json.Marshal(snapshot.Services)
	if err != nil {
		return fmt.Errorf("marshal services: %w", err)
	}
	metadataJSON, err := json.Marshal(snapshot.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO system_status_snapshots
			(id, overall_status, services, active_alert_count, resolved_alert_count,
			 uptime_percent, latency_ms, error_rate, metadata, snapshot_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		snapshot.ID, snapshot.OverallStatus, servicesJSON,
		snapshot.ActiveAlertCount, snapshot.ResolvedAlertCount,
		snapshot.UptimePercent, snapshot.LatencyMs, snapshot.ErrorRate,
		metadataJSON, snapshot.SnapshotTime, snapshot.CreatedAt)
	return err
}

// GetStatusSnapshot retrieves a status snapshot by ID.
func (r *PostgresRepository) GetStatusSnapshot(ctx context.Context, snapshotID string) (*SystemStatusSnapshot, error) {
	var snapshot SystemStatusSnapshot
	var servicesJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, overall_status, services, active_alert_count, resolved_alert_count,
		       uptime_percent, latency_ms, error_rate, metadata, snapshot_time, created_at
		FROM system_status_snapshots
		WHERE id = $1`, snapshotID).Scan(
		&snapshot.ID, &snapshot.OverallStatus, &servicesJSON,
		&snapshot.ActiveAlertCount, &snapshot.ResolvedAlertCount,
		&snapshot.UptimePercent, &snapshot.LatencyMs, &snapshot.ErrorRate,
		&metadataJSON, &snapshot.SnapshotTime, &snapshot.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get snapshot: %w", err)
	}

	if len(servicesJSON) > 0 {
		_ = json.Unmarshal(servicesJSON, &snapshot.Services)
	}
	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &snapshot.Metadata)
	}
	return &snapshot, nil
}

// GetLatestStatusSnapshot retrieves the most recent status snapshot.
func (r *PostgresRepository) GetLatestStatusSnapshot(ctx context.Context) (*SystemStatusSnapshot, error) {
	var snapshot SystemStatusSnapshot
	var servicesJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, overall_status, services, active_alert_count, resolved_alert_count,
		       uptime_percent, latency_ms, error_rate, metadata, snapshot_time, created_at
		FROM system_status_snapshots
		ORDER BY snapshot_time DESC
		LIMIT 1`).Scan(
		&snapshot.ID, &snapshot.OverallStatus, &servicesJSON,
		&snapshot.ActiveAlertCount, &snapshot.ResolvedAlertCount,
		&snapshot.UptimePercent, &snapshot.LatencyMs, &snapshot.ErrorRate,
		&metadataJSON, &snapshot.SnapshotTime, &snapshot.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest snapshot: %w", err)
	}

	if len(servicesJSON) > 0 {
		_ = json.Unmarshal(servicesJSON, &snapshot.Services)
	}
	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &snapshot.Metadata)
	}
	return &snapshot, nil
}

// ListStatusSnapshots retrieves a list of status snapshots.
func (r *PostgresRepository) ListStatusSnapshots(ctx context.Context, limit int) ([]SystemStatusSnapshot, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, overall_status, services, active_alert_count, resolved_alert_count,
		       uptime_percent, latency_ms, error_rate, metadata, snapshot_time, created_at
		FROM system_status_snapshots
		ORDER BY snapshot_time DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStatusSnapshots(rows)
}

// GetStatusSnapshotsByTimeRange retrieves status snapshots within a time range.
func (r *PostgresRepository) GetStatusSnapshotsByTimeRange(ctx context.Context, start, end time.Time) ([]SystemStatusSnapshot, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, overall_status, services, active_alert_count, resolved_alert_count,
		       uptime_percent, latency_ms, error_rate, metadata, snapshot_time, created_at
		FROM system_status_snapshots
		WHERE snapshot_time BETWEEN $1 AND $2
		ORDER BY snapshot_time ASC`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStatusSnapshots(rows)
}

// GetSystemStatusSummary computes a summary of the current system status.
func (r *PostgresRepository) GetSystemStatusSummary(ctx context.Context) (*SystemStatusSummary, error) {
	// Get latest snapshot
	snapshot, err := r.GetLatestStatusSnapshot(ctx)
	if err != nil || snapshot == nil {
		return nil, err
	}

	// Get alert counts
	alertCounts, err := r.GetAlertCountsBySeverity(ctx)
	if err != nil {
		return nil, err
	}

	// Get health check results for services
	servicesQuery := `
		SELECT DISTINCT service_name
		FROM service_health_checks
		WHERE is_active = true`

	rows, err := r.db.QueryContext(ctx, servicesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			continue
		}
		services = append(services, service)
	}

	// Count healthy services
	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0
	unknownCount := 0

	for _, service := range services {
		latest, err := r.GetLatestHealthCheckResult(ctx, service)
		if err != nil || latest == nil {
			unknownCount++
			continue
		}
		switch latest.Status {
		case HealthCheckStatusHealthy:
			healthyCount++
		case HealthCheckStatusDegraded:
			degradedCount++
		case HealthCheckStatusUnhealthy:
			unhealthyCount++
		default:
			unknownCount++
		}
	}

	totalAlerts := 0
	for _, count := range alertCounts {
		totalAlerts += count
	}

	return &SystemStatusSummary{
		OverallStatus:     snapshot.OverallStatus,
		TotalServices:     len(services),
		HealthyServices:   healthyCount,
		DegradedServices:  degradedCount,
		UnhealthyServices: unhealthyCount,
		UnknownServices:   unknownCount,
		ActiveAlerts:      totalAlerts,
		TotalAlerts:       totalAlerts,
		UptimePercent:     *snapshot.UptimePercent,
		AvgLatencyMs:      *snapshot.LatencyMs,
		ErrorRate:         *snapshot.ErrorRate,
		Timestamp:         snapshot.SnapshotTime,
	}, nil
}

// scanStatusSnapshots is a helper to scan status snapshots from rows.
func scanStatusSnapshots(rows *sql.Rows) ([]SystemStatusSnapshot, error) {
	var snapshots []SystemStatusSnapshot
	for rows.Next() {
		var s SystemStatusSnapshot
		var servicesJSON, metadataJSON []byte
		if err := rows.Scan(&s.ID, &s.OverallStatus, &servicesJSON,
			&s.ActiveAlertCount, &s.ResolvedAlertCount,
			&s.UptimePercent, &s.LatencyMs, &s.ErrorRate,
			&metadataJSON, &s.SnapshotTime, &s.CreatedAt); err != nil {
			return nil, err
		}
		if len(servicesJSON) > 0 {
			_ = json.Unmarshal(servicesJSON, &s.Services)
		}
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &s.Metadata)
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}
