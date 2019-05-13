package smtp

import (
	"io"
	"errors"
	"log"
	"bytes"
	netmail "net/mail"

	
	"github.com/emersion/go-smtp"
	"golang.org/x/crypto/openpgp"
	myopenpgp "github.com/zsusag/pangolin/openpgp"
)

type session struct {
	c *smtp.Client
	be *Backend

	kr openpgp.EntityList
}

func (s *session) Mail(from string) error {
	return nil
}

func (s *session) Rcpt(to string) error {
	return nil
}

func (s *session) Data(r io.Reader) error {
	var pubkeys openpgp.EntityList
	var plaintextTo []string
	var encryptedTo []string
	
	var buf bytes.Buffer

	if _, err := io.Copy(&buf, r); err != nil {
		return err
	}
	r = bytes.NewReader(buf.Bytes())
	r1 := bytes.NewReader(buf.Bytes())

	msg, err := netmail.ReadMessage(r1)
	if err != nil {
		log.Fatal(err)
	}

	header := msg.Header
	fromList, _ := header.AddressList("From")
	toList, _ := header.AddressList("To")
	ccList, _ := header.AddressList("Cc")
	bccList, _ := header.AddressList("Bcc")

	// Make one big list of recipients
	var rcptList []*netmail.Address
	rcptList = append(rcptList, toList...)
	rcptList = append(rcptList, ccList...)
	rcptList = append(rcptList, bccList...)
	
	if len(fromList) != 1 {
		return errors.New("the From field must contain exactly one address")
	}

	if len(rcptList) == 0 {
		return errors.New("no recipient specified")
	}

	rawFrom := fromList[0]
	fromAddrStr := rawFrom.Address

	for _, addr := range rcptList {
		keys, err := s.be.KR.Source.Search("<"+addr.Address+">")
		if err != nil {
			return err
		}

		if len(keys) == 0 {
			plaintextTo = append(plaintextTo, addr.Address)
		} else {
			encryptedTo = append(encryptedTo, addr.Address)
			pubkeys = append(pubkeys, keys[0])
		}
	}

	// Keep a copy of the plaintext message to be able to send it to plaintext
	// recipients
	plaintext := r
	if len(encryptedTo) > 0 && len(plaintextTo) > 0 {
		b := &bytes.Buffer{}
		r = io.TeeReader(r,b)
		plaintext = b
	}

	// Start encrypted message
	var ret error 
	if len(encryptedTo) > 0 {
		// Start an email message
		if err := s.c.Mail(fromAddrStr); err != nil {
			return err
		}

		// Add all of the recipients
		for _, rcpt := range encryptedTo {
			if err := s.c.Rcpt(rcpt); err != nil {
				return err
			}
		}

		// Encrypt the body of the email
		b := new(bytes.Buffer)
		if err := myopenpgp.Encrypt(b, r, pubkeys, s.kr[0]); err != nil {
			return err
		}

		// Start the data of the mail message
		wc, err := s.c.Data()
		if err != nil {
			return err
		}

		// Write the header
		// for k,v := range header {
		// 	hdrString := fmt.Sprintf("%s: %s\n", k, v[0])
		// 	if _, err := io.WriteString(wc, hdrString); err != nil {
		// 		wc.Close()
		// 		return err
		// 	}
		// }

		// Write the body
		if _, err := io.WriteString(wc, b.String()); err != nil {
			wc.Close()
			return err
		}

		ret = wc.Close()
	}

	// Send the message to plaintext recipients
	if len(plaintextTo) > 0 {
		// Start an email message
		if err := s.c.Mail(fromAddrStr); err != nil {
			return err
		}

		// Add all of the recipients
		for _, rcpt := range plaintextTo {
			if err := s.c.Rcpt(rcpt); err != nil {
				return err
			}
		}

		// Start the data of the mail message
		wc, err := s.c.Data()
		if err != nil {
			return err
		}

		// Write the header
		// for k,v := range header {
		// 	hdrString := fmt.Sprintf("%s: %s\n", k, v[0])
		// 	if _, err := io.WriteString(wc, hdrString); err != nil {
		// 		wc.Close()
		// 		return err
		// 	}
		// }

		// Write the body
		if _, err := io.Copy(wc, plaintext); err != nil {
			wc.Close()
			return err
		}

		if ret == nil {
			ret = wc.Close()
		}
	}
	return ret
}

func (s *session) Reset() {}

func (s *session) Logout() error {
	return s.c.Close()
}
