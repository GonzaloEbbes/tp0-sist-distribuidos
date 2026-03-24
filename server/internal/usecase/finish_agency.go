package usecase

import (
	"strconv"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
)

var finishAgencyLog = logging.MustGetLogger("log")

type FinishAgency struct {
	drawState *DrawState
}

func NewFinishAgency(drawState *DrawState) *FinishAgency {
	return &FinishAgency{drawState: drawState}
}

func (uc *FinishAgency) Finish(request domain.FinishAgencyRequest) domain.Response {
	agency, err := strconv.Atoi(request.Agency)
	if err != nil {
		return domain.NewErrorResponse("invalid_field_format", err.Error())
	}

	finishAgencyLog.Infof("action: fin_agencia | result: in_progress | agency: %d", agency)
	uc.drawState.FinishAgency(agency)
	finishAgencyLog.Infof(
		"action: fin_agencia | result: success | agency: %d | finished_agencies: %d | expected_agencies: %d",
		agency,
		len(uc.drawState.finishedAgencies),
		uc.drawState.expectedAgencies,
	)
	return domain.NewSuccessResponse()
}
