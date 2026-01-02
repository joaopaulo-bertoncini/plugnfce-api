package ports

import (
	"context"
	"time"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
)

// CompanyRepository defines the persistence boundary for companies.
type CompanyRepository interface {
	Create(ctx context.Context, company *entity.Company) error
	GetByID(ctx context.Context, id string) (*entity.Company, error)
	GetByCNPJ(ctx context.Context, cnpj string) (*entity.Company, error)
	Update(ctx context.Context, company *entity.Company) error
	List(ctx context.Context, limit, offset int) ([]*entity.Company, int, error)
	Count(ctx context.Context) (int, error)
	CountByStatus(ctx context.Context, status entity.CompanyStatus) (int, error)

	// Certificate methods
	GetCertificateByCompanyID(ctx context.Context, companyID string) (*entity.Certificate, error)

	// NFC-e sequencing methods
	GetNextNFCeNumber(ctx context.Context, companyID string) (int64, error)
}

// PlanRepository defines the persistence boundary for plans.
type PlanRepository interface {
	Create(ctx context.Context, plan *entity.Plan) error
	GetByID(ctx context.Context, id string) (*entity.Plan, error)
	Update(ctx context.Context, plan *entity.Plan) error
	List(ctx context.Context, limit, offset int) ([]*entity.Plan, int, error)
	Count(ctx context.Context) (int, error)
}

// SubscriptionRepository defines the persistence boundary for subscriptions.
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *entity.Subscription) error
	GetByID(ctx context.Context, id string) (*entity.Subscription, error)
	GetActiveByCompanyID(ctx context.Context, companyID string) (*entity.Subscription, error)
	Update(ctx context.Context, subscription *entity.Subscription) error
	List(ctx context.Context, limit, offset int) ([]*entity.Subscription, int, error)
	Count(ctx context.Context) (int, error)
	CountByStatus(ctx context.Context, status entity.SubscriptionStatus) (int, error)
}

// WebhookRepository defines the persistence boundary for webhooks.
type WebhookRepository interface {
	Create(ctx context.Context, webhook *entity.Webhook) error
	GetByID(ctx context.Context, id string) (*entity.Webhook, error)
	Update(ctx context.Context, webhook *entity.Webhook) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*entity.Webhook, int, error)
	ListByCompanyID(ctx context.Context, companyID string, limit, offset int) ([]*entity.Webhook, int, error)
	Count(ctx context.Context) (int, error)
}

// NFCeRepository defines the persistence boundary for NFC-e requests.
type NFCeRepository interface {
	Create(ctx context.Context, req *entity.NFCE) error
	Update(ctx context.Context, nfce *entity.NFCE) error
	UpdateFields(ctx context.Context, id string, updates map[string]interface{}) error
	UpdateStatus(ctx context.Context, id string, from entity.RequestStatus, to entity.RequestStatus, mutate func(*entity.NFCE)) error
	GetByID(ctx context.Context, id string) (*entity.NFCE, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*entity.NFCE, error)
	List(ctx context.Context, limit, offset int) ([]*entity.NFCE, error)
	ListWithFilters(ctx context.Context, limit, offset int, companyID, status string) ([]*entity.NFCE, int, error)
	GetStats(ctx context.Context, companyID string, since time.Time) (map[string]int, error)
	Count(ctx context.Context) (int, error)
	CountByStatus(ctx context.Context, status entity.RequestStatus) (int, error)
	AppendEvent(ctx context.Context, evt *entity.Event) error
	CreateEvent(ctx context.Context, event *entity.Event) error
	GetEventsByRequestID(ctx context.Context, requestID string, limit, offset int) ([]*entity.Event, error)
	GetPendingRetries(ctx context.Context, beforeTime time.Time, limit int) ([]*entity.NFCE, error)
}

// Tx defines the minimal transaction contract used by the service layer.
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
