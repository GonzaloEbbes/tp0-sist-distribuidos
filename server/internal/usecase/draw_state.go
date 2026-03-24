package usecase

import (
	"sync"

	"github.com/op/go-logging"
)

var drawStateLog = logging.MustGetLogger("log")

type DrawState struct {
	expectedAgencies  int
	finishedAgencies  map[int]struct{}
	drawAlreadyLogged bool
	mu                sync.Mutex
}

func NewDrawState(expectedAgencies int) *DrawState {
	return &DrawState{
		expectedAgencies: expectedAgencies,
		finishedAgencies: make(map[int]struct{}, expectedAgencies),
	}
}

func (s *DrawState) FinishAgency(agency int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.finishedAgencies[agency] = struct{}{}
	if !s.drawAlreadyLogged && len(s.finishedAgencies) >= s.expectedAgencies {
		s.drawAlreadyLogged = true
		drawStateLog.Infof("action: sorteo | result: success")
	}
}

func (s *DrawState) IsReady() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.finishedAgencies) >= s.expectedAgencies
}
