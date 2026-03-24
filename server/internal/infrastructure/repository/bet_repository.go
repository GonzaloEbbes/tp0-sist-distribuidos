package repository

import (
	"sync"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/common"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
)

type BetRepository struct {
	mu sync.RWMutex
}

func NewBetRepository() *BetRepository {
	return &BetRepository{}
}

// Deprecated: use StoreBatch instead.
func (r *BetRepository) Store(bet domain.Bet) error {
	return r.StoreBatch([]domain.Bet{bet})
}

func (r *BetRepository) StoreBatch(bets []domain.Bet) error {
	commonBets := make([]common.Bet, 0, len(bets))
	for _, bet := range bets {
		commonBets = append(commonBets, common.Bet{
			Agency:    bet.Agency,
			FirstName: bet.FirstName,
			LastName:  bet.LastName,
			Document:  bet.Document,
			Birthdate: bet.Birthdate,
			Number:    bet.Number,
		})
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return common.StoreBets(commonBets)
}

func (r *BetRepository) LoadBets() ([]domain.Bet, error) {
	r.mu.RLock()
	commonBets, err := common.LoadBets()
	r.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	bets := make([]domain.Bet, 0, len(commonBets))
	for _, bet := range commonBets {
		bets = append(bets, domain.Bet{
			Agency:    bet.Agency,
			FirstName: bet.FirstName,
			LastName:  bet.LastName,
			Document:  bet.Document,
			Birthdate: bet.Birthdate,
			Number:    bet.Number,
		})
	}

	return bets, nil
}
