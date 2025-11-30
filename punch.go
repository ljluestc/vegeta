package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type punchOpts struct {
	targetsf      string
	format        string
	headers       headers
	proxyHeaders  headers
	laddr         localAddr
	maxBody       int64
	timeout       time.Duration
	keepAlive     bool
	insecure      bool
	h2c           bool
	http2         bool
	redirects     int
	resolvers     csl
	connectTo     connectToFlag
	rootCerts     csl
	certf         string
	keyf          string
	unixSocket    string
	verbose       bool
	stopOnError   bool
}

func punchCmd() command {
	fs := flag.NewFlagSet("vegeta punch", flag.ExitOnError)
	opts := &punchOpts{
		headers:      headers{http.Header{}},
		proxyHeaders: headers{http.Header{}},
		laddr:        localAddr{&vegeta.DefaultLocalAddr},
		maxBody:      vegeta.DefaultMaxBody,
		timeout:      30 * time.Second,
		redirects:    10,
		keepAlive:    true,
		http2:        true,
	}
	
	fs.StringVar(&opts.targetsf, "targets", "stdin", "Targets file")
	fs.StringVar(&opts.format, "format", vegeta.HTTPTargetFormat,
		fmt.Sprintf("Targets format [%s]", strings.Join(vegeta.TargetFormats, ", ")))
	fs.Var(&opts.headers, "header", "Request header")
	fs.Var(&opts.proxyHeaders, "proxy-header", "Proxy CONNECT header")
	fs.Var(&opts.laddr, "laddr", "Local IP address")
	fs.Var(&maxBodyFlag{&opts.maxBody}, "max-body", "Maximum number of bytes to capture from response bodies")
	fs.DurationVar(&opts.timeout, "timeout", opts.timeout, "Requests timeout")
	fs.BoolVar(&opts.keepAlive, "keepalive", opts.keepAlive, "Use persistent connections")
	fs.BoolVar(&opts.insecure, "insecure", opts.insecure, "Ignore invalid server TLS certificates")
	fs.BoolVar(&opts.h2c, "h2c", opts.h2c, "Send HTTP/2 requests without TLS encryption")
	fs.BoolVar(&opts.http2, "http2", opts.http2, "Send HTTP/2 requests when supported by the server")
	fs.IntVar(&opts.redirects, "redirects", opts.redirects, "Number of redirects to follow")
	fs.Var(&opts.resolvers, "resolvers", "List of addresses (ip:port) to use for DNS resolution")
	fs.Var(&opts.connectTo, "connect-to", "Custom address to use for connections")
	fs.Var(&opts.rootCerts, "root-certs", "TLS root certificate files")
	fs.StringVar(&opts.certf, "cert", opts.certf, "TLS client PEM encoded certificate file")
	fs.StringVar(&opts.keyf, "key", opts.keyf, "TLS client PEM encoded private key file")
	fs.StringVar(&opts.unixSocket, "unix-socket", opts.unixSocket, "Connect over a unix socket")
	fs.BoolVar(&opts.verbose, "verbose", false, "Verbose output showing request/response details")
	fs.BoolVar(&opts.stopOnError, "stop-on-error", false, "Stop on first error")
	
	return command{fs, func(args []string) error {
		fs.Parse(args)
		return punch(opts)
	}}
}

