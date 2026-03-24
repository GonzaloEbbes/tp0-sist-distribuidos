package usecase

import (
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/common"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
)

type QueryWinners struct {
	drawState *DrawState
}

func NewQueryWinners(drawState *DrawState) *QueryWinners {
	return &QueryWinners{drawState: drawState}
}

func (uc *QueryWinners) Query(request domain.WinnersQueryRequest) domain.Response {
	agency, err := strconv.Atoi(request.Agency)
	if err != nil {
		return domain.NewErrorResponse("invalid_field_format", err.Error())
	}
	if !uc.drawState.IsReady() {
		return domain.NewErrorResponse("draw_not_ready", "draw is not ready yet")
	}

	storedBets, err := common.LoadBets()
	if err != nil {
		return domain.NewErrorResponse("storage_error", err.Error())
	}

	winners := make([]string, 0)
	for _, bet := range storedBets {
		if bet.Agency != agency {
			continue
		}
		if common.HasWon(bet) {
			winners = append(winners, bet.Document)
		}
	}

	return domain.Response{
		Status:  "ok",
		Message: strings.Join(winners, ","),
	}
}
