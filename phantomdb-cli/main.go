package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	host := flag.String("host", "https://localhost:8080", "PhantomDB server address")
	user := flag.String("user", "", "username")
	pass := flag.String("pass", "", "password")
	insecure := flag.Bool("insecure", false, "skip TLS certificate verification (for self-signed certs)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: phantomdb-cli --user <u> --pass <p> [--host <url>] [--insecure] <command> [args]")
		fmt.Println("Commands: get <key> | put <key> <value> | scan <start> <end> | delete <key>")
		os.Exit(1)
	}

	client := &http.Client{}
	if *insecure {
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}

	cmd := args[0]

	switch cmd {
	case "get":
		if len(args) < 2 {
			fmt.Println("usage: get <key>")
			os.Exit(1)
		}
		body := doRequest(client, "GET", *host+"/api/get?key="+args[1], *user, *pass, nil)
		fmt.Println(body)

	case "put":
		if len(args) < 3 {
			fmt.Println("usage: put <key> <value>")
			os.Exit(1)
		}
		payload, _ := json.Marshal(map[string]string{"Key": args[1], "Value": args[2]})
		body := doRequest(client, "POST", *host+"/api/put", *user, *pass, strings.NewReader(string(payload)))
		if body == "" {
			fmt.Println("OK")
		} else {
			fmt.Println(body)
		}

	case "scan":
		if len(args) < 3 {
			fmt.Println("usage: scan <start> <end>")
			os.Exit(1)
		}
		body := doRequest(client, "GET", *host+"/api/scan?start="+args[1]+"&end="+args[2], *user, *pass, nil)
		fmt.Println(body)

	case "delete":
		if len(args) < 2 {
			fmt.Println("usage: delete <key>")
			os.Exit(1)
		}
		body := doRequest(client, "DELETE", *host+"/api/delete?key="+args[1], *user, *pass, nil)
		if body == "" {
			fmt.Println("OK")
		} else {
			fmt.Println(body)
		}

	default:
		fmt.Println("unknown command:", cmd)
		os.Exit(1)
	}
}

func doRequest(client *http.Client, method, url, user, pass string, body io.Reader) string {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		fmt.Printf("error (%d): %s\n", resp.StatusCode, strings.TrimSpace(string(data)))
		os.Exit(1)
	}

	return strings.TrimSpace(string(data))
}