package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func Parse() {
	fileName := "_lua5.1-tests/literals.lua"
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	l := InitLexer(bufio.NewReader(f), fileName)
	keywordsToken2Str := make(map[tokenType]string)

	for k, v := range keywordsStr2Token {
		keywordsToken2Str[v] = k
	}

	for {
		t, err := l.Scan()

		if err != nil {
			fmt.Println(err)
			break
		}

		if t.typ > 1<<8 {
			fmt.Printf("line %d column(%d) %s\t%s\n", t.pos.line, t.pos.column, strings.ToUpper(tokenName[t.typ]), t.val)
		} else {
			fmt.Printf("line %d column(%d) %s\n", t.pos.line, t.pos.column, strings.ToUpper(keywordsToken2Str[t.typ]))
		}
	}
}
