package smtp

import (
	"crypto/tls"
	"net"
	"log"
	
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/emersion/go-pgp-pubkey"
	"github.com/zsusag/pangolin/unlock"
)

type Security int

const (
	SecurityTLS Security = iota
	SecurityStartTLS
	SecurityNone
)

type KeyRing struct {
	Source pubkey.Source
	Unlock unlock.UnlockFunction
}

type Backend struct {
	Addr      string
	Security  Security
	TLSConfig *tls.Config
	Host      string

	KR KeyRing

	unexported struct{}
}

func New(addr string, kr KeyRing) *Backend {
	return &Backend{Addr: addr, Security: SecurityStartTLS, KR: kr}
}

func NewTLS(addr string, tlsConfig *tls.Config, kr KeyRing) *Backend {
	return &Backend{
		Addr: addr,
		Security: SecurityTLS,
		TLSConfig: tlsConfig,
		KR: kr,
	}
}

func (be *Backend) newConn() (*smtp.Client, error) {
	var conn net.Conn
	var err error
	if be.Security == SecurityTLS {
		conn, err = tls.Dial("tcp", be.Addr, be.TLSConfig)
	} else {
		conn, err = net.Dial("tcp", be.Addr)
	}

	if err != nil {
		return nil, err
	}

	var c *smtp.Client

	host := be.Host
	if host == "" {
		host, _, _ = net.SplitHostPort(be.Addr)
	}
	c, err = smtp.NewClient(conn, host)

	if err != nil {
		return nil, err
	}

	if be.Security == SecurityStartTLS {
		if err := c.StartTLS(be.TLSConfig); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (be *Backend) login(username, password string) (*smtp.Client, error) {
	c, err := be.newConn()
	if err != nil {
		return nil, err
	}

	auth := sasl.NewPlainClient("", username, password)
	if err := c.Auth(auth); err != nil {
		return nil, err
	}
	log.Println("Successful Authentication!")
	return c, nil
}

func (be *Backend) Login(_ *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	c, err := be.login(username, password)
	if err != nil {
		return nil, err
	} else if kr, err := be.KR.Unlock(username, password); err != nil {
		return nil, err
	} else {
		s := &session {
			c:  c,
			be: be,
			kr: kr,
		}

		return s, nil
	}
}

func (be *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return nil, smtp.ErrAuthRequired
}
