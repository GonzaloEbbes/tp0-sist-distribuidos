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

func (e *BetMessageEncoder) EncodeBet(bet domain.BetRequest, maxBatchAmount int, maxBatchSize int) ([][]byte, error) {
	if maxBatchAmount <= 0 {
		return nil, fmt.Errorf("invalid max batch amount")
	}

	batches := make([][]byte, 0)
	currentBatch := make([]string, 0, maxBatchAmount)

	for _, singleBet := range bet.Bets {
		encodedBet, err := encodeSingleBet(singleBet)
		if err != nil {
			return nil, err
		}

		candidateBatch := append(currentBatch, encodedBet)
		candidateMessage := []byte(strings.Join(candidateBatch, ",") + string(messageDelimiter))
		if len(candidateMessage) > maxBatchSize {
			if len(currentBatch) == 0 {
				return nil, fmt.Errorf("bet exceeds maximum batch size")
			}

			batches = append(batches, []byte(strings.Join(currentBatch, ",")+string(messageDelimiter)))
			currentBatch = []string{encodedBet}
			continue
		}

		currentBatch = candidateBatch
		if len(currentBatch) == maxBatchAmount {
			batches = append(batches, []byte(strings.Join(currentBatch, ",")+string(messageDelimiter)))
			currentBatch = make([]string, 0, maxBatchAmount)
		}
	}

	if len(currentBatch) > 0 {
		batches = append(batches, []byte(strings.Join(currentBatch, ",")+string(messageDelimiter)))
	}

	return batches, nil
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

func encodeSingleBet(singleBet domain.Bet) (string, error) {
	fields := []struct {
		key   string
		value string
	}{
		{key: "agency", value: singleBet.Agency},
		{key: "nombre", value: singleBet.FirstName},
		{key: "apellido", value: singleBet.LastName},
		{key: "documento", value: singleBet.Document},
		{key: "nacimiento", value: singleBet.Birthdate},
		{key: "numero", value: singleBet.Number},
	}

	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		if containsReservedChar(field.value) {
			return "", fmt.Errorf("invalid value for field %s", field.key)
		}
		parts = append(parts, fmt.Sprintf("%s=%s", field.key, field.value))
	}

	return strings.Join(parts, "|"), nil
}
