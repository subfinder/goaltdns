package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/bobesa/go-domain-util/domainutil"
)

// AltDNS holds words, etc
type AltDNS struct {
	PermutationWords []string
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

// New Returns a new altdns object
func New(wordList string) (*AltDNS, error) {
	altdns := AltDNS{}

	f, err := os.Open(wordList)
	if err != nil {
		return &altdns, err
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		altdns.PermutationWords = append(altdns.PermutationWords, scanner.Text())
	}

	return &altdns, nil
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
	var wordlist, host, list string
	hostList := []string{}
	flag.StringVar(&host, "host", "", "Host to generate permutations for")
	flag.StringVar(&list, "l", "", "List of hosts to generate permutations for")
	flag.StringVar(&wordlist, "w", "words.txt", "Wordlist to generate permutations with")

	flag.Parse()

	if host == "" && list == "" {
		fmt.Printf("%s: no host/hosts specified!\n", os.Args[0])
		os.Exit(1)
	}

	if host != "" {
		hostList = append(hostList, host)
	} else if list != "" {
		f, _ := os.Open(list)
		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			hostList = append(hostList, scanner.Text())
		}
	}

	altdns := New(wordlist)
	jobs := sync.WaitGroup{}

	for _, u := range hostList {
		subdomain := domainutil.Subdomain(u)
		domainSuffix := domainutil.Domain(u)
		jobs.Add(1)
		go func(domain string) {
			defer jobs.Done()
			for r := range altdns.Permute(subdomain) {
				fmt.Printf("%s.%s\n", r, domainSuffix)
			}
		}(u)
	}

	jobs.Wait()
}
