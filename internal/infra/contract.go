package infra

import (
	"context"
	"io"
	"time"
)

// =============================================================================
// Core Infrastructure Contracts (Ports)
// clean-arch: These interfaces define contracts that infrastructure must implement
// =============================================================================

// Connection defines the unified contract for all connectable infrastructure.
// Both DatabaseClient and RedisClient implement this interface.
type Connection interface {
	io.Closer
	// Ping checks if the connection is alive
	Ping(ctx context.Context) error
	// IsHealthy performs a health check with timeout
	IsHealthy(ctx context.Context) bool
	// Name returns the identifier for this connection
	Name() string
}

// =============================================================================
// Database-Specific Contracts
// =============================================================================

// TransactionalDB extends Connection with transaction support.
// clean-arch: Port for transactional database operations
type TransactionalDB interface {
	Connection
	BeginTx(ctx context.Context) (Transaction, error)
	Stats() ConnectionStats
}

// Transaction represents a database transaction.
type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// ConnectionStats represents connection pool statistics.
type ConnectionStats struct {
	MaxConnections      int32 `json:"max_connections"`
	CurrentConnections  int32 `json:"current_connections"`
	IdleConnections     int32 `json:"idle_connections"`
	AcquiredConnections int32 `json:"acquired_connections"`
}

// =============================================================================
// Cache-Specific Contracts
// =============================================================================

// CacheStore extends Connection with cache operations.
// clean-arch: Port for cache operations
type CacheStore interface {
	Connection
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// =============================================================================
// Health Check Contracts
// =============================================================================

// HealthChecker defines the contract for health checking.
type HealthChecker interface {
	Check(ctx context.Context) HealthStatus
}

// HealthStatus represents the health status of a component.
type HealthStatus struct {
	Name      string        `json:"name"`
	Status    Status        `json:"status"`
	Latency   time.Duration `json:"latency_ms"`
	Message   string        `json:"message,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// Status represents the health status enum.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// =============================================================================
// Message Queue Contracts (for future integration)
// =============================================================================

// MessagePublisher defines the contract for publishing messages.
// clean-arch: Port for message queue publishing (Kafka, RabbitMQ, etc.)
type MessagePublisher interface {
	Connection
	Publish(ctx context.Context, topic string, message []byte) error
	PublishWithKey(ctx context.Context, topic, key string, message []byte) error
}

// MessageSubscriber defines the contract for subscribing to messages.
// clean-arch: Port for message queue subscription
type MessageSubscriber interface {
	Connection
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Unsubscribe(topic string) error
}

// MessageHandler is the callback function for handling received messages.
type MessageHandler func(ctx context.Context, message Message) error

// Message represents a message from a message queue.
type Message struct {
	ID        string            `json:"id"`
	Topic     string            `json:"topic"`
	Key       string            `json:"key,omitempty"`
	Value     []byte            `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// =============================================================================
// External HTTP Client Contracts
// =============================================================================

// HTTPClient defines the contract for making HTTP requests to external services.
// clean-arch: Port for external HTTP service integration
type HTTPClient interface {
	Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
	Post(ctx context.Context, url string, body []byte, headers map[string]string) (*HTTPResponse, error)
	Put(ctx context.Context, url string, body []byte, headers map[string]string) (*HTTPResponse, error)
	Delete(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error)
}

// HTTPResponse represents an HTTP response from an external service.
type HTTPResponse struct {
	StatusCode int               `json:"status_code"`
	Body       []byte            `json:"body"`
	Headers    map[string]string `json:"headers"`
}

// =============================================================================
// Object Storage Contracts (for future integration)
// =============================================================================

// ObjectStorage defines the contract for object storage operations.
// clean-arch: Port for object storage (S3, GCS, MinIO, etc.)
type ObjectStorage interface {
	Connection
	Upload(ctx context.Context, bucket, key string, data io.Reader, contentType string) error
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
}

// =============================================================================
// Notification Contracts (for future integration)
// =============================================================================

// NotificationSender defines the contract for sending notifications.
// clean-arch: Port for notification services (email, SMS, push)
type NotificationSender interface {
	SendEmail(ctx context.Context, req EmailRequest) error
	SendSMS(ctx context.Context, req SMSRequest) error
	SendPush(ctx context.Context, req PushRequest) error
}

// EmailRequest represents an email notification request.
type EmailRequest struct {
	To          []string     `json:"to"`
	CC          []string     `json:"cc,omitempty"`
	BCC         []string     `json:"bcc,omitempty"`
	Subject     string       `json:"subject"`
	Body        string       `json:"body"`
	IsHTML      bool         `json:"is_html"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// SMSRequest represents an SMS notification request.
type SMSRequest struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

// PushRequest represents a push notification request.
type PushRequest struct {
	DeviceTokens []string          `json:"device_tokens"`
	Title        string            `json:"title"`
	Body         string            `json:"body"`
	Data         map[string]string `json:"data,omitempty"`
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}
