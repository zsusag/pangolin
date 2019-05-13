package imap

import (
	"bytes"
	"io"

	"golang.org/x/crypto/openpgp"

	myopenpgp "github.com/zsusag/pangolin/openpgp"
)

func decryptMessage(kr openpgp.KeyRing, r io.Reader) (io.Reader, error) {
	b := new(bytes.Buffer)
	if err := myopenpgp.Decrypt(b,r,kr); err != nil {
		return nil, err
	}

	return b, nil
}
