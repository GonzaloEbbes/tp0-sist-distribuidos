package usecase

import (
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/ports"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/common"
)

const winnersSeparator = ";"

type QueryWinners struct {
	drawState *DrawState
	repository ports.BetRepository
}

func NewQueryWinners(drawState *DrawState, repository ports.BetRepository) *QueryWinners {
	return &QueryWinners{
		drawState: drawState,
		repository: repository,
	}
}

func (uc *QueryWinners) Query(request domain.WinnersQueryRequest) domain.Response {
	agency, err := strconv.Atoi(request.Agency)
	if err != nil {
		return domain.NewErrorResponse("invalid_field_format", err.Error())
	}
	if !uc.drawState.IsReady() {
		return domain.NewErrorResponse("draw_not_ready", "draw is not ready yet")
	}

	storedBets, err := uc.repository.LoadBets()
	if err != nil {
		return domain.NewErrorResponse("storage_error", err.Error())
	}

	winners := make([]string, 0)
	for _, bet := range storedBets {
		if bet.Agency != agency {
			continue
		}
		if common.HasWon(common.Bet{
			Agency:    bet.Agency,
			FirstName: bet.FirstName,
			LastName:  bet.LastName,
			Document:  bet.Document,
			Birthdate: bet.Birthdate,
			Number:    bet.Number,
		}) {
			winners = append(winners, bet.Document)
		}
	}

	return domain.Response{
		Status:  "ok",
		Message: strings.Join(winners, winnersSeparator),
	}
}
