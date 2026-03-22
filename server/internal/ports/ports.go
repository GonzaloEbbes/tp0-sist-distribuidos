package ports

import "github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"

type BetMessageDecoder interface {
	DecodeRequest([]byte) (domain.BetRequest, *domain.Response)
}

type ResponseEncoder interface {
	Encode(domain.Response) ([]byte, error)
}

type BetRegistrar interface {
	Register(domain.BetRequest) domain.Response
}

type BetRepository interface {
	Store(domain.Bet) error
}
