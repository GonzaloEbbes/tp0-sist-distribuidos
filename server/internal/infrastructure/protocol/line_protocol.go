package protocol

import (
	"fmt"
	"strings"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/ports"
)

const messageDelimiter = '\n'

var reservedValueChars = []string{"|", "=", ",", "\n"}
var protocolLog = logging.MustGetLogger("log")

type BetMessageDecoder struct{}
type MessageProcessor struct {
	betRegistrar   ports.BetRegistrar
	agencyFinisher ports.AgencyFinisher
	winnersQuerier ports.WinnersQuerier
}

type ResponseEncoder struct{}

func NewBetMessageDecoder() *BetMessageDecoder {
	return &BetMessageDecoder{}
}

func NewMessageProcessor(betRegistrar ports.BetRegistrar, agencyFinisher ports.AgencyFinisher, winnersQuerier ports.WinnersQuerier) *MessageProcessor {
	return &MessageProcessor{
		betRegistrar:   betRegistrar,
		agencyFinisher: agencyFinisher,
		winnersQuerier: winnersQuerier,
	}
}

func NewResponseEncoder() *ResponseEncoder {
	return &ResponseEncoder{}
}

func (p *MessageProcessor) Process(message []byte) domain.Response {
	messageType, fields, protocolErr := decodeMessage(message)
	if protocolErr != nil {
		protocolLog.Infof("action: process_message | result: fail | error_code: %s | error_message: %s", protocolErr.Code, protocolErr.Message)
		return *protocolErr
	}

	protocolLog.Infof("action: process_message | result: success | type: %s", messageType)
	return p.buildRequestByType(messageType, fields)
}

func decodeMessage(message []byte) (string, []domain.BetAttempt, *domain.Response) {
	fields, err := decodeKeyValueMessage(message)
	if err != nil {
		errorResponse := domain.NewErrorResponse("malformed_message", err.Error())
		return "", nil, &errorResponse
	}

	if len(fields) == 0 {
		errorResponse := domain.NewErrorResponse("malformed_message", "empty message")
		return "", nil, &errorResponse
	}

	messageType, ok := fields[0]["type"]
	if !ok || messageType == "" {
		errorResponse := domain.NewErrorResponse("malformed_message", "missing type in request")
		return "", nil, &errorResponse
	}

	return messageType, fields, nil
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

func (p *MessageProcessor) buildRequestByType(messageType string, fields []domain.BetAttempt) domain.Response {
	// This would probably fit better in another layer. A use case whose only job is
	// switching by message type does not sound especially appropriate, and adding a
	// dedicated handler layer right now would be mostly boilerplate for too little
	// behavior. If the amount of message types grows, this should likely move.
	switch messageType {
	case domain.MessageTypeBetBatch:
		protocolLog.Infof("action: route_message | result: success | type: %s", messageType)
		delete(fields[0], "type")
		return p.betRegistrar.Register(domain.BetBatchRequest{Bets: fields})
	case domain.MessageTypeFinish:
		agency, ok := fields[0]["agency"]
		if !ok || agency == "" {
			return domain.NewErrorResponse("malformed_message", "missing agency in request")
		}
		protocolLog.Infof("action: route_message | result: success | type: %s | agency: %s", messageType, agency)
		return p.agencyFinisher.Finish(domain.FinishAgencyRequest{Agency: agency})
	case domain.MessageTypeWinnersQuery:
		agency, ok := fields[0]["agency"]
		if !ok || agency == "" {
			return domain.NewErrorResponse("malformed_message", "missing agency in request")
		}
		protocolLog.Infof("action: route_message | result: success | type: %s | agency: %s", messageType, agency)
		return p.winnersQuerier.Query(domain.WinnersQueryRequest{Agency: agency})
	default:
		return domain.NewErrorResponse("malformed_message", fmt.Sprintf("unsupported message type %s", messageType))
	}
}