func punch(opts *punchOpts) error {
	// Read targets
	targets, err := readTargets(opts)
	if err != nil {
		return fmt.Errorf("failed to read targets: %w", err)
	}
	
	if len(targets) == 0 {
		return fmt.Errorf("no targets found")
	}
	
	fmt.Printf("👊 Punching %d target(s) sequentially...\n\n", len(targets))
	
	// Create HTTP client
	client, err := createPunchClient(opts)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	
	// Execute each target sequentially
	var success, failed int
	var totalTime time.Duration
	
	for i, target := range targets {
		fmt.Printf(" [%d/%d] %s %s", i+1, len(targets), target.Method, target.URL)
		
		start := time.Now()
		result, err := executePunch(client, target, opts)
		duration := time.Since(start)
		totalTime += duration
		
		if err != nil {
			fmt.Printf(" ❌\n     Error: %v\n", err)
			failed++
			if opts.stopOnError {
				fmt.Printf("\n🛑 Stopping due to error (stop-on-error enabled)\n")
				break
			}
		} else {
			fmt.Printf(" ✅ %dms [%s]\n", duration.Milliseconds(), result.Status)
			if opts.verbose && result.Details != "" {
				fmt.Printf("     %s\n", result.Details)
			}
			success++
		}
		
		// Add small delay between requests to be gentle
		if i < len(targets)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	// Print summary
	fmt.Printf("\n📊 Punch Summary:\n")
	fmt.Printf("   Total:   %d\n", len(targets))
	fmt.Printf("   Success: %d\n", success)
	fmt.Printf("   Failed:  %d\n", failed)
	fmt.Printf("   Time:    %v\n", totalTime)
	if success > 0 {
		fmt.Printf("   Avg:     %v\n", totalTime/time.Duration(success))
	}
	
	if failed > 0 {
		return fmt.Errorf("%d targets failed", failed)
	}
	
	return nil
}

type punchResult struct {
	Status  string
	Details string
}

func readTargets(opts *punchOpts) ([]*vegeta.Target, error) {
	// Use the same file reading approach as attack command
	reader, err := fileReader(opts.targetsf)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	var targeter vegeta.Targeter
	switch opts.format {
	case vegeta.JSONTargetFormat:
		targeter = vegeta.NewJSONTargeter(reader, nil, opts.headers.Header)
	case vegeta.HTTPTargetFormat:
		targeter = vegeta.NewHTTPTargeter(reader, nil, opts.headers.Header)
	default:
		return nil, fmt.Errorf("unsupported format: %s", opts.format)
	}
	
	var targets []*vegeta.Target
	for {
		target := &vegeta.Target{}
		err := targeter(target)
		if err != nil {
			if err.Error() == "no targets to attack" {
				break // This is the actual EOF condition
			}
			return nil, fmt.Errorf("targeter error: %w", err)
		}
		if target.URL == "" {
			continue // Skip empty targets
		}
		targets = append(targets, target)
	}
	
	if len(targets) == 0 {
		return nil, fmt.Errorf("no targets to attack")
	}
	
	return targets, nil
}

func createPunchClient(opts *punchOpts) (*http.Client, error) {
	// Create TLS config
	tlsc, err := tlsConfig(opts.insecure, opts.certf, opts.keyf, opts.rootCerts)
	if err != nil {
		return nil, err
	}
	
	// Create transport
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: 1,
		IdleConnTimeout:     opts.timeout,
		TLSClientConfig:     tlsc,
		DisableKeepAlives:   !opts.keepAlive,
	}
	
	if opts.h2c {
		transport.ForceAttemptHTTP2 = true
	}
	
	// Create client
	client := &http.Client{
		Transport: transport,
		Timeout:   opts.timeout,
	}
	
	if !opts.h2c && opts.http2 {
		// Enable HTTP/2
	}
	
	if opts.redirects >= 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= opts.redirects {
				return fmt.Errorf("stopped after %d redirects", opts.redirects)
			}
			return nil
		}
	}
	
	return client, nil
}

func executePunch(client *http.Client, target *vegeta.Target, opts *punchOpts) (*punchResult, error) {
	// Create request
	req, err := http.NewRequest(target.Method, target.URL, nil)
	if err != nil {
		return nil, err
	}
	
	// Set headers
	for key, values := range target.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	
	// Set body if present
	if target.Body != nil {
		req.Body = io.NopCloser(strings.NewReader(string(target.Body)))
	}
	
	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Read response body (limited)
	body, err := io.ReadAll(io.LimitReader(resp.Body, int64(opts.maxBody)))
	if err != nil {
		return nil, err
	}
	
	result := &punchResult{
		Status: fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
	}
	
	// Add details for verbose mode
	if opts.verbose {
		var details []string
		
		// Response headers
		details = append(details, "Response Headers:")
		for key, values := range resp.Header {
			details = append(details, fmt.Sprintf("  %s: %s", key, strings.Join(values, ", ")))
		}
		
		// Response body (truncated)
		if len(body) > 0 {
			bodyStr := string(body)
			if len(bodyStr) > 200 {
				bodyStr = bodyStr[:200] + "..."
			}
			// Try to format as JSON if possible
			if json.Valid([]byte(bodyStr)) {
				var prettyJSON interface{}
				if err := json.Unmarshal([]byte(bodyStr), &prettyJSON); err == nil {
					if formatted, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
						bodyStr = string(formatted)
						if len(bodyStr) > 300 {
							bodyStr = bodyStr[:300] + "..."
						}
					}
				}
			}
			details = append(details, "Response Body:")
			for _, line := range strings.Split(bodyStr, "\n") {
				details = append(details, "  "+line)
			}
		}
		
		result.Details = strings.Join(details, "\n")
	}
	
	// Check for error status codes
	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("HTTP %s", result.Status)
	}
	
	return result, nil
}
