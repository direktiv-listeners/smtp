package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	mail "github.com/wneessen/go-mail"
)

var receiver testServer

func init() {

	receiver = testServer{}
	receiver.prepareReceiver()

	go receiver.startReceiver()

}

func TestServer(t *testing.T) {

	addr := "localhost:8888"
	os.Setenv(ENV_ADDRESS, addr)
	os.Setenv(ENV_ENDPOINT, "http://testme.com")

	s, _ := newServer()

	if s.Addr != addr {
		t.Logf("address not set correctly")
		t.Fail()
	}

}

func TestSendServerAuth(t *testing.T) {

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	addr := l.Addr().String()
	os.Setenv(ENV_ADDRESS, addr)

	os.Setenv(ENV_USERNAME, "user")
	os.Setenv(ENV_PASSWORD, "pwd")

	os.Setenv(ENV_ENDPOINT, fmt.Sprintf("http://%s", receiver.addr))
	s, _ := newServer()
	go s.Serve(l)

	port := l.Addr().(*net.TCPAddr).Port
	ip := l.Addr().(*net.TCPAddr).IP

	// successful login and send
	client, err := mail.NewClient(ip.String(), mail.WithPort(port), mail.WithDebugLog(),
		mail.WithTLSPolicy(mail.NoTLS), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername("user"), mail.WithPassword("pwd"))
	if err != nil {
		t.Fatal(err)
	}

	err = mailSend(client)
	if err != nil {
		t.Fatal(err)
	}

	// no authentication, should fail
	client, err = mail.NewClient(ip.String(), mail.WithPort(port), mail.WithDebugLog(),
		mail.WithTLSPolicy(mail.NoTLS))
	if err != nil {
		t.Fatal(err)
	}

	err = mailSend(client)
	if err == nil {
		t.Fatal(err)
	}

	client, err = mail.NewClient(ip.String(), mail.WithPort(port), mail.WithDebugLog(),
		mail.WithTLSPolicy(mail.NoTLS), mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername("wrong"), mail.WithPassword("credentials"))
	if err != nil {
		t.Fatal(err)
	}

	err = mailSend(client)
	if err == nil {
		t.Fatal(err)
	}

}

func TestSendServerToken(t *testing.T) {

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv(ENV_USERNAME, "")
	os.Setenv(ENV_PASSWORD, "")
	os.Setenv(ENV_TOKEN, "123")

	addr := l.Addr().String()
	os.Setenv(ENV_ADDRESS, addr)

	os.Setenv(ENV_ENDPOINT, fmt.Sprintf("http://%s", receiver.addr))
	s, _ := newServer()
	go s.Serve(l)

	port := l.Addr().(*net.TCPAddr).Port
	ip := l.Addr().(*net.TCPAddr).IP

	client, err := mail.NewClient(ip.String(), mail.WithPort(port), mail.WithDebugLog(),
		mail.WithTLSPolicy(mail.NoTLS))
	if err != nil {
		t.Fatal(err)
	}

	err = mailSend(client)
	if err != nil {
		t.Fatal(err)
	}

	if receiver.lastHeaders["Direktiv-Token"] != "123" {
		t.Fatal("direrktiv-token not set")
	}

}

func TestSendServer(t *testing.T) {

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv(ENV_USERNAME, "")
	os.Setenv(ENV_PASSWORD, "")

	addr := l.Addr().String()
	os.Setenv(ENV_ADDRESS, addr)

	os.Setenv(ENV_ENDPOINT, fmt.Sprintf("http://%s", receiver.addr))
	s, _ := newServer()
	go s.Serve(l)

	port := l.Addr().(*net.TCPAddr).Port
	ip := l.Addr().(*net.TCPAddr).IP

	client, err := mail.NewClient(ip.String(), mail.WithPort(port), mail.WithDebugLog(),
		mail.WithTLSPolicy(mail.NoTLS))
	if err != nil {
		t.Fatal(err)
	}

	err = mailSend(client)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	data := receiver.lastRequest["data"].(map[string]interface{})

	if data["from"] != "info@direktiv.io" {
		t.Log("from not set correctly")
		t.Fail()
	}

	to := data["to"].([]interface{})
	if to[0] != "info@direktiv.io" {
		t.Log("to not set correctly")
		t.Fail()
	}

	if data["subject"] != "Hello World" {
		t.Log("subject not set correctly")
		t.Fail()
	}

	if data["message"] != "This is a text" {
		t.Log("message not set correctly")
		t.Fail()
	}

	attachments := data["attachments"].([]interface{})

	if len(attachments) != 2 {
		t.Log("attachment number is not 2")
		t.Fail()
	}

	a := attachments[0].(map[string]interface{})
	if a["name"] != "test1" {
		t.Log("first attachment wrong")
		t.Fail()
	}

}

func TestSendServerE2ETLS(t *testing.T) {

	addr := os.Getenv("TEST_SERVER")
	port := os.Getenv("TEST_PORT")

	if addr == "" || port == "" {
		t.Skip("skipping knative test")
	}

	t.Log("running kubernetes tls test")

	ps, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	client, err := mail.NewClient(addr, mail.WithPort(ps),
		mail.WithTLSPolicy(mail.TLSOpportunistic), mail.WithTLSConfig(tlsConfig), mail.WithSSL(),
		mail.WithDebugLog())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err = mailSend(client)
	if err != nil {
		t.Fatal(err)
	}

}

func TestSendServerE2E(t *testing.T) {

	addr := os.Getenv("TEST_SERVER")
	port := os.Getenv("TEST_PORT")

	if addr == "" || port == "" {
		t.Skip("skipping knative test")
	}

	t.Log("running kubernetes test")

	ps, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}

	client, err := mail.NewClient(addr, mail.WithPort(ps),
		mail.WithTLSPolicy(mail.NoTLS), mail.WithDebugLog())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err = mailSend(client)
	if err != nil {
		t.Fatal(err)
	}

}

type testServer struct {
	addr        string
	hasError    bool
	lastRequest map[string]interface{}
	lastHeaders map[string]string
}

func (s *testServer) startReceiver() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		s.lastHeaders = make(map[string]string)

		for name, values := range r.Header {
			for _, value := range values {
				s.lastHeaders[name] = value
			}
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			s.hasError = true
			return
		}

		var resp map[string]interface{}

		err = json.Unmarshal(b, &resp)
		if err != nil {
			s.hasError = true
			return
		}

		s.lastRequest = resp

	})

	log.Fatal(http.ListenAndServe(s.addr, nil))

}

func (s *testServer) prepareReceiver() {

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("can not get listener")
	}
	defer l.Close()

	s.addr = l.Addr().String()

}

func mailSend(client *mail.Client) error {

	m := mail.NewMsg()
	m.From("info@direktiv.io")
	m.To("info@direktiv.io")
	m.Subject("Hello World")
	m.SetBodyString(mail.TypeTextPlain, "This is a text")

	r1 := strings.NewReader("this is an attachment")
	m.AttachReader("test1", r1)

	r2 := strings.NewReader("this is an attachment")
	m.AttachReader("test2", r2)

	return client.DialAndSend(m)

}
