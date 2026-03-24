package domain

import "time"

type BetAttempt map[string]string

const MessageTypeBetBatch = "bet_batch"
const MessageTypeFinish = "finish"
const MessageTypeWinnersQuery = "winners_query"

type BetBatchRequest struct {
	Bets []BetAttempt
}

type FinishAgencyRequest struct {
	Agency string
}

type WinnersQueryRequest struct {
	Agency string
}

type Bet struct {
	Agency    int
	FirstName string
	LastName  string
	Document  string
	Birthdate time.Time
	Number    int
}

type Response struct {
	Status  string
	Code    string
	Message string
}

func NewSuccessResponse() Response {
	return Response{Status: "ok"}
}

func NewErrorResponse(code string, message string) Response {
	return Response{
		Status:  "error",
		Code:    code,
		Message: message,
	}
}
