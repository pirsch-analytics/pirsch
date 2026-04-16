package session

import "github.com/pirsch-analytics/pirsch/v7/pkg/ingest"

type Session struct{}

func NewSession() *Session {
	return &Session{}
}

func (session *Session) Step(request *ingest.Request) (bool, error) {
	return false, nil
}
