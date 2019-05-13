package verification

import (
	"crypto/rand"
	"strings"
	"log"
	"os"
	"io/ioutil"
	"math/big"
)

type PhraseGenerator struct {
	Count int
	Separator string
	Words []string
}

func readAvailableDictionary () (words []string) {
	file, err := os.Open("/usr/share/dict/words")
	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	words = strings.Split(string(bytes), "\n")
	return
}

func NewPhraseGenerator() (pg PhraseGenerator) {
	pg.Count = 3
	pg.Separator = " "
	pg.Words = readAvailableDictionary()
	return
}

func (pg PhraseGenerator) GenPhrase() string {
	words := []string{}
	for i := 0; i < pg.Count; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(pg.Words))))
		if err != nil {
			log.Fatal("Unable to generate random index")
			os.Exit(1)
		}
		words = append(words, pg.Words[index.Int64()])
	}

	return strings.Join(words, pg.Separator)
}

