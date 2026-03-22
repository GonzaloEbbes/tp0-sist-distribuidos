package usecase

import (
	"testing"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
)

type repositoryStub struct {
	stored []domain.Bet
	err    error
}

func (r *repositoryStub) Store(bet domain.Bet) error {
	if r.err != nil {
		return r.err
	}
	r.stored = append(r.stored, bet)
	return nil
}

func TestRegisterMustPersistValidBet(t *testing.T) {
	repository := &repositoryStub{}
	useCase := NewRegisterBet(repository)

	response := useCase.Register(domain.BetRequest{
		Fields: map[string]string{
			"agency":     "1",
			"nombre":     "John",
			"apellido":   "Doe",
			"documento":  "123",
			"nacimiento": "2000-01-01",
			"numero":     "7574",
		},
	})

	if response.Status != "ok" {
		t.Fatalf("expected ok response, got %+v", response)
	}
	if len(repository.stored) != 1 {
		t.Fatalf("expected 1 stored bet, got %d", len(repository.stored))
	}
}

func TestRegisterMustRejectUnknownField(t *testing.T) {
	repository := &repositoryStub{}
	useCase := NewRegisterBet(repository)

	response := useCase.Register(domain.BetRequest{
		Fields: map[string]string{
			"agency":     "1",
			"nombre":     "John",
			"apellido":   "Doe",
			"documento":  "123",
			"nacimiento": "2000-01-01",
			"numero":     "7574",
			"extra":      "value",
		},
	})

	if response.Code != "unknown_field" {
		t.Fatalf("expected unknown_field, got %+v", response)
	}
	if len(repository.stored) != 0 {
		t.Fatalf("expected no stored bets, got %d", len(repository.stored))
	}
}
