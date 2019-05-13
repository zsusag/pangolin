package main

import (
	"log"
	"time"
	"flag"
	"fmt"
	"strconv"
	
	"github.com/emersion/go-imap/server"
	emersionsmtp "github.com/emersion/go-smtp"
	"github.com/zsusag/pangolin/imap"
	"github.com/zsusag/pangolin/smtp"
	"github.com/zsusag/pangolin/account"
//	"github.com/zsusag/pangolin/openpgp"
	"github.com/zsusag/pangolin/unlock"
	"github.com/emersion/go-pgp-pubkey/hkp"
	"github.com/zsusag/pangolin/verification"
)

func startIMAP(addr string) {
	be := imap.NewTLS(addr, nil, unlock.Unlock)

	// Create a new server
	s := server.New(be)
	s.Addr = ":1143"

	// Since we will use this server for testing only, we can allow plain text
	// authentication over unencrypted connections
	s.AllowInsecureAuth = true

	log.Println("Starting IMAP server at localhost:1143")
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func startSMTP(addr string) {
	source := hkp.New("http://ha.pool.sks-keyservers.net")
	kr := smtp.KeyRing{Source: source, Unlock: unlock.Unlock}
	be := smtp.NewTLS(addr, nil, kr)

	s := emersionsmtp.NewServer(be)
	s.Addr = ":1025"
	s.Domain = "localhost"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	log.Println("Starting SMTP server at " + s.Domain + s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	setupFlag := flag.Bool("setup", false, "Setup Pangolin for the first time")
	newDeviceFlag := flag.Bool("new-device", false, "Add a new device to your Pangolin network")
	
	flag.Parse()

	if(*setupFlag) {
		// TODO: Start making a new account and saving the information
		account.CreateAccount()
//		openpgp.ReadKeyring("/home/zsusag/.pangolin/email/secring.gpg", "Zach")
	} else if(*newDeviceFlag) {
		if account.CheckConfigStatus() {
			// Wait for a verification email from the new device
			verification.WaitForVerification()
		} else {
			// This is the new device, so send the verification email

			// Generate pangolin PGP key
			uuid := account.GenPangolinPGP(nil, "")
			conf,_,_ := account.AskInfo(false)
			conf.UUID = uuid

			account.WriteConfig(conf)
			pg := verification.NewPhraseGenerator()
			phrase := pg.GenPhrase()
			fmt.Println("Enter the following phrase on an existing device: ", phrase)
			verification.SendVerificationEmail(conf, phrase)
		} 
	} else {
		conf := account.ReadConfig()
		go startIMAP(conf.IMAPServer+":"+strconv.Itoa(conf.IMAPPort))
		startSMTP(conf.SMTPServer+":"+strconv.Itoa(conf.SMTPPort))
	}
}
