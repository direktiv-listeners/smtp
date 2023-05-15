package main

import (
	"crypto/tls"
	"os"
	"strconv"

	smtp "github.com/emersion/go-smtp"
	"go.uber.org/zap"
)

const (
	ENV_USERNAME = "DIREKTIV_SMTP_USERNAME"
	ENV_PASSWORD = "DIREKTIV_SMTP_PASSWORD"
	ENV_ENDPOINT = "DIREKTIV_SMTP_ENDPOINT"
	ENV_TOKEN    = "DIREKTIV_SMTP_TOKEN"
	ENV_INSECURE = "DIREKTIV_SMTP_INSEURE_TLS"
	ENV_ADDRESS  = "DIREKTIV_SMTP_ADDRESS"
	ENV_HASH     = "DIREKTIV_SMTP_HASH"
)

func newServer() (*smtp.Server, error) {

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()

	user := os.Getenv(ENV_USERNAME)
	pwd := os.Getenv(ENV_PASSWORD)
	token := os.Getenv(ENV_TOKEN)
	insecure := os.Getenv(ENV_INSECURE)
	hash := os.Getenv(ENV_HASH)
	addr := os.Getenv(ENV_ADDRESS)

	endpoint := os.Getenv(ENV_ENDPOINT)
	if os.Getenv("K_SINK") != "" {
		endpoint = os.Getenv("K_SINK")
	}

	insecureBool, err := strconv.ParseBool(insecure)
	if err != nil {
		log.Warnf("can not parse value for tls insecure, setting to false")
		insecureBool = false
	}

	hashBool, err := strconv.ParseBool(hash)
	if err != nil {
		log.Warnf("can not parse value for hash, setting to false")
		hashBool = false
	}

	smtpConfig, err := newConfig(user, pwd, endpoint, token, insecureBool, hashBool, log)
	if err != nil {
		return nil, err
	}

	be := newBackend(log, smtpConfig)
	s := smtp.NewServer(be)

	s.Addr = ":2525"
	if addr != "" {
		s.Addr = addr
	}

	log.Infof("listening to %s", s.Addr)
	s.Domain = "localhost"
	s.AllowInsecureAuth = true
	s.TLSConfig = &tls.Config{}

	return s, nil

}
