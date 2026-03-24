package usecase

import (
	"strconv"
	"strings"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/common"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
)

var queryWinnersLog = logging.MustGetLogger("log")

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
	queryWinnersLog.Infof("action: consulta_ganadores | result: in_progress | agency: %d", agency)
	if !uc.drawState.IsReady() {
		queryWinnersLog.Infof(
			"action: consulta_ganadores | result: retry | agency: %d | finished_agencies: %d | expected_agencies: %d",
			agency,
			len(uc.drawState.finishedAgencies),
			uc.drawState.expectedAgencies,
		)
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

	queryWinnersLog.Infof(
		"action: consulta_ganadores | result: ready | agency: %d | cant_ganadores: %d",
		agency,
		len(winners),
	)
	return domain.Response{
		Status:  "ok",
		Message: strings.Join(winners, ","),
	}
}
