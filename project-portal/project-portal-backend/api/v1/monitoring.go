package v1

import (
	"net/http"

	"carbon-scribe/project-portal/project-portal-backend/internal/monitoring"
	"carbon-scribe/project-portal/project-portal-backend/internal/monitoring/ingestion"

	"github.com/gin-gonic/gin"
)

// MonitoringHandler handles monitoring-related API endpoints.
type MonitoringHandler struct {
	service *monitoring.Service
}

// NewMonitoringHandler creates a new monitoring handler.
func NewMonitoringHandler(service *monitoring.Service) *MonitoringHandler {
	return &MonitoringHandler{service: service}
}

// IngestSatellite handles POST /api/v1/monitoring/satellite.
func (h *MonitoringHandler) IngestSatellite(c *gin.Context) {
	var req monitoring.IngestSatelliteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reading, err := h.service.IngestSatellite(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      reading.ID,
		"message": "satellite reading ingested successfully",
	})
}

// IngestWebhook handles POST /api/v1/monitoring/webhook.
func (h *MonitoringHandler) IngestWebhook(c *gin.Context) {
	var req monitoring.IngestWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate webhook request
	auth := ingestion.NewWebhookAuthenticator(ingestion.WebhookAuthConfig{
		APIKey:     c.GetHeader("X-API-Key"),
		HMACSecret: "",         // Will be configured from env
		AllowedIPs: []string{}, // Will be configured from env
	})
	if err := auth.Authenticate(c.Request); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	reading, err := h.service.IngestWebhook(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      reading.ID,
		"message": "webhook reading ingested successfully",
	})
}

// ListSatelliteReadings handles GET /api/v1/monitoring/satellite.
func (h *MonitoringHandler) ListSatelliteReadings(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		// Parse limit
	}

	readings, err := h.service.ListReadings(c.Request.Context(), projectID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"readings": readings,
		"count":    len(readings),
	})
}

// ListWebhookReadings handles GET /api/v1/monitoring/webhook.
func (h *MonitoringHandler) ListWebhookReadings(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		// Parse limit
	}

	metricName := c.Query("metric_name")
	source := c.Query("source")

	var readings []monitoring.WebhookReading
	var err error

	if metricName != "" {
		readings, err = h.service.ListWebhookReadingsByMetric(c.Request.Context(), projectID, metricName, limit)
	} else if source != "" {
		readings, err = h.service.ListWebhookReadingsBySource(c.Request.Context(), projectID, source, limit)
	} else {
		readings, err = h.service.ListWebhookReadings(c.Request.Context(), projectID, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"readings": readings,
		"count":    len(readings),
	})
}

// RegisterMonitoringRoutes registers all monitoring routes.
func RegisterMonitoringRoutes(r *gin.RouterGroup, handler *MonitoringHandler) {
	monitoring := r.Group("/monitoring")
	{
		monitoring.POST("/satellite", handler.IngestSatellite)
		monitoring.POST("/webhook", handler.IngestWebhook)
		monitoring.GET("/satellite", handler.ListSatelliteReadings)
		monitoring.GET("/webhook", handler.ListWebhookReadings)
	}
}
