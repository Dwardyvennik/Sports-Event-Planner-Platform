package usecase

import "github.com/dwardyvennik/sports-event-planner-platform/services/api-gateway/internal/domain"

type GatewayUseCase struct {
	upstreams []domain.Upstream
}

func NewGatewayUseCase(upstreams []domain.Upstream) *GatewayUseCase {
	return &GatewayUseCase{upstreams: upstreams}
}

func (u *GatewayUseCase) Status() domain.GatewayStatus {
	return domain.GatewayStatus{
		Service:   "api-gateway",
		Upstreams: append([]domain.Upstream(nil), u.upstreams...),
	}
}
