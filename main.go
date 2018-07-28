package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/bobesa/go-domain-util/domainutil"
)

const (
	Dictionary = "words.txt"
)

// AltDNS holds words, etc
type AltDNS struct {
	PermutationWords []string
	OutputChan       chan<- string
}

func (a *AltDNS) insertDashes(domain string, results chan string) {
	for _, w := range a.PermutationWords {
		// prefixes
		results <- fmt.Sprint(w + "-" + domain)
		// suffixes
		results <- fmt.Sprint(domain + "-" + w)
	}

	for i, rune := range domain {
		if rune == '.' {
			for _, w := range a.PermutationWords {
				results <- fmt.Sprint(domain[:i] + "." + w + "-" + domain[i+1:])
				results <- fmt.Sprintf(domain[:i] + "-" + w + domain[i:])
			}
		}
	}
}

func (a *AltDNS) insertIndexes(domain string, results chan string) {
	for _, w := range a.PermutationWords {
		// prefixes
		results <- fmt.Sprint(w + "." + domain)
		// suffixes
		results <- fmt.Sprint(domain + "." + w)
	}

	for i, rune := range domain {
		if rune == '.' {
			for _, w := range a.PermutationWords {
				results <- fmt.Sprint(domain[:i] + "." + w + domain[i:])
			}
		}
	}
}

func (a *AltDNS) insertNumberSuffixes(domain string, results chan string) {
	for j := 0; j < 10; j++ {
		// suffixes
		results <- fmt.Sprintf("%s-%d", domain, j)
	}

	for i, rune := range domain {
		if rune == '.' {
			for j := 0; j < 10; j++ {
				results <- fmt.Sprintf("%s-%d%s", domain[:i], j, domain[i:])
				results <- fmt.Sprintf("%s%d%s", domain[:i], j, domain[i:])
			}
		}
	}
}

func (a *AltDNS) insertWordsSubdomains(domain string, results chan string) {
	for _, w := range a.PermutationWords {
		// prefixes
		results <- fmt.Sprint(w + domain)
		// suffixes
		results <- fmt.Sprint(domain + w)
	}

	for i, rune := range domain {
		if rune == '.' {
			for _, w := range a.PermutationWords {
				results <- fmt.Sprint(domain[:i] + w + domain[i:])
				results <- fmt.Sprint(domain[:i] + "." + w + domain[i+1:])
			}
		}
	}
}

// Permute permutes a given domain and sends output on a channel
func (a *AltDNS) Permute(domain string) chan string {
	wg := sync.WaitGroup{}
	results := make(chan string)

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
		go func(domain string, results chan string) {
			defer wg.Done()
			a.insertDashes(domain, results)
		}(domain, results)

		// Insert Number Suffix Subdomains
		wg.Add(1)
		go func(domain string, results chan string) {
			defer wg.Done()
			a.insertNumberSuffixes(domain, results)
		}(domain, results)

		// Join Words Subdomains
		wg.Add(1)
		go func(domain string, results chan string) {
			defer wg.Done()
			a.insertWordsSubdomains(domain, results)
		}(domain, results)

		wg.Wait()
	}(domain)

	return results
}

func main() {

	urls := []string{"aa.bb.cc.dd.com", "www.ee.com"}

	altnds := &AltDNS{}

	f, _ := os.Open(Dictionary)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		altnds.PermutationWords = append(altnds.PermutationWords, scanner.Text())
	}

	jobs := sync.WaitGroup{}

	for _, u := range urls {
		subdomain := domainutil.Subdomain(u)
		domainSuffix := domainutil.Domain(u)
		jobs.Add(1)
		go func(domain string) {
			defer jobs.Done()
			for r := range altnds.Permute(subdomain) {
				fmt.Printf("%s.%s\n", r, domainSuffix)
			}
		}(u)
	}

	jobs.Wait()
}
