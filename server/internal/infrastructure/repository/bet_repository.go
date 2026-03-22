package repository

import (
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/common"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
)

type BetRepository struct{}

func NewBetRepository() *BetRepository {
	return &BetRepository{}
}

func (r *BetRepository) Store(bet domain.Bet) error {
	return common.StoreBets([]common.Bet{
		{
			Agency:    bet.Agency,
			FirstName: bet.FirstName,
			LastName:  bet.LastName,
			Document:  bet.Document,
			Birthdate: bet.Birthdate,
			Number:    bet.Number,
		},
	})
}
