package main

import (
	"testing"

	"go.uber.org/zap"
)

func TestConfigURL(t *testing.T) {

	logger, _ := zap.NewProduction()

	_, err := newConfig("", "", "invalid", "", true, true, logger.Sugar())
	if err == nil {
		t.Logf("URL should be invalid")
		t.Fail()
	}

	_, err = newConfig("", "", "http://myserver.com", "", true, true, logger.Sugar())
	if err != nil {
		t.Logf("URL should be valid")
		t.Fail()
	}

}

func TestConfigUserPwd(t *testing.T) {

	logger, _ := zap.NewProduction()

	_, err := newConfig("user", "", "http://myserver.com", "", true, true, logger.Sugar())
	if err == nil {
		t.Logf("user password combination should be invalid")
		t.Fail()
	}

	_, err = newConfig("", "pwd", "http://myserver.com", "", true, true, logger.Sugar())
	if err == nil {
		t.Logf("user password combination should be invalid")
		t.Fail()
	}

	_, err = newConfig("user", "pwd", "http://myserver.com", "", true, true, logger.Sugar())
	if err != nil {
		t.Logf("user password combination should be valid")
		t.Fail()
	}

}
