package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	Dictionary = "words.txt"
)

type AltDNS struct {
}

func (a *AltDNS) Permute(domain string) {

	// Read all permutations words
	var permutationWords []string
	f, _ := os.Open(Dictionary)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		permutationWords = append(permutationWords, scanner.Text())
	}

	// Grabs placeholders
	var preSub, postSub string
	subParts := strings.SplitN(domain, ".", 2)
	preSub = subParts[0]
	if len(subParts) > 1 {
		postSub = subParts[1]
	}

	// Insert all indexes
	for i, rune := range domain {
		if rune == '.' {
			for _, w := range permutationWords {
				fmt.Println(domain[:i] + "." + w + domain[i:])
			}
		}
	}

	// Insert all dash
	for _, w := range permutationWords {
		fmt.Println(w + "-" + domain)
		fmt.Println(domain + "-" + w)
		fmt.Println(preSub + "-" + w + "." + postSub)
	}

	// Insert Number Suffix Subdomains
	for i := 0; i < 10; i++ {
		fmt.Println(preSub + "-" + string(i) + "." + postSub)
		fmt.Println(preSub + string(i) + "." + postSub)
	}

	// Join Words Subdomains
	for _, w := range permutationWords {
		fmt.Println(w + preSub + "." + postSub)
		fmt.Println(preSub + w + "." + postSub)
	}

}

func main() {
	urls := []string{"aa.bb.cc.dd", "www.kk.cc.dd.oo.pp", "g.a.b.c.d.e"}

	altnds := &AltDNS{}

	for _, u := range urls {
		altnds.Permute(u)
	}
}
