package protocol

import (
	"testing"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
)

func TestDecodeRequestMustParseValidMessage(t *testing.T) {
	decoder := NewBetMessageDecoder()

	request, protocolErr := decoder.DecodeRequest([]byte("agency=1|nombre=John|apellido=Doe|documento=1|nacimiento=2000-01-01|numero=7574\n"))
	if protocolErr != nil {
		t.Fatalf("unexpected protocol error: %+v", *protocolErr)
	}

	if request.Fields["agency"] != "1" {
		t.Fatalf("expected agency 1, got %s", request.Fields["agency"])
	}
	if request.Fields["numero"] != "7574" {
		t.Fatalf("expected numero 7574, got %s", request.Fields["numero"])
	}
}

func TestDecodeRequestMustRejectDuplicateFields(t *testing.T) {
	decoder := NewBetMessageDecoder()

	_, protocolErr := decoder.DecodeRequest([]byte("agency=1|agency=2\n"))
	if protocolErr == nil {
		t.Fatal("expected duplicate field error")
	}
	if protocolErr.Code != "malformed_message" {
		t.Fatalf("expected malformed_message, got %s", protocolErr.Code)
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
