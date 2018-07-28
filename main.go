package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

const (
	Dictionary = "words.txt"
)

// AltDNS holds words, etc
type AltDNS struct {
	PermutationWords []string
	OutputChan       chan<- string
}

func (a *AltDNS) insertDashes(domain string, preSub string, postSub string, results chan string) {
	for _, w := range a.PermutationWords {
		results <- fmt.Sprintf(w + "-" + domain)
		results <- fmt.Sprintf(preSub + "-" + w + "." + postSub)
	}
}

func (a *AltDNS) insertIndexes(domain string, results chan string) {
	for i, rune := range domain {
		if rune == '.' {
			for _, w := range a.PermutationWords {
				results <- fmt.Sprintf(domain[:i] + "." + w + domain[i:])
			}
		}
	}
}

func (a *AltDNS) insertNumberSuffixes(domain string, preSub string, postSub string, results chan string) {
	for i := 0; i < 10; i++ {
		results <- fmt.Sprintf("%s-%d.%s", preSub, i, postSub)
		results <- fmt.Sprintf("%s%d.%s", preSub, i, postSub)
	}
}

func (a *AltDNS) insertWordsSubdomains(domain string, preSub string, postSub string, results chan string) {
	for _, w := range a.PermutationWords {
		results <- fmt.Sprintf(w + preSub + "." + postSub)
		results <- fmt.Sprintf(preSub + w + "." + postSub)
	}
}

// Permute permutes a given domain and sends output on a channel
func (a *AltDNS) Permute(domain string) chan string {
	wg := sync.WaitGroup{}
	results := make(chan string)

	var preSub, postSub string
	subParts := strings.SplitN(domain, ".", 2)
	preSub = subParts[0]
	if len(subParts) > 1 {
		postSub = subParts[1]
	}

	go func(domain string) {
		defer close(results)

		// Insert all indexes
		wg.Add(1)
		go func(domain string, results chan string) {
			defer wg.Done()
			a.insertIndexes(domain, results)
		}(domain, results)

		// Insert all dash
		wg.Add(1)
		go func(domain string, preSub string, postSub string, results chan string) {
			defer wg.Done()
			a.insertDashes(domain, preSub, postSub, results)
		}(domain, preSub, postSub, results)

		// Insert Number Suffix Subdomains
		wg.Add(1)
		go func(domain string, preSub string, postSub string, results chan string) {
			defer wg.Done()
			a.insertNumberSuffixes(domain, preSub, postSub, results)
		}(domain, preSub, postSub, results)

		// Join Words Subdomains
		wg.Add(1)
		go func(domain string, preSub string, postSub string, results chan string) {
			defer wg.Done()
			a.insertWordsSubdomains(domain, preSub, postSub, results)
		}(domain, preSub, postSub, results)

		wg.Wait()
	}(domain)

	return results
}

func main() {
	urls := []string{"aa.bb.cc"}

	altnds := &AltDNS{}

	f, _ := os.Open(Dictionary)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		altnds.PermutationWords = append(altnds.PermutationWords, scanner.Text())
	}

	jobs := sync.WaitGroup{}

	for _, u := range urls {
		jobs.Add(1)
		go func(domain string) {
			defer jobs.Done()
			for result := range altnds.Permute(domain) {
				fmt.Printf("%s\n", result)
			}
		}(u)
	}

	jobs.Wait()
}
