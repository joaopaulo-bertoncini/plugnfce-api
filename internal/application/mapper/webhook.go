package mapper

import (
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
)

// WebhookMapper handles mapping between webhook entities and DTOs
type WebhookMapper struct{}

// NewWebhookMapper creates a new WebhookMapper
func NewWebhookMapper() *WebhookMapper {
	return &WebhookMapper{}
}

// ToWebhookDTO converts a Webhook entity to a WebhookDTO
func (m *WebhookMapper) ToWebhookDTO(webhook *entity.Webhook) *dto.WebhookDTO {
	// Convert events
	events := make([]dto.WebhookEvent, len(webhook.Events))
	for i, event := range webhook.Events {
		events[i] = dto.WebhookEvent(event)
	}

	return &dto.WebhookDTO{
		ID:          webhook.ID,
		CompanyID:   webhook.CompanyID,
		Name:        webhook.Name,
		Description: webhook.Description,
		URL:         webhook.URL,
		Method:      dto.HTTPMethod(webhook.Method),
		Status:      dto.WebhookStatus(webhook.Status),
		Events:      events,
		Headers:     dto.WebhookHeaders(webhook.Headers),
		Secret:      webhook.Secret,
		RetryConfig: dto.WebhookRetryConfig{
			MaxRetries:    webhook.RetryConfig.MaxRetries,
			RetryInterval: webhook.RetryConfig.RetryInterval,
			MaxInterval:   webhook.RetryConfig.MaxInterval,
		},
		TotalDeliveries:      webhook.TotalDeliveries,
		SuccessfulDeliveries: webhook.SuccessfulDeliveries,
		FailedDeliveries:     webhook.FailedDeliveries,
		CreatedAt:            webhook.CreatedAt,
		UpdatedAt:            webhook.UpdatedAt,
		LastDeliveryAt:       webhook.LastDeliveryAt,
	}
}

// ToWebhookEntity converts a WebhookDTO to a Webhook entity
func (m *WebhookMapper) ToWebhookEntity(webhook *dto.WebhookDTO) *entity.Webhook {
	// Convert events
	events := make([]entity.WebhookEvent, len(webhook.Events))
	for i, event := range webhook.Events {
		events[i] = entity.WebhookEvent(event)
	}

	return &entity.Webhook{
		ID:          webhook.ID,
		CompanyID:   webhook.CompanyID,
		Name:        webhook.Name,
		Description: webhook.Description,
		URL:         webhook.URL,
		Method:      entity.HTTPMethod(webhook.Method),
		Status:      entity.WebhookStatus(webhook.Status),
		Events:      events,
		Headers:     entity.WebhookHeaders(webhook.Headers),
		Secret:      webhook.Secret,
		RetryConfig: entity.WebhookRetryConfig{
			MaxRetries:    webhook.RetryConfig.MaxRetries,
			RetryInterval: webhook.RetryConfig.RetryInterval,
			MaxInterval:   webhook.RetryConfig.MaxInterval,
		},
		TotalDeliveries:      webhook.TotalDeliveries,
		SuccessfulDeliveries: webhook.SuccessfulDeliveries,
		FailedDeliveries:     webhook.FailedDeliveries,
		CreatedAt:            webhook.CreatedAt,
		UpdatedAt:            webhook.UpdatedAt,
		LastDeliveryAt:       webhook.LastDeliveryAt,
	}
}

// ToWebhookListDTO converts a slice of Webhook entities to WebhookListResponse
func (m *WebhookMapper) ToWebhookListDTO(webhooks []*entity.Webhook) dto.WebhookListResponse {
	dtos := make([]dto.WebhookDTO, len(webhooks))
	for i, webhook := range webhooks {
		dtos[i] = *m.ToWebhookDTO(webhook)
	}

	return dto.WebhookListResponse{
		Webhooks: dtos,
		Total:    len(dtos),
	}
}

// ToWebhookDeliveryDTO converts a WebhookDelivery entity to a WebhookDelivery DTO
func (m *WebhookMapper) ToWebhookDeliveryDTO(delivery *entity.WebhookDelivery) *dto.WebhookDelivery {
	return &dto.WebhookDelivery{
		ID:           delivery.ID,
		WebhookID:    delivery.WebhookID,
		Event:        dto.WebhookEvent(delivery.Event),
		Payload:      delivery.Payload,
		Attempt:      delivery.Attempt,
		StatusCode:   delivery.StatusCode,
		ResponseBody: delivery.ResponseBody,
		ErrorMessage: delivery.ErrorMessage,
		Succeeded:    delivery.Succeeded,
		DeliveredAt:  delivery.DeliveredAt,
		CreatedAt:    delivery.CreatedAt,
	}
}

// ToWebhookDeliveryListDTO converts a slice of WebhookDelivery entities to WebhookDeliveryListResponse
func (m *WebhookMapper) ToWebhookDeliveryListDTO(deliveries []*entity.WebhookDelivery) dto.WebhookDeliveryListResponse {
	dtos := make([]dto.WebhookDelivery, len(deliveries))
	for i, delivery := range deliveries {
		dtos[i] = *m.ToWebhookDeliveryDTO(delivery)
	}

	return dto.WebhookDeliveryListResponse{
		Deliveries: dtos,
		Total:      len(dtos),
	}
}
