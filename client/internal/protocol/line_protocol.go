package protocol

import (
	"fmt"
	"strings"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/domain"
)

const messageDelimiter = '\n'

var reservedValueChars = []string{"|", "=", ",", "\n"}

type BetMessageEncoder struct{}

type ServerResponseDecoder struct{}

func NewBetMessageEncoder() *BetMessageEncoder {
	return &BetMessageEncoder{}
}

func NewServerResponseDecoder() *ServerResponseDecoder {
	return &ServerResponseDecoder{}
}

func (e *BetMessageEncoder) EncodeBet(bet domain.BetRequest) ([]byte, error) {
	fields := []struct {
		key   string
		value string
	}{
		{key: "agency", value: bet.Agency},
		{key: "nombre", value: bet.FirstName},
		{key: "apellido", value: bet.LastName},
		{key: "documento", value: bet.Document},
		{key: "nacimiento", value: bet.Birthdate},
		{key: "numero", value: bet.Number},
	}

	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		if containsReservedChar(field.value) {
			return nil, fmt.Errorf("invalid value for field %s", field.key)
		}
		parts = append(parts, fmt.Sprintf("%s=%s", field.key, field.value))
	}

	return []byte(strings.Join(parts, "|") + string(messageDelimiter)), nil
}

func (d *ServerResponseDecoder) DecodeResponse(message []byte) (domain.ServerResponse, error) {
	fields, err := decodeKeyValueMessage(message)
	if err != nil {
		return domain.ServerResponse{}, err
	}

	status, ok := fields["status"]
	if !ok {
		return domain.ServerResponse{}, fmt.Errorf("missing status in server response")
	}

	return domain.ServerResponse{
		Status:  status,
		Code:    fields["code"],
		Message: fields["message"],
	}, nil
}

func decodeKeyValueMessage(message []byte) (map[string]string, error) {
	if len(message) == 0 || message[len(message)-1] != messageDelimiter {
		return nil, fmt.Errorf("message does not end with delimiter")
	}

	payload := string(message[:len(message)-1])
	if payload == "" {
		return nil, fmt.Errorf("empty message")
	}

	tokens := strings.Split(payload, "|")
	fields := make(map[string]string, len(tokens))
	for _, token := range tokens {
		if token == "" {
			return nil, fmt.Errorf("empty token")
		}

		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("malformed token %q", token)
		}
		if _, exists := fields[parts[0]]; exists {
			return nil, fmt.Errorf("duplicate field %s", parts[0])
		}
		fields[parts[0]] = parts[1]
	}

	return fields, nil
}

func containsReservedChar(value string) bool {
	for _, reservedChar := range reservedValueChars {
		if strings.Contains(value, reservedChar) {
			return true
		}
	}
	return false
}
