package usecase

import (
	"fmt"
	"strconv"
	"time"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/ports"
)

var log = logging.MustGetLogger("log")

type RegisterBet struct {
	repository ports.BetRepository
}

func NewRegisterBet(repository ports.BetRepository) *RegisterBet {
	return &RegisterBet{repository: repository}
}

func (uc *RegisterBet) Register(request domain.BetRequest) domain.Response {
	requiredFields := []string{"agency", "nombre", "apellido", "documento", "nacimiento", "numero"}
	for _, field := range requiredFields {
		value, ok := request.Fields[field]
		if !ok || value == "" {
			return domain.NewErrorResponse("missing_field", fmt.Sprintf("%s is required", field))
		}
	}

	for field := range request.Fields {
		if !isAllowedField(field) {
			return domain.NewErrorResponse("unknown_field", fmt.Sprintf("unknown field %s", field))
		}
	}

	bet, err := buildDomainBet(request.Fields)
	if err != nil {
		return domain.NewErrorResponse("invalid_field_format", err.Error())
	}

	if err := uc.repository.Store(bet); err != nil {
		return domain.NewErrorResponse("storage_error", err.Error())
	}

	log.Infof(
		"action: apuesta_almacenada | result: success | dni: %s | numero: %s",
		request.Fields["documento"],
		request.Fields["numero"],
	)
	return domain.NewSuccessResponse()
}

func isAllowedField(field string) bool {
	switch field {
	case "agency", "nombre", "apellido", "documento", "nacimiento", "numero":
		return true
	default:
		return false
	}
}

func buildDomainBet(fields map[string]string) (domain.Bet, error) {
	agency, err := strconv.Atoi(fields["agency"])
	if err != nil {
		return domain.Bet{}, err
	}

	birthdate, err := time.Parse("2006-01-02", fields["nacimiento"])
	if err != nil {
		return domain.Bet{}, err
	}

	number, err := strconv.Atoi(fields["numero"])
	if err != nil {
		return domain.Bet{}, err
	}

	return domain.Bet{
		Agency:    agency,
		FirstName: fields["nombre"],
		LastName:  fields["apellido"],
		Document:  fields["documento"],
		Birthdate: birthdate,
		Number:    number,
	}, nil
}
