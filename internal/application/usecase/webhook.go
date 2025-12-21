package usecase

import (
	"context"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/mapper"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
)

// WebhookUseCase defines the interface for webhook operations
type WebhookUseCase interface {
	Create(ctx context.Context, req dto.CreateWebhookRequest) (*dto.WebhookDTO, error)
	GetByID(ctx context.Context, id string) (*dto.WebhookDTO, error)
	List(ctx context.Context, companyID string, limit, offset int) (*dto.WebhookListResponse, error)
	Update(ctx context.Context, id string, req dto.UpdateWebhookRequest) error
	Delete(ctx context.Context, id string) error
}

// WebhookUseCaseImpl handles webhook operations
type WebhookUseCaseImpl struct {
	webhookRepo   ports.WebhookRepository
	webhookMapper *mapper.WebhookMapper
}

// NewWebhookUseCase creates a new WebhookUseCase
func NewWebhookUseCase(webhookRepo ports.WebhookRepository) WebhookUseCase {
	return &WebhookUseCaseImpl{
		webhookRepo:   webhookRepo,
		webhookMapper: mapper.NewWebhookMapper(),
	}
}

// Create creates a new webhook
func (uc *WebhookUseCaseImpl) Create(ctx context.Context, req dto.CreateWebhookRequest) (*dto.WebhookDTO, error) {
	// Convert events from DTO to entity
	events := make([]entity.WebhookEvent, len(req.Events))
	for i, event := range req.Events {
		events[i] = entity.WebhookEvent(event)
	}

	webhook, err := entity.NewWebhook(req.CompanyID, req.Name, req.URL, events)
	if err != nil {
		return nil, err
	}

	// Apply additional fields
	webhook.Description = req.Description
	webhook.Method = entity.HTTPMethod(req.Method)
	webhook.Headers = entity.WebhookHeaders(req.Headers)
	webhook.Secret = req.Secret
	if req.RetryConfig != nil {
		webhook.RetryConfig = entity.WebhookRetryConfig{
			MaxRetries:    req.RetryConfig.MaxRetries,
			RetryInterval: req.RetryConfig.RetryInterval,
			MaxInterval:   req.RetryConfig.MaxInterval,
		}
	}

	err = uc.webhookRepo.Create(ctx, webhook)
	if err != nil {
		return nil, err
	}

	return uc.webhookMapper.ToWebhookDTO(webhook), nil
}

// GetByID gets a webhook by ID
func (uc *WebhookUseCaseImpl) GetByID(ctx context.Context, id string) (*dto.WebhookDTO, error) {
	webhook, err := uc.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.webhookMapper.ToWebhookDTO(webhook), nil
}

// List lists webhooks for a company
func (uc *WebhookUseCaseImpl) List(ctx context.Context, companyID string, limit, offset int) (*dto.WebhookListResponse, error) {
	// TODO: Filter by companyID when repository supports it
	webhooks, total, err := uc.webhookRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	response := uc.webhookMapper.ToWebhookListDTO(webhooks)
	response.Total = total
	return &response, nil
}

// Update updates a webhook
func (uc *WebhookUseCaseImpl) Update(ctx context.Context, id string, req dto.UpdateWebhookRequest) error {
	webhook, err := uc.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates from request
	if req.Name != nil {
		webhook.Name = *req.Name
	}
	if req.Description != nil {
		webhook.Description = *req.Description
	}
	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.Method != nil {
		webhook.Method = entity.HTTPMethod(*req.Method)
	}
	if req.Status != nil {
		webhook.Status = entity.WebhookStatus(*req.Status)
	}
	if len(req.Events) > 0 {
		events := make([]entity.WebhookEvent, len(req.Events))
		for i, event := range req.Events {
			events[i] = entity.WebhookEvent(event)
		}
		webhook.Events = events
	}
	if req.Headers != nil {
		webhook.Headers = entity.WebhookHeaders(req.Headers)
	}
	if req.Secret != nil {
		webhook.Secret = *req.Secret
	}
	if req.RetryConfig != nil {
		webhook.RetryConfig = entity.WebhookRetryConfig{
			MaxRetries:    req.RetryConfig.MaxRetries,
			RetryInterval: req.RetryConfig.RetryInterval,
			MaxInterval:   req.RetryConfig.MaxInterval,
		}
	}

	return uc.webhookRepo.Update(ctx, webhook)
}

// Delete deletes a webhook
func (uc *WebhookUseCaseImpl) Delete(ctx context.Context, id string) error {
	return uc.webhookRepo.Delete(ctx, id)
}
