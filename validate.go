package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	listFile = flag.String("listFile", "/etc/pihole/adlists.list", "pihole adlists file")
	ipv4     = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`)
)

// all our allowances
type allowances []allowance

// allowance lets an ip with optional hostname past the filter
type allowance struct {
	ip       string
	hostname string
}

// what hostnames aren't allow to appear from the pi-hole lists?
// eg 8.8.8.8 will forbid any entry with that IP
type restriction string
type restrictions []string

// validateLine takes an /etc/hosts
// returns true if the line is valid
func validateLine(line string, a allowances, r restrictions) bool {
	//fmt.Printf("validateLine: line=x%sx\n", line)
	if len(line) == 0 || line[0] == '#' || line[0] == ' ' || line[0] == '\n' {
		return true
	}

	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return true
	}
	// this will skip ipv6 and treat them all as valid.
	matchedIP := ipv4.FindString(parts[0])
	if matchedIP != "" {
		// it's an ipv4 address
		//fmt.Printf("Matched IP = %v\n", matchedIP)
		for _, allowed := range a {
			if (allowed.hostname != "" && parts[1] == allowed.hostname) || matchedIP == allowed.ip {
				//fmt.Printf("Allowing\n")
				return true
			}
		}
		for _, restrict := range r {
			if matchedIP == restrict {
				//fmt.Printf("Forbidding\n")
				return false
			}
		}
	}
	return true
}

func main() {
	flag.Parse()

	allowed := []allowance{
		allowance{
			ip:       "255.255.255.255",
			hostname: "broadcasthost",
		},
		allowance{
			ip: "0.0.0.0",
		},
		allowance{
			ip: "127.0.0.1",
		},
	}
	restricted := restrictions{}

	input, err := os.Open(*listFile)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	defer input.Close()

	lineReader := bufio.NewScanner(input)
	for lineReader.Scan() {
		url := lineReader.Text()
		if url[0] == '#' {
			continue
		}
		response, err := http.Get(url)
		if err != nil {
			fmt.Printf("Couldn't fetch %s (which is probably bad): %s\n", url, err)
			os.Exit(0)
		}
		defer response.Body.Close()

		lr := bufio.NewScanner(response.Body)
		for lr.Scan() {
			line := lr.Text()
			if !validateLine(line, allowed, restricted) {
				fmt.Printf("failed validation on %s in %s\n", line, url)
				os.Exit(1)
			}
		}
	}
	fmt.Printf("All valid.\n")
}
