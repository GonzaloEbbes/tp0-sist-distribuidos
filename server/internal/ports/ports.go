package ports

import "github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"

type BetMessageDecoder interface {
	Process([]byte) domain.Response
}

type ResponseEncoder interface {
	Encode(domain.Response) ([]byte, error)
}

type BetRegistrar interface {
	Register(domain.BetBatchRequest) domain.Response
}

type AgencyFinisher interface {
	Finish(domain.FinishAgencyRequest) domain.Response
}

type WinnersQuerier interface {
	Query(domain.WinnersQueryRequest) domain.Response
}

type BetRepository interface {
	// Deprecated: use StoreBatch instead.
	Store(domain.Bet) error
	StoreBatch([]domain.Bet) error
}
