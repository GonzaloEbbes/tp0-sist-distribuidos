package common

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	code := m.Run()
	_ = os.Remove(StorageFilepath)
	os.Exit(code)
}

func tearDownStorageFile(t *testing.T) {
	t.Helper()
	_ = os.Remove(StorageFilepath)
}

func TestNewBetMustKeepFields(t *testing.T) {
	defer tearDownStorageFile(t)

	bet, err := NewBet("1", "first", "last", "10000000", "2000-12-20", "7500")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bet.Agency != 1 {
		t.Fatalf("expected agency 1, got %d", bet.Agency)
	}
	if bet.FirstName != "first" {
		t.Fatalf("expected first name first, got %s", bet.FirstName)
	}
	if bet.LastName != "last" {
		t.Fatalf("expected last name last, got %s", bet.LastName)
	}
	if bet.Document != "10000000" {
		t.Fatalf("expected document 10000000, got %s", bet.Document)
	}
	if !bet.Birthdate.Equal(time.Date(2000, 12, 20, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected birthdate: %v", bet.Birthdate)
	}
	if bet.Number != 7500 {
		t.Fatalf("expected number 7500, got %d", bet.Number)
	}
}

func TestHasWonWithWinnerNumberMustBeTrue(t *testing.T) {
	defer tearDownStorageFile(t)

	bet, err := NewBet("1", "first", "last", "10000000", "2000-12-20", "7574")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !HasWon(bet) {
		t.Fatal("expected winner bet to return true")
	}
}

func TestHasWonWithNonWinnerNumberMustBeFalse(t *testing.T) {
	defer tearDownStorageFile(t)

	bet, err := NewBet("1", "first", "last", "10000000", "2000-12-20", "7575")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if HasWon(bet) {
		t.Fatal("expected non winner bet to return false")
	}
}

func TestStoreBetsAndLoadBetsKeepsFieldsData(t *testing.T) {
	defer tearDownStorageFile(t)

	bet, err := NewBet("1", "first", "last", "10000000", "2000-12-20", "7500")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := StoreBets([]Bet{bet}); err != nil {
		t.Fatalf("unexpected error storing bets: %v", err)
	}

	loaded, err := LoadBets()
	if err != nil {
		t.Fatalf("unexpected error loading bets: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("expected 1 bet, got %d", len(loaded))
	}

	assertEqualBets(t, bet, loaded[0])
}

func TestStoreBetsAndLoadBetsKeepsRegistryOrder(t *testing.T) {
	defer tearDownStorageFile(t)

	first, err := NewBet("0", "first_0", "last_0", "10000000", "2000-12-20", "7500")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := NewBet("1", "first_1", "last_1", "10000001", "2000-12-21", "7501")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := StoreBets([]Bet{first, second}); err != nil {
		t.Fatalf("unexpected error storing bets: %v", err)
	}

	loaded, err := LoadBets()
	if err != nil {
		t.Fatalf("unexpected error loading bets: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 bets, got %d", len(loaded))
	}

	assertEqualBets(t, first, loaded[0])
	assertEqualBets(t, second, loaded[1])
}

func assertEqualBets(t *testing.T, expected Bet, actual Bet) {
	t.Helper()

	if expected.Agency != actual.Agency {
		t.Fatalf("expected agency %d, got %d", expected.Agency, actual.Agency)
	}
	if expected.FirstName != actual.FirstName {
		t.Fatalf("expected first name %s, got %s", expected.FirstName, actual.FirstName)
	}
	if expected.LastName != actual.LastName {
		t.Fatalf("expected last name %s, got %s", expected.LastName, actual.LastName)
	}
	if expected.Document != actual.Document {
		t.Fatalf("expected document %s, got %s", expected.Document, actual.Document)
	}
	if !expected.Birthdate.Equal(actual.Birthdate) {
		t.Fatalf("expected birthdate %v, got %v", expected.Birthdate, actual.Birthdate)
	}
	if expected.Number != actual.Number {
		t.Fatalf("expected number %d, got %d", expected.Number, actual.Number)
	}
}
