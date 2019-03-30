package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/subfinder/goaltdns/altdns"
	"github.com/subfinder/goaltdns/util"
)

func main() {
	var wordlist, host, list, output string
	hostList := []string{}
	flag.StringVar(&host, "h", "", "Host to generate permutations for")
	flag.StringVar(&list, "l", "", "List of hosts to generate permutations for")
	flag.StringVar(&wordlist, "w", "words.txt", "Wordlist to generate permutations with")
	flag.StringVar(&output, "o", "", "File to write permutation output to (optional)")

	flag.Parse()

	if host == "" && list == "" && !util.PipeGiven() {
		fmt.Printf("%s: no host/hosts specified!\n", os.Args[0])
		os.Exit(1)
	}

	if host != "" {
		hostList = append(hostList, host)
	}

	if list != "" {
		hostList = append(hostList, util.LinesInFile(list)...)
	}

	if util.PipeGiven() {
		hostList = append(hostList, util.LinesInStdin()...)
	}

	var f *os.File
	var err error
	if output != "" {
		f, err = os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Printf("output: %s\n", err)
			os.Exit(1)
		}

		defer f.Close()
	}

	altdns, err := altdns.New(wordlist)
	if err != nil {
		fmt.Printf("wordlist: %s\n", err)
		os.Exit(1)
	}

	writerJob := sync.WaitGroup{}

	writequeue := make(chan string)

	writerJob.Add(1)
	go func() {
		defer writerJob.Done()

		w := bufio.NewWriter(f)
		defer w.Flush()

		for permutation := range writequeue {
			w.WriteString(permutation)
		}
	}()

	jobs := sync.WaitGroup{}

	for _, u := range hostList {
		subdomain := domainutil.Subdomain(u)
		domainSuffix := domainutil.Domain(u)
		jobs.Add(1)
		go func(domain string) {
			defer jobs.Done()
			uniq := make(map[string]bool)
			for r := range altdns.Permute(subdomain) {
				permutation := fmt.Sprintf("%s.%s\n", r, domainSuffix)

				// avoid duplicates
				if _, ok := uniq[permutation]; ok {
					continue
				}

				uniq[permutation] = true

				if output == "" {
					fmt.Printf("%s", permutation)
				} else {
					writequeue <- permutation
				}
			}
		}(u)
	}

	jobs.Wait()

	close(writequeue)

	writerJob.Wait()
}
