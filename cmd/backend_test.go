package main

import (
	"testing"

	"github.com/emersion/go-smtp"
	"go.uber.org/zap"
)

func TestAuthPlain(t *testing.T) {

	logger, _ := zap.NewProduction()
	config, err := newConfig("username", "password", "http://myserver.com", "",
		true, true, logger.Sugar())

	be := newBackend(logger.Sugar(), config)

	conn := &smtp.Conn{}

	session, err := be.NewSession(conn)
	if err != nil {
		t.Fail()
	}

	err = session.AuthPlain("invalid", "invalid")
	if err == nil {
		t.Logf("username and password auth should fail")
		t.Fail()
	}

	err = session.AuthPlain("username", "password")
	if err != nil {
		t.Logf("username and password auth should not fail")
		t.Fail()
	}

}
