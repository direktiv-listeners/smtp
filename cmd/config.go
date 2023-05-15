package main

import (
	"fmt"
	"net/url"

	"go.uber.org/zap"
)

type config struct {
	user, pwd string

	endpoint          *url.URL
	token             string
	insecureTLS, hash bool

	log *zap.SugaredLogger
}

func newConfig(user, pwd, endpoint, token string, insecure, hash bool,
	log *zap.SugaredLogger) (*config, error) {

	var err error
	c := &config{
		user:        user,
		pwd:         pwd,
		log:         log,
		token:       token,
		hash:        hash,
		insecureTLS: insecure,
	}

	c.log.Infof("parsing endpoint %s", endpoint)
	c.endpoint, err = url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, err
	}

	c.log.Infof("checking username and password")
	if (user != "" && pwd == "") ||
		(user == "" && pwd != "") {
		return nil, fmt.Errorf("username and password both needed")
	}

	return c, nil

}
