package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/emersion/go-message/mail"
	smtp "github.com/emersion/go-smtp"
	"github.com/mitchellh/hashstructure/v2"
	"go.uber.org/zap"
)

type backend struct {
	log        *zap.SugaredLogger
	smtpConfig *config
}

type session struct {
	to         []string
	data       map[string]interface{}
	log        *zap.SugaredLogger
	smtpConfig *config
}

type Attachment struct {
	Data        []byte `json:"data"`
	ContentType string `json:"type"`
	Name        string `json:"name"`
}

func newBackend(log *zap.SugaredLogger, config *config) *backend {

	return &backend{
		log:        log,
		smtpConfig: config,
	}
}

func (bkd *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &session{
		log:        bkd.log,
		smtpConfig: bkd.smtpConfig,
		data:       make(map[string]interface{}),
	}, nil
}

func (s *session) AuthPlain(username, password string) error {

	if s.smtpConfig.user == "" {
		return nil
	}

	if username != s.smtpConfig.user || password != s.smtpConfig.pwd {
		return fmt.Errorf("username or password invalid")
	}

	return nil
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	s.data["from"] = from
	return nil
}

func (s *session) Rcpt(to string) error {
	s.to = append(s.to, to)
	return nil
}

func (s *session) Data(r io.Reader) error {

	mr, err := mail.CreateReader(r)

	if err != nil {
		s.log.Errorf("can not create mail reader: %s", err.Error())
		return err
	}

	subj, err := mr.Header.Subject()
	if err != nil {
		s.log.Errorf("can not read subject: %s", err.Error())
		return err
	}
	s.data["subject"] = subj

	attachments, message, err := handleAttachments(mr)
	if err != nil {
		s.log.Errorf("can not read content and attachments: %s", err.Error())
		return err
	}

	s.data["to"] = s.to
	s.data["attachments"] = attachments
	s.data["message"] = message

	event := basicEvent()
	event.SetData(s.data)

	event.SetID(fmt.Sprintf("id%d", rand.Int()))
	if s.smtpConfig.hash {
		hash, err := hashstructure.Hash(s.data, hashstructure.FormatV2, nil)
		if err != nil {
			s.log.Errorf("can not hash data: %s", err.Error())
			return err
		}

		event.SetID(fmt.Sprintf("%d", hash))
	}

	s.log.Infof("sending cloud event to %s", s.smtpConfig.endpoint.String())
	err = sendCloudEvent(event, s.smtpConfig.endpoint.String(),
		s.smtpConfig.token, s.smtpConfig.insecureTLS)
	if err != nil {
		s.log.Errorf("can not send cloud event: %s", err.Error())
	}

	return err
}

func (s *session) Reset() {
	s.data = make(map[string]interface{})
}

func (s *session) Logout() error {
	return nil
}

func basicEvent() cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetSource("direktiv/listener/smtp")
	event.SetType("smtp.message")
	return event
}

func handleAttachments(mr *mail.Reader) ([]*Attachment, string, error) {

	attachments := make([]*Attachment, 0)
	var message string

	for {
		p, err := mr.NextPart()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, "", err
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			b, _ := ioutil.ReadAll(p.Body)
			if len(string(b)) > 0 {
				message = string(b)
			}
		case *mail.AttachmentHeader:

			ct, _, err := h.ContentType()
			if err != nil {
				return nil, "", err
			}

			filename, err := h.Filename()
			if err != nil {
				return nil, "", err
			}

			b, err := ioutil.ReadAll(p.Body)
			if err != nil {
				return nil, "", err
			}

			a := &Attachment{
				Data:        b,
				ContentType: ct,
				Name:        filename,
			}

			attachments = append(attachments, a)
		}
	}

	return attachments, message, nil

}

func sendCloudEvent(event cloudevents.Event, endpoint, token string, insecure bool) error {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	options := []cehttp.Option{
		cloudevents.WithTarget(endpoint),
		cloudevents.WithStructuredEncoding(),
		cloudevents.WithHTTPTransport(tr),
	}

	if len(token) > 0 {
		log.Printf("using token to login")
		options = append(options,
			cehttp.WithHeader("Direktiv-Token", token))
	}

	t, err := cloudevents.NewHTTPTransport(
		options...,
	)
	if err != nil {
		return err
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		log.Printf("unable to create cloudevent client: " + err.Error())
		return err
	}

	_, _, err = c.Send(context.Background(), event)
	if err != nil {
		log.Printf("unable to send cloudevent client: " + err.Error())
		return err
	}

	return nil

}

// func (bkd *backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
// 	return &session{
// 		data: make(map[string]interface{}),
// 		// event: basicEvent(),
// 	}, nil
// }

// func (bkd *backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
// 	return &session{
// 		data: make(map[string]interface{}),
// 		// event: basicEvent(),
// 	}, nil
// }
