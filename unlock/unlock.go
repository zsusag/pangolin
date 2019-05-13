package unlock

import (
	"os/user"
	"log"
	"sync"
	
	"golang.org/x/crypto/openpgp"

	myopenpgp "github.com/zsusag/pangolin/openpgp"
)

type UnlockFunction func(username, password string) (openpgp.EntityList, error)

func Unlock(username, _ string) (openpgp.EntityList, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return myopenpgp.ReadKeyring(usr.HomeDir + "/.pangolin/email/secring.gpg", ""), nil
}

func UnlockRemember(f UnlockFunction) UnlockFunction {
	cache := map[string]openpgp.EntityList{}
	return func(username, password string) (openpgp.EntityList, error) {
		if kr, ok := cache[username]; ok {
			return kr,nil
		}

		kr, err := f(username, password)
		if err != nil {
			return nil, err
		}

		cache[username] = kr
		return kr, nil
	}
}

func UnlockSync(f UnlockFunction) UnlockFunction {
	locker := &sync.Mutex{}
	return func(username, password string) (openpgp.EntityList, error) {
		locker.Lock()
		defer locker.Unlock()

		return f(username, password)
	}
}
