package verification

import (
//	"strings"
	"bytes"
	"strconv"
	"fmt"
	"log"
	"os"
	"os/user"
	"net/mail"
	"io"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"camlistore.org/pkg/misc/pinentry"
	"github.com/zsusag/pangolin/account"
	myopenpgp "github.com/zsusag/pangolin/openpgp"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func SendVerificationEmail(conf account.Config, phrase string) {
	passphrase := getPassphrase(conf)

	// Set up authentication information
	auth := sasl.NewPlainClient("", conf.Email, passphrase)

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	
	// Get the Pangolin key
	kr := myopenpgp.ReadKeyring(usr.HomeDir+"/.pangolin/pangolin/pubring.gpg", "")

	pubkey := kr[0]
	// Get armored version of public key
	buf := new(bytes.Buffer)
	w, err := armor.Encode(buf, openpgp.PublicKeyType, nil)
	if err != nil {
		fmt.Println("I'm failing here like a dumbo")
		log.Fatal(err)
		os.Exit(1)
	}
	pubkey.Serialize(w)
	w.Close()
	pubkeyString := buf.String()
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email
	msgString := fmt.Sprintf("From: %s\r\n", conf.Email) +
		fmt.Sprintf("To: %s\r\n", conf.Email) +
		"Subject: New Pangolin Device Request\r\n" +
		fmt.Sprintf("X-PANG-UUID: %s\r\n", conf.UUID) +
		fmt.Sprintf("X-PANG-VERIFICATION: %s\r\n", phrase) +
		fmt.Sprintf("X-PANG-PUBKEY: %s\r\n", pubkeyString) +
		"\r\n" +
		"Please ignore this message as it is an automated message sent by Pangolin.\r\n"
	c, err := smtp.DialTLS(conf.SMTPServer+":"+strconv.Itoa(conf.SMTPPort), nil)
	if err != nil {
		log.Fatal("Could not dial")
	}

	defer c.Close()

	if err = c.Auth(auth); err != nil {
		log.Fatal("Authentication failed...")
	}

	if err = c.Mail(conf.Email); err != nil {
		log.Fatal("Could not make a mail message")
	}
	if err = c.Rcpt(conf.Email); err != nil {
		log.Fatal("Could not set rcpt")
	}

	w, err = c.Data()
	if err != nil {
		log.Fatal("Could not get data writer")
	}

	if _, err := io.WriteString(w, msgString); err != nil {
		log.Fatal("Could not write string")
	}

	err = w.Close()
	if err != nil {
		log.Fatal("Could not close writer")
	}

	c.Quit()
}

func WaitForVerification() {
	// Read configuration file
	conf := account.ReadConfig()

	// Connect to the IMAP server
	c, err := client.DialTLS(conf.IMAPServer+":"+strconv.Itoa(conf.IMAPPort), nil)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Logout()

	// Get passphrase for email account
	passphrase := getPassphrase(conf)
	
	// Login to the IMAP server
	if err := c.Login(conf.Email, passphrase); err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	inbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	var verificationMsg *imap.Message
	waiting := true
	for waiting {
		// Get the last 4 mesages
		from := uint32(1)
		to := inbox.Messages
		if inbox.Messages > 3 {
			from = inbox.Messages - 3
		}

		seqset := new(imap.SeqSet)
		seqset.AddRange(from, to)

		section := &imap.BodySectionName{}
		items := []imap.FetchItem{section.FetchItem()}

		messages := make(chan *imap.Message,10)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, items, messages)
		}()

		for msg := range messages {
			if msg.Envelope.Subject == "New Pangolin Device Request" {
				verificationMsg = msg
				waiting = false
			}
		}
	}

	r := verificationMsg.GetBody(&imap.BodySectionName{})
	if r == nil {
		log.Fatal("Server didn't return message body")
	}

	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Fatal(err)
	}

	header := m.Header
	phrase := header.Get("X-PANG-VERIFICATION")
	pubkey := header.Get("X-PANG-PUBKEY")
	uuid := header.Get("X-PANG-UUID")

	fmt.Println("Phrase: ", phrase)
	fmt.Println("Pubkey: ", pubkey)
	fmt.Println("UUID: ", uuid)
	
}

func getPassphrase(conf account.Config) string {
	// Request the email password to connect to the smtp server
	reqDesc := fmt.Sprintf("Please enter the passphrase for %s.", conf.Email)
	
	req := &pinentry.Request{
		Desc: reqDesc,
	}

	passphrase, err := req.GetPIN()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return passphrase
}
