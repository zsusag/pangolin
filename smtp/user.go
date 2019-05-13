package smtp

import (
	"io"

	"github.com/emersion/go-smtp"

	"golang.org/x/crypto/openpgp"
)

type user struct {
	c *smtp.Client
	be *Backend

	kr openpgp.EntityList
}

func (u *user) Send(from string, to []string, r io.Reader) error {
	if err := u.c.Mail(from); err != nil {
		return err
	}

	for _, rcpt := range to {
		if err := u.c.Rcpt(rcpt); err != nil {
			return err
		}
	}

	wc, err := u.c.Data()
	if err != nil {
		return err
	}

	_, err = io.Copy(wc, r)
	if err != nil {
		wc.Close()
		return err
	}

	return wc.Close()
}

func (u *user) Logout() error {
	return u.c.Close()
}
