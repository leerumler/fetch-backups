package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type targetinfo struct {
	ip, user, pass, dir, port, proto, email string
}

type sourceinfo struct {
	ip, user, pass string
}

func prompt(question string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(question)
	input, _ := reader.ReadString('\n')
	answer := strings.TrimSuffix(input, "\n")
	return answer
}

func parsePrompt(flagname, usage, question string) string {
	var answer string
	flag.StringVar(&answer, flagname, "", usage)
	flag.Parse()
	if answer == "" {
		answer = prompt(question)
	}
	return answer
}

func checkPrivilege() string {
	access := parsePrompt("access", "access level to source (single|reseller)", "Level of access (single|reseller): ")
	if access != "single" && access != "reseller" {
		checkPrivilege()
	}
	return access
}

func getInfo() (targetinfo, sourceinfo) {

	// Collect target info.
	var target targetinfo
	target.ip = parsePrompt("tip", "target ip address", "Target IP: ")
	target.user = parsePrompt("tuser", "target username", "Target User: ")
	target.pass = parsePrompt("tpass", "target password", "Target Pass: ")
	target.dir = parsePrompt("tdir", "target directory", "Target Directory: ")
	target.port = parsePrompt("tport", "target port", "Target Port: ")
	target.proto = parsePrompt("proto", "target protocol", "Transport Protocol (scp|ftp): ")
	target.email = parsePrompt("email", "email address (for notifications)", "Email (for notifications): ")

	// Collect source info.
	var source sourceinfo
	source.ip = parsePrompt("sip", "source ip address", "Source IP: ")
	source.user = parsePrompt("suser", "source username", "Source User: ")
	source.pass = parsePrompt("spass", "source password", "Source Pass: ")

	return target, source
}

func genPOST(users []string, target targetinfo, source sourceinfo) {
	userNum := len(users)
	fmt.Println("Found", userNum, "remote users.")

	urlPieces := []string{"https://", target.ip, ":2083/json-api/cpanel"}
	cPurl := strings.Join(urlPieces, "")

	for i, user := range users {
		fmt.Printf("Sending request for user %v of %v: %v\n", i+1, userNum, user)
		fmt.Println("to", cPurl)

		data := url.Values{}
		data.Set("api.version", "1")
		data.Set("cpanel_jsonapi_user", user)
		data.Set("cpanel_jsonapi_module", "Fileman")
		data.Set("cpanel_jsonapi_func", "fullbackup")
		data.Set("cpanel_jsonapi_version", "1")
		data.Set("dest", target.proto)
		data.Set("server", target.ip)
		data.Set("user", target.user)
		data.Set("pass", target.pass)
		data.Set("email", target.email)
		data.Set("port", target.port)
		data.Set("rdir", target.dir)

		body := data.Encode()
		fmt.Println("Post Body:", body)
		trans := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: trans}
		request, _ := http.NewRequest("POST", cPurl, bytes.NewBufferString(body))
		request.Header.Add("Content Type:", "application/x-www-form-urlencoded")
		request.SetBasicAuth(user, source.pass)
		response, _ := client.Do(request)
		defer response.Body.Close()

		fmt.Println("Response:", response.Status)

	}
}

func main() {
	access := checkPrivilege()
	target, source := getInfo()

	switch access {
	case "single":
		user := []string{source.user}
		genPOST(user, target, source)
	case "reseller":
		// here's were we'll get the user list
		var users []string
		genPOST(users, target, source)
	}

	// dumpInfo(&target, &source)
}
