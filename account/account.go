package account

import (
	"time"
	"fmt"
	"bufio"
	"strings"
	"os"
	"os/user"
	"log"
	"io/ioutil"
	"strconv"
	"encoding/json"
	
	"github.com/badoux/checkmail"
	"github.com/google/uuid"
	"github.com/zsusag/pangolin/openpgp"
)

type Config struct {
	IMAPServer string
	IMAPPort int
	SMTPServer string
	SMTPPort int
	Email string
	UUID string
}

func ReadStringTrimmed(r *bufio.Reader) string {
	s,_ := r.ReadString('\n')
	return strings.TrimSpace(s)
}

func GenPangolinPGP(conf *openpgp.Config, homedir string) string {
	if conf == nil {
		conf = &openpgp.Config{Expiry: 365 * 24 * time.Hour}
	}

	if homedir == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		homedir = usr.HomeDir
	}
	
	// Generate pangolin keypair
	uuidString := uuid.New().String()
	pangolinKeypair, err := openpgp.CreateKey(uuidString, "", uuidString+"@example.com", conf)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	err = os.Mkdir(homedir + "/.pangolin/pangolin", 0700)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	pangPubring := pangolinKeypair.Pubring()
	pangSecring := pangolinKeypair.Secring()

	err = ioutil.WriteFile(homedir +"/.pangolin/pangolin/pubring.gpg", pangPubring, 0600)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(homedir +"/.pangolin/pangolin/secring.gpg", pangSecring, 0600)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return uuidString
}

func AskInfo(askName bool) (Config,string,string) {
	// Prompt user for configuration details
	var name, email, comment, imapserver, smtpserver string
	var imapport, smtpport int
	for {
		reader := bufio.NewReader(os.Stdin)
		if askName {
			fmt.Print("Enter your name: ")
			name = ReadStringTrimmed(reader)
		}
		for {
			fmt.Print("Enter your email: ")
			email = ReadStringTrimmed(reader)
			err := checkmail.ValidateFormat(email)
			if err != nil {
				fmt.Println("Invalid format for email. Try again.")
			} else {
				break
			}
		}
		fmt.Printf("Enter the IMAP Server for %s: ", email)
		imapserver = ReadStringTrimmed(reader)
		for {
			fmt.Printf("Enter the IMAP Server port for %s: ", email)
			tmp, err := strconv.Atoi(ReadStringTrimmed(reader))
			if err != nil {
				fmt.Println("Error: Not an integer. Please try again.")
			} else {
				imapport = tmp
				break
			}
		}
		fmt.Printf("Enter the SMTP server for %s: ", email)
		smtpserver = ReadStringTrimmed(reader)
		for {
			fmt.Printf("Enter the SMTP Server port for %s: ", email)
			tmp, err := strconv.Atoi(ReadStringTrimmed(reader))
			if err != nil {
				fmt.Println("Error: Not an integer. Please try again.")
			} else {
				smtpport = tmp
				break
			}
		}
		if askName {
			fmt.Print("(Optional) Enter a comment: ")
			comment = ReadStringTrimmed(reader)
		}
		fmt.Println("Please confirm the below information is correct:")
		fmt.Printf("Name: %s\n", name)
		fmt.Printf("Email: %s\n", email)
		fmt.Printf("IMAP Server: %s\n", imapserver)
		fmt.Printf("IMAP Port: %d\n", imapport)
		fmt.Printf("SMTP Server: %s\n", smtpserver)
		fmt.Printf("SMTP Port :%d\n", smtpport)
		fmt.Printf("Comment: %s\n", comment)
		fmt.Print("Is this correct (y/n)? ")
		confirm := ReadStringTrimmed(reader)
		if confirm == "y" {
			break
		}
	}

	// Write configuration
	configFile := Config{
		IMAPServer: imapserver,
		IMAPPort: imapport,
		SMTPServer: smtpserver,
		SMTPPort: smtpport,
		Email: email,
	}

	return configFile, name, comment
}

func CreateAccount() {
	configFile, name, comment := AskInfo(true)

	// Generate email keypair
	config := openpgp.Config{Expiry: 365 * 24 * time.Hour}
	emailKeypair, err := openpgp.CreateKey(name, comment, configFile.Email, &config)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	
	err = os.MkdirAll(usr.HomeDir + "/.pangolin/email", 0700)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	pubring := emailKeypair.Pubring()
	secring := emailKeypair.Secring()
	
	err = ioutil.WriteFile(usr.HomeDir +"/.pangolin/email/pubring.gpg", pubring, 0600)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(usr.HomeDir +"/.pangolin/email/secring.gpg", secring, 0600)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Generate the pangolin PGP keypair
	uuidString := GenPangolinPGP(&config, usr.HomeDir)
	configFile.UUID = uuidString

	WriteConfig(configFile)
}

func WriteConfig(conf Config) {
	confJSON, err := json.Marshal(conf)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(usr.HomeDir + "/.pangolin/config.json", confJSON, 0600)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func ReadConfig() (conf Config) {
	if CheckConfigStatus() {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		dat, err := ioutil.ReadFile(usr.HomeDir + "/.pangolin/config.json")
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		conf = Config{}
		err = json.Unmarshal(dat, &conf)
		return
	} else {
		log.Fatal("No configuration file was found. Please make an account using pangolin --setup.")
		os.Exit(1)
		return
	}
}

func CheckConfigStatus() bool {
	// Check to see if config file exists
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	
	_, err = os.Stat(usr.HomeDir + "/.pangolin/config.json")

	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
