package common

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"
)

const StorageFilepath = "./bets.csv"
const LotteryWinnerNumber = 7574

type Bet struct {
	Agency    int
	FirstName string
	LastName  string
	Document  string
	Birthdate time.Time
	Number    int
}

func NewBet(agency string, firstName string, lastName string, document string, birthdate string, number string) (Bet, error) {
	agencyInt, err := strconv.Atoi(agency)
	if err != nil {
		return Bet{}, err
	}

	birthdateDate, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		return Bet{}, err
	}

	numberInt, err := strconv.Atoi(number)
	if err != nil {
		return Bet{}, err
	}

	return Bet{
		Agency:    agencyInt,
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdateDate,
		Number:    numberInt,
	}, nil
}

func HasWon(bet Bet) bool {
	return bet.Number == LotteryWinnerNumber
}

func StoreBets(bets []Bet) error {
	file, err := os.OpenFile(StorageFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	for _, bet := range bets {
		if err := writer.Write([]string{
			strconv.Itoa(bet.Agency),
			bet.FirstName,
			bet.LastName,
			bet.Document,
			bet.Birthdate.Format("2006-01-02"),
			strconv.Itoa(bet.Number),
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func LoadBets() ([]Bet, error) {
	file, err := os.Open(StorageFilepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	bets := make([]Bet, 0, len(rows))
	for _, row := range rows {
		bet, err := NewBet(row[0], row[1], row[2], row[3], row[4], row[5])
		if err != nil {
			return nil, err
		}
		bets = append(bets, bet)
	}

	return bets, nil
}
