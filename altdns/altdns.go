package altdns

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
)

var (
	nbrRe = regexp.MustCompile("[0-9]+")
)

// AltDNS holds words, etc
type AltDNS struct {
	PermutationWords []string
}

func (a *AltDNS) insertDashes(domain string, results chan string) {
	for _, w := range a.PermutationWords {
		if w == "" || domain == "" {
			continue
		}
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
		if w == "" || domain == "" {
			continue
		}
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
	if domain != "" {
		for j := 0; j < 10; j++ {
			// suffixes
			results <- fmt.Sprintf("%s-%d", domain, j)
		}
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

func (a *AltDNS) expandNumbers(domain string, results chan string) {
	for _, ind := range nbrRe.FindAllStringIndex(domain, -1) {
		padSize := strconv.Itoa(ind[1] - ind[0])
		for i := 1; i <= 10; i++ {
			results <- fmt.Sprintf("%s%0"+padSize+"d%s", domain[:ind[0]], i, domain[ind[1]:])
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

		// Permute numbers 0x -> 01, 02, 03, ...
		wg.Add(1)
		go func(domain string, results chan string) {
			defer wg.Done()
			a.expandNumbers(domain, results)
		}(domain, results)

		wg.Wait()
	}(domain)

	return results
}

// Permutations permutes a given domain and return a list of permutations
func (a *AltDNS) Permutations(domain string) (permutations []string) {
	uniq := make(map[string]bool)
	for r := range a.Permute(domain) {
		// avoid duplicates
		if _, ok := uniq[r]; ok {
			continue
		}

		uniq[r] = true
		permutations = append(permutations, r)
	}
	return permutations
}
