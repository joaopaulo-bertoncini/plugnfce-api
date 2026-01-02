package mapper

import (
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
)

// NFceMapper handles mapping between NFC-e entities and DTOs
type NFceMapper struct{}

// NewNFceMapper creates a new NFceMapper
func NewNFceMapper() *NFceMapper {
	return &NFceMapper{}
}

// ToEmitPayload converts EmitNFceRequest to EmitPayload
func (m *NFceMapper) ToEmitPayload(req dto.EmitNFceRequest) entity.EmitPayload {
	// Convert items
	itens := make([]entity.Item, len(req.Itens))
	for i, item := range req.Itens {
		itens[i] = entity.Item{
			Descricao:  item.Descricao,
			NCM:        item.NCM,
			CFOP:       item.CFOP,
			GTIN:       item.GTIN,
			Valor:      item.Valor,
			Quantidade: item.Quantidade,
			Unidade:    item.Unidade,
		}
	}

	// Convert payments
	pagamentos := make([]entity.Payment, len(req.Pagamentos))
	for i, payment := range req.Pagamentos {
		pagamentos[i] = entity.Payment{
			Forma: payment.Forma,
			Valor: payment.Valor,
			Troco: payment.Troco,
		}
	}

	return entity.EmitPayload{
		UF:       req.UF,
		Ambiente: req.Ambiente,
		Emitente: entity.Emitente{
			CNPJ:     req.Emitente.CNPJ,
			IE:       req.Emitente.IE,
			Regime:   req.Emitente.Regime,
			CSCID:    req.Emitente.CSCID,
			CSCToken: req.Emitente.CSCToken,
		},
		Itens:      itens,
		Pagamentos: pagamentos,
		Options: entity.EmitOptions{
			Contingencia: req.Options.Contingencia,
			Sync:         req.Options.Sync,
		},
	}
}

// ToResponse converts Request entity to NFceResponse
func (m *NFceMapper) ToResponse(req *entity.Request) dto.NFceResponse {
	return dto.NFceResponse{
		ID:             req.ID,
		IdempotencyKey: req.IdempotencyKey,
		Status:         dto.RequestStatus(req.Status),
		ChaveAcesso:    req.ChaveAcesso,
		Protocolo:      req.Protocolo,
		RejectionCode:  req.RejectionCode,
		RejectionMsg:   req.RejectionMsg,
		RetryCount:     req.RetryCount,
		NextRetryAt:    req.NextRetryAt,
		CreatedAt:      req.CreatedAt,
		UpdatedAt:      req.UpdatedAt,
	}
}

// ToResponseList converts a slice of Request entities to NFceListResponse
func (m *NFceMapper) ToResponseList(requests []*entity.Request) dto.NFceListResponse {
	responses := make([]dto.NFceResponse, len(requests))
	for i, req := range requests {
		responses[i] = m.ToResponse(req)
	}

	return dto.NFceListResponse{
		NFces: responses,
		Total: len(responses),
	}
}

// ToEventResponse converts Event entity to NFceEventResponse
func (m *NFceMapper) ToEventResponse(event *entity.Event) dto.NFceEventResponse {
	return dto.NFceEventResponse{
		ID:         event.ID,
		RequestID:  event.RequestID,
		StatusFrom: dto.RequestStatus(event.StatusFrom),
		StatusTo:   dto.RequestStatus(event.StatusTo),
		CStat:      event.CStat,
		Message:    event.Message,
		CreatedAt:  event.CreatedAt,
	}
}

// ToEventResponseList converts a slice of Event entities to NFceEventListResponse
func (m *NFceMapper) ToEventResponseList(events []*entity.Event) dto.NFceEventListResponse {
	responses := make([]dto.NFceEventResponse, len(events))
	for i, event := range events {
		responses[i] = m.ToEventResponse(event)
	}

	return dto.NFceEventListResponse{
		Events: responses,
		Total:  len(responses),
	}
}
