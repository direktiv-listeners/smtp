package main

import (
	"crypto/tls"
	"os"
	"strconv"
	"time"

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
	ENV_DEBUG    = "DIREKTIV_SMTP_DEBUG"
)

func newServer() (*smtp.Server, error) {

	user := os.Getenv(ENV_USERNAME)
	pwd := os.Getenv(ENV_PASSWORD)
	token := os.Getenv(ENV_TOKEN)
	insecure := os.Getenv(ENV_INSECURE)
	hash := os.Getenv(ENV_HASH)
	addr := os.Getenv(ENV_ADDRESS)
	debug := os.Getenv(ENV_DEBUG)

	endpoint := os.Getenv(ENV_ENDPOINT)
	if os.Getenv("K_SINK") != "" {
		endpoint = os.Getenv("K_SINK")
	}

	var logger *zap.Logger
	if debug != "" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()
	log := logger.Sugar()

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

	s.WriteTimeout = 10 * time.Second
	s.ReadTimeout = 10 * time.Second
	s.AllowInsecureAuth = true
	s.MaxRecipients = 50

	if debug != "" {
		s.Debug = os.Stdout
	}

	s.AuthDisabled = true
	if smtpConfig.user != "" {
		s.AuthDisabled = false
	}

	if _, err := os.Stat("/smtp-certs"); !os.IsNotExist(err) {
		log.Infof("certifcates found")

		cert, err := tls.LoadX509KeyPair("/smtp-certs/tls.crt", "/smtp-certs/tls.key")
		if err != nil {
			return nil, err
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		s.TLSConfig = tlsConfig
	}

	s.MaxMessageBytes = 100 * 1024 * 1024

	return s, nil

}
