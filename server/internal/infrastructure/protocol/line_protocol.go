package protocol

import (
	"fmt"
	"strings"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
)

const messageDelimiter = '\n'

var reservedValueChars = []string{"|", "=", ",", "\n"}

type BetMessageDecoder struct{}

type ResponseEncoder struct{}

func NewBetMessageDecoder() *BetMessageDecoder {
	return &BetMessageDecoder{}
}

func NewResponseEncoder() *ResponseEncoder {
	return &ResponseEncoder{}
}

func (d *BetMessageDecoder) DecodeRequest(message []byte) (domain.BetRequest, *domain.Response) {
	fields, err := decodeKeyValueMessage(message)
	if err != nil {
		errorResponse := domain.NewErrorResponse("malformed_message", err.Error())
		return domain.BetRequest{}, &errorResponse
	}

	if len(fields) == 0 {
		errorResponse := domain.NewErrorResponse("malformed_message", "empty message")
		return domain.BetRequest{}, &errorResponse
	}

	messageType, ok := fields[0]["type"]
	if !ok || messageType == "" {
		errorResponse := domain.NewErrorResponse("malformed_message", "missing type in request")
		return domain.BetRequest{}, &errorResponse
	}

	return buildRequestByType(messageType, fields)
}

func (e *ResponseEncoder) Encode(response domain.Response) ([]byte, error) {
	fields := []struct {
		key   string
		value string
	}{
		{key: "status", value: response.Status},
	}

	if response.Code != "" {
		fields = append(fields, struct {
			key   string
			value string
		}{key: "code", value: response.Code})
	}
	if response.Message != "" {
		fields = append(fields, struct {
			key   string
			value string
		}{key: "message", value: response.Message})
	}

	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		if containsReservedChar(field.value) {
			return nil, fmt.Errorf("invalid response value for field %s", field.key)
		}
		parts = append(parts, fmt.Sprintf("%s=%s", field.key, field.value))
	}

	return []byte(strings.Join(parts, "|") + string(messageDelimiter)), nil
}

func decodeKeyValueMessage(message []byte) ([]domain.BetAttempt, error) {
	if len(message) == 0 || message[len(message)-1] != messageDelimiter {
		return nil, fmt.Errorf("message does not end with delimiter")
	}

	payload := string(message[:len(message)-1])
	if payload == "" {
		return nil, fmt.Errorf("empty message")
	}

	registers := strings.Split(payload, ",")
	bets := make([]domain.BetAttempt, 0, len(registers))
	for index, register := range registers {
		tokens := strings.Split(register, "|")
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
			if containsReservedChar(parts[1]) {
				return nil, fmt.Errorf("invalid value for field %s", parts[0])
			}
			fields[parts[0]] = parts[1]
		}
		if index >= len(bets) {
			bets = append(bets, fields)
			continue
		}
		bets[index] = fields
	}

	return bets, nil
}

func containsReservedChar(value string) bool {
	for _, reservedChar := range reservedValueChars {
		if strings.Contains(value, reservedChar) {
			return true
		}
	}
	return false
}

func buildRequestByType(messageType string, fields []domain.BetAttempt) (domain.BetRequest, *domain.Response) {
	// This would probably fit better in another layer. A use case whose only job is
	// switching by message type does not sound especially appropriate, and adding a
	// dedicated handler layer right now would be mostly boilerplate for too little
	// behavior. If the amount of message types grows, this should likely move.
	switch messageType {
	case domain.MessageTypeBetBatch:
		delete(fields[0], "type")
		return domain.BetRequest{Type: messageType, Bets: fields}, nil
	default:
		errorResponse := domain.NewErrorResponse("malformed_message", fmt.Sprintf("unsupported message type %s", messageType))
		return domain.BetRequest{}, &errorResponse
	}
}
