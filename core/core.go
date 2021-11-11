package core

import (
	"io"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"

	"github.com/oschwald/geoip2-golang"
)

func NewChannelBackend(incomingMessage chan MessageMeta) smtp.Backend {
	return &ChannelBackend{
		channel: incomingMessage,
	}
}

type GeoPoint struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}
type MessageMeta struct {
	To         string      `json:"to"`
	From       string      `json:"from"`
	FromDomain string      `json:"from_domain"`
	FromIP     net.IP      `json:"from_ip"`
	ToIP       net.IP      `json:"to_ip"`
	ToDomain   string      `json:"to_domain"`
	Size       int         `json:"size"`
	Location   GeoPoint    `json:"loc"`
	Geo        geoip2.City `json:"geo"`
	Milis      int64       `json:"ts"`
	Data       []byte      `json:"-"`
}

// The Backend implements SMTP server methods.
type ChannelBackend struct {
	channel chan MessageMeta
}

func (bkd *ChannelBackend) NewSession(info smtp.ConnectionState, _ string) (smtp.Session, error) {
	msg := &MessageMeta{
		FromIP: net.ParseIP(strings.Split(info.RemoteAddr.String(), ":")[0]),
		ToIP:   net.ParseIP(strings.Split(info.LocalAddr.String(), ":")[0]),
		Milis:  time.Now().UTC().UnixNano() / int64(time.Millisecond),
	}
	return &Session{
		channel: bkd.channel,
		msg:     msg,
	}, nil
}

func (bkd *ChannelBackend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	log.Debug(state.Hostname, username, password)

	return bkd.NewSession(*state, "")
}
func (bkd *ChannelBackend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	log.Debug(state.Hostname)
	return bkd.NewSession(*state, "")
}

// A Session is returned after EHLO.
type Session struct {
	channel chan MessageMeta
	msg     *MessageMeta
}

func (s *Session) AuthPlain(username, password string) error {
	log.Debug(username, password)
	return nil
}

func (s *Session) Mail(from string, opts smtp.MailOptions) error {
	s.msg.From = from
	tmp := strings.Split(from, "@")
	if len(tmp) > 1 {
		s.msg.FromDomain = tmp[1]
	}
	return nil
}

func (s *Session) Rcpt(to string) error {
	s.msg.To = to
	tmp := strings.Split(to, "@")
	if len(tmp) > 1 {
		s.msg.ToDomain = tmp[1]
	}
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if b, err := ioutil.ReadAll(r); err != nil {
		return err
	} else {
		s.msg.Data = b
		s.msg.Size = len(b)
	}

	s.SendData()
	return nil
}

func (s *Session) Reset() {
	s.SendData()
}

func (s *Session) Logout() error {
	s.SendData()
	return nil
}

func (s *Session) SendData() {
	if s.msg != nil && s.msg.From != "" {
		s.channel <- *s.msg
		// reset the msg
		// while preserving the connection meta
		s.msg.From = ""
		s.msg.FromDomain = ""
		s.msg.To = ""
		s.msg.ToDomain = ""
		s.msg.Size = 0
	}
}
