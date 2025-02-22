package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// resolveCDN queries the DNS server for the CDN domain and returns the resolved IP.
func resolveCDN(dnsServer string) (string, error) {
	dnsResolver := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.Dial("udp", dnsServer+":53")
		},
	}

	domain := "cdn.local"

	ips, err := dnsResolver.LookupHost(nil, domain)
	if err != nil {
		return "", err
	}

	if len(ips) > 0 {
		log.Printf("Resolved %s to %s", domain, ips[0])
		return ips[0], nil
	}
	return "", fmt.Errorf("no IP resolved")
}

// makeRequest sends a request to the resolved CDN load balancer.
func makeRequest(targetIP string) {
	url := fmt.Sprintf("http://%s/image/sample.jpg", targetIP)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching image from %s: %v", targetIP, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Client [%s] received response: %s", os.Getenv("REGION"), resp.Status)
}

func main() {
	dnsServer := os.Getenv("DNS_SERVER") // The BIND9 server IP
	region := os.Getenv("REGION")        // Client's region (Ukraine or Europe)

	log.Printf("Client from [%s] querying DNS server %s for CDN resolution...", region, dnsServer)

	targetIP, err := resolveCDN(dnsServer)
	if err != nil {
		log.Fatalf("Failed to resolve CDN: %v", err)
	}

	for {
		makeRequest(targetIP)
		time.Sleep(5 * time.Second)
	}
}
