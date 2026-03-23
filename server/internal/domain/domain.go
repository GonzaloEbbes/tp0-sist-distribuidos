package domain

import "time"

type BetAttempt map[string]string

type BetRequest struct {
	Bets []BetAttempt
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
