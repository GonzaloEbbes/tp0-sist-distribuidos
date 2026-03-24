package protocol

import (
	"testing"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
)

type betRegistrarStub struct {
	request domain.BetBatchRequest
}

func (s *betRegistrarStub) Register(request domain.BetBatchRequest) domain.Response {
	s.request = request
	return domain.NewSuccessResponse()
}

type agencyFinisherStub struct {
	request domain.FinishAgencyRequest
}

func (s *agencyFinisherStub) Finish(request domain.FinishAgencyRequest) domain.Response {
	s.request = request
	return domain.NewSuccessResponse()
}

type winnersQuerierStub struct {
	request domain.WinnersQueryRequest
}

func (s *winnersQuerierStub) Query(request domain.WinnersQueryRequest) domain.Response {
	s.request = request
	return domain.NewSuccessResponse()
}

func TestProcessMustRouteBetBatchMessage(t *testing.T) {
	registrar := &betRegistrarStub{}
	processor := NewMessageProcessor(registrar, &agencyFinisherStub{}, &winnersQuerierStub{})

	response := processor.Process([]byte("type=bet_batch|agency=1|nombre=John|apellido=Doe|documento=1|nacimiento=2000-01-01|numero=7574\n"))
	if response.Status != "ok" {
		t.Fatalf("expected ok response, got %+v", response)
	}
	if len(registrar.request.Bets) != 1 {
		t.Fatalf("expected 1 bet, got %d", len(registrar.request.Bets))
	}
	if registrar.request.Bets[0]["agency"] != "1" {
		t.Fatalf("expected agency 1, got %s", registrar.request.Bets[0]["agency"])
	}
}

func TestProcessMustRejectDuplicateFields(t *testing.T) {
	processor := NewMessageProcessor(&betRegistrarStub{}, &agencyFinisherStub{}, &winnersQuerierStub{})

	response := processor.Process([]byte("type=bet_batch|agency=1|agency=2\n"))
	if response.Code != "malformed_message" {
		t.Fatalf("expected malformed_message, got %s", response.Code)
	}
}

func TestProcessMustRejectMissingType(t *testing.T) {
	processor := NewMessageProcessor(&betRegistrarStub{}, &agencyFinisherStub{}, &winnersQuerierStub{})

	response := processor.Process([]byte("agency=1|nombre=John|apellido=Doe|documento=1|nacimiento=2000-01-01|numero=7574\n"))
	if response.Code != "malformed_message" {
		t.Fatalf("expected malformed_message, got %s", response.Code)
	}
}

func TestProcessMustRouteFinishMessage(t *testing.T) {
	finisher := &agencyFinisherStub{}
	processor := NewMessageProcessor(&betRegistrarStub{}, finisher, &winnersQuerierStub{})

	response := processor.Process([]byte("type=finish|agency=1\n"))
	if response.Status != "ok" {
		t.Fatalf("expected ok response, got %+v", response)
	}
	if finisher.request.Agency != "1" {
		t.Fatalf("expected agency 1, got %s", finisher.request.Agency)
	}
}

func TestEncodeResponseMustKeepExpectedFormat(t *testing.T) {
	encoder := NewResponseEncoder()

	message, err := encoder.Encode(domain.NewErrorResponse("missing_field", "documento requerido"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "status=error|code=missing_field|message=documento requerido\n"
	if string(message) != expected {
		t.Fatalf("expected %q, got %q", expected, string(message))
	}
}
