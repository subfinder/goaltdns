package altdns

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bobesa/go-domain-util/domainutil"
)

func TestPermute(t *testing.T) {
	urls := []string{"abc.xyz.freelancer.com", "aa.bb.cc"}

	altdns, _ := New("words.txt")

	jobs := sync.WaitGroup{}

	for _, u := range urls {
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
