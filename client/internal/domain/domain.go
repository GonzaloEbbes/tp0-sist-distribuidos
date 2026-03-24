package domain

const MessageTypeBetBatch = "bet_batch"

type BetRequest struct {
	Type string
	Bets []Bet
}

type Bet struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    string
}

type ServerResponse struct {
	Status  string
	Code    string
	Message string
}

func (r ServerResponse) IsSuccess() bool {
	return r.Status == "ok"
}
