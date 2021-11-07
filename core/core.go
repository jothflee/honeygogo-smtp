package core

import (
	"io"
	"io/ioutil"

	"github.com/emersion/go-smtp"
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
	To       string      `json:"to"`
	From     string      `json:"from"`
	FromAddr string      `json:"from_addr"`
	ToAddr   string      `json:"to_addr"`
	Size     int         `json:"size"`
	Location GeoPoint    `json:"loc"`
	Geo      geoip2.City `json:"geo"`
	Milis    int64       `json:"ts"`
}

// The Backend implements SMTP server methods.
type ChannelBackend struct {
	channel chan MessageMeta
}

func (bkd *ChannelBackend) NewSession(info smtp.ConnectionState, _ string) (smtp.Session, error) {
	msg := &MessageMeta{
		FromAddr: info.RemoteAddr.String(),
		ToAddr:   info.LocalAddr.String(),
	}
	return &Session{
		channel: bkd.channel,
		msg:     msg,
	}, nil
}

func (bkd *ChannelBackend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return bkd.NewSession(*state, "")
}
func (bkd *ChannelBackend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return bkd.NewSession(*state, "")
}

// A Session is returned after EHLO.
type Session struct {
	channel chan MessageMeta
	msg     *MessageMeta
}

func (s *Session) AuthPlain(username, password string) error {
	return nil
}

func (s *Session) Mail(from string, opts smtp.MailOptions) error {
	s.msg.From = from
	return nil
}

func (s *Session) Rcpt(to string) error {
	s.msg.To = to
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if b, err := ioutil.ReadAll(r); err != nil {
		return err
	} else {
		s.msg.Size = len(b)
	}
	s.channel <- *s.msg
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}
