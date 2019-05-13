package openpgp

import (
	"io"
	"log"
	"strings"

	"github.com/emersion/go-message"
	"golang.org/x/crypto/openpgp"
)

func decryptEntity(msgWriter *message.Writer, ciphertext *message.Entity, kr openpgp.KeyRing) error {
	if msgReader := ciphertext.MultipartReader(); msgReader != nil {
		for {
			part, err := msgReader.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			partWriter, err := msgWriter.CreatePart(part.Header)
			if err != nil {
				return err
			}

			if err := decryptEntity(partWriter, part, kr); err != nil {
				log.Println("WARN: cannot decrypt child part:", err)
			}
			partWriter.Close()
		}
	} else {
		// A normal part which is to be decrypted
		mediaType, _, err := ciphertext.Header.ContentType()
		if err != nil {
			log.Println("WARN: cannot parse ContentType:", err)
			mediaType = "text/plain"
		}
		isPlainText := strings.HasPrefix(mediaType, "text/")
		var md *openpgp.MessageDetails
		if mediaType == "application/pgp-encrypted" {
			// The message is an encrypted binary
			md, err = decrypt(ciphertext.Body, kr)
		} else if isPlainText {
			// The message is possibly encrypted with inline PGP
			md, err = decryptArmored(ciphertext.Body, kr)
		} else {
			// An unencrypted binary part
			md = &openpgp.MessageDetails{UnverifiedBody: ciphertext.Body}
			err = nil
		}
		if err != nil {
			return err
		}

		if _, err := io.Copy(msgWriter, md.UnverifiedBody); err != nil {
			return err
		}

		// Fail if signature is incorrect
		if err := md.SignatureError; err != nil {
			return err
		}
	}

	return nil
}

func Decrypt(w io.Writer, r io.Reader, kr openpgp.KeyRing) error {
	ciphertext, err := message.Read(r)
	if err != nil {
		return err
	}

	msgWriter, err := message.CreateWriter(w, ciphertext.Header)
	if err != nil {
		return err
	}

	if err := decryptEntity(msgWriter, ciphertext, kr); err != nil {
		return err
	}
	return msgWriter.Close()
}

func encryptEntity(mw *message.Writer, e *message.Entity, to []*openpgp.Entity, signed *openpgp.Entity) error {
	if mr := e.MultipartReader(); mr != nil {
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}

			if err != nil {
				return nil
			}

			pw, err := mw.CreatePart(e.Header)
			if err != nil {
				return err
			}

			if err := encryptEntity(pw, p, to, signed); err != nil {
				return err
			}

			pw.Close()
			
		}
	} else {
		// A normal part; just encrypt it

		mediaType, _, err := e.Header.ContentType()
		if err != nil {
			log.Println("WARN: cannot parse Content-Type:", err)
			mediaType = "text/plain"
		}

		disp, _, err := e.Header.ContentDisposition()
		if err != nil {
			log.Println("WARN: cannot parse Content-Disposition:", err)
		}

		var plaintext io.WriteCloser
		if strings.HasPrefix(mediaType, "text/") && disp != "attachment" {
			// The message text, encrypt it with inline PGP
			plaintext, err = encryptArmored(mw, to, signed)
		} else {
			plaintext, err = encrypt(mw, to, signed)
		}

		if err != nil {
			return err
		}
		defer plaintext.Close()

		if _, err := io.Copy(plaintext, e.Body); err != nil {
			return err
		}
	}

	return nil
}

func Encrypt(w io.Writer, r io.Reader, to []*openpgp.Entity, signed *openpgp.Entity) error {
	e, err := message.Read(r)
	if err != nil {
		return err
	}

	mw, err := message.CreateWriter(w, e.Header)
	if err != nil {
		return err
	}

	if err := encryptEntity(mw, e, to, signed); err != nil {
		return err
	}

	return mw.Close()
}

	
