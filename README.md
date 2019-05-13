# Pangolin

Pangolin is an automated, multidevice PGP solution.

## Building
### Go
`pangolin` is implemented in Go. Please visit the [Go website](https://golang.org/) for setup instructions.

### Dependencies
The `go` binary will handle the installation of all dependencies. However, `pangolin` requires at least version
`1.10`.

### Installation
To install `pangolin`, simply run the command
```
go get github.com/zsusag/pangolin/cmd/pangolin
```
This will download the `pangolin` source code, all necessary dependencies, and install the compiled binary into your `GOPATH`. If this is the first time using a Go application, you may need to set your `GOPATH`. I recommend setting it to `GOPATH=$HOME/go`. Furthermore, append the `GOROOT` and `GOPATH` directories to your `PATH` environment variable:
```
PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```
This will ensure that the compiled binary may be accessed normally like any other program on your computer.

## Usage
| Command | Description |
| --- | --- |
| `pangolin` | The default mode; Start a local IMAP and SMTP server that will encrypt all outgoing emails using PGP and decrypt all PGP-encrypted emails |
| `pangolin --setup` | Create a new PGP keypair. Currently, a new keypair needs to be made as import functionality has not been implemented |
| `pangolin --new-device` | Add a new device to the Pangolin network. This command needs to be run on both the new and existing devices|

### Usage Example
First, we will create a new PGP keypair for a test email account, `pangolin@example.com`. First, run
```
$ pangolin --setup
Enter your name: Cute Pangolin
Enter your email: pangolin@example.com
Enter the IMAP Server for pangolin@example.com: imap.example.com
Enter the IMAP Server port for pangolin@example.com: 993
Enter the SMTP server for pangolin@example.com: smtp.example.com
Enter the SMTP Server port for pangolin@example.com: 465
(Optional) Enter a comment: I wish Pangolins weren't an endangered species.
Please confirm the below information is correct:
Name: Cute Pangolin
Email: pangolin@example.com
IMAP Server: imap.example.com
IMAP Port: 993
SMTP Server: smtp.example.com
SMTP Port 465
Comment: I wish Pangolins weren't an endangered species.
Is this correct (y/n)? y

```

We are now fully set up! If you only have a single device, just run `pangolin` to start the IMAP and SMTP servers.
```
$ pangolin
2019/05/13 04:41:48 Starting IMAP server at localhost:1143
2019/05/13 04:41:48 Starting SMTP server at localhost:1025

```

Now, direct your favorite email client to talk to the above IMAP and SMTP servers. Your email credentials will be transferred to `pangolin` via the email client automatically.

Once completed, any emails that you send will be intercepted by `pangolin`, who will then search PGP keyservers for a corresponding public key, encrypt the email using the public key, and send the encrypted version. If a public key could not be found, the email is simply send in plaintext.

Any email that is received will be scanned by `pangolin` for an inline PGP message. If such a message is found, `pangolin` will decrypt the message and forward the decrypted version to your email client.

Pangolin can also handle multiple devices. To add a new device, have your existing device and your new device handy. On the existing device, run
```
$ pangolin --new-device
Waiting for an email from your new device...
```

Now, on your new device, run the same command:
```
$ pangolin --new-device
Enter your email: pangolin@example.com
Enter the IMAP Server for pangolin@example.com: imap.example.com
Enter the IMAP Server port for pangolin@example.com: 993
Enter the SMTP server for pangolin@example.com: smtp.example.com
Enter the SMTP Server port for pangolin@example.com: 465
Please confirm the below information is correct:
Email: pangolin@example.com
IMAP Server: imap.example.com
IMAP Port: 993
SMTP Server: smtp.example.com
SMTP Port 465
Is this correct (y/n)? y

Enter the following phrase on your existing device: french new notepad
Waiting for a reply...
```

After a few moments, on your existing device you should see the following:
```
$ pangolin --new-device
Waiting for an email from your new device...

Enter the verification phrase: 
```

Enter in the verification phrase displayed on your new device. If correct, you will see the following:
```
$ pangolin --new-device
Waiting for an email from your new device...

Enter the verification phrase: french new notepad
Enter the new verification phrase on your new device: phone frontier shale
```

Now, on your new device, enter the new verification phrase to complete the verification process.
$ pangolin --new-device
...

Enter the following phrase on your existing device: french new notepad
Waiting for a reply...
Enter the verification phrase: phone frontier shale
```

You have now successfully added a new device to your Pangolin network.

## Limitations
- Only inline PGP messages will be decrypted and encrypted.
- In order for the decrypted message to appear in your mail client, your mail client must be configured to only display plaintext, not HTML.
- Importing of existing PGP keys is not yet supported.
- Revokation of devices is not yet supported.

## License
GPLv3
