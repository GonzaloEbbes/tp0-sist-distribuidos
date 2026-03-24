package domain

const MessageTypeBetBatch = "bet_batch"
const MessageTypeFinish = "finish"
const MessageTypeWinnersQuery = "winners_query"

type BetRequest struct {
	Type   string
	Agency string
	Bets   []Bet
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

func (r ServerResponse) IsDrawNotReady() bool {
	return r.Code == "draw_not_ready"
}
