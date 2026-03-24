package usecase

import (
	"strconv"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
)

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

	uc.drawState.FinishAgency(agency)
	return domain.NewSuccessResponse()
}
