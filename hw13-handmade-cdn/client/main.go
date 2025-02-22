package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

var dnsServer string
var region string

// resolveCDN queries the DNS server for the CDN domain and returns the resolved IP.
func resolveCDN(dnsServer string) (string, error) {
	dnsResolver := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.Dial("udp", dnsServer+":53")
		},
	}

	domain := "cdn.local"

	ips, err := dnsResolver.LookupHost(context.Background(), domain)
	if err != nil {
		return "", fmt.Errorf("DNS resolution failed: %w", err)
	}

	if len(ips) > 0 {
		log.Printf("Resolved %s to %s", domain, ips[0])
		return ips[0], nil
	}
	return "", fmt.Errorf("no IP resolved for %s", domain)
}

// handleRequest is an HTTP endpoint that triggers a single image request.
func handleRequest(w http.ResponseWriter, r *http.Request) {
	targetIP, err := resolveCDN(dnsServer)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to resolve CDN: %v", err), http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("http://%s/image/sample.jpg", targetIP)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching image from %s: %v", targetIP, err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	log.Printf("Client [%s] received response from %s: %s", region, targetIP, resp.Status)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Request to %s successful: %s", targetIP, resp.Status)))
}

func main() {
	dnsServer = os.Getenv("DNS_SERVER") // The BIND9 server IP
	region = os.Getenv("REGION")        // Client's region (Ukraine or Europe)

	log.Printf("Client from [%s] is ready to query DNS server %s", region, dnsServer)

	http.HandleFunc("/request-image", handleRequest)

	log.Println("Client server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}