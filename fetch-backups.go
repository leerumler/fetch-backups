package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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

	if answer == "" {
		answer = prompt(question)
	}
	return answer
}

func checkPrivilege() string {
	access := parsePrompt("access", "access level to source (single|reseller)", "Level of access (single|reseller): ")
	for access != "single" && access != "reseller" {
		access = prompt("Acceptible values are single or reseller: ")
	}
	return access
}

func getInfo() (string, targetinfo, sourceinfo) {

	access := checkPrivilege()

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

	flag.Parse()

	return access, target, source
}

// func getUsers(source sourceinfo) []string {
//
// }

func sendPOST(cPurl, body, user, pass string) {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	// debugging info
	// fmt.Printf("Sending request for user %v of %v: %v\n", i+1, userNum, user)
	// fmt.Println("to", cPurl)
	// fmt.Println("Post Body:", body)

	request, err := http.NewRequest("POST", cPurl, bytes.NewBufferString(body))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Add("Content Type:", "application/x-www-form-urlencoded")
	request.SetBasicAuth(user, pass)
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
	}

	fmt.Println("Response Status:", response.Status)
	if _, err := io.Copy(os.Stdout, response.Body); err != nil {
		log.Fatal(err)
	}
}

func reqBackups(users []string, target targetinfo, source sourceinfo) {
	userNum := len(users)
	fmt.Println("Found", userNum, "remote users.")

	urlPieces := []string{"https://", target.ip, ":2083/json-api/cpanel"}
	cPurl := strings.Join(urlPieces, "")

	for i, user := range users {
		data := url.Values{}
		data.Set("api.version", "1")
		data.Set("cpanel_jsonapi_user", user)
		data.Set("cpanel_jsonapi_module", "Fileman")
		data.Set("cpanel_jsonapi_func", "fullbackup")
		data.Set("cpanel_jsonapi_apiversion", "1")
		data.Set("arg-0", target.proto)
		data.Set("arg-1", target.ip)
		data.Set("arg-2", target.user)
		data.Set("arg-3", target.pass)
		data.Set("arg-4", target.email)
		data.Set("arg-5", target.port)
		data.Set("arg-6", target.dir)
		body := data.Encode()

		sendPOST(cPurl, body, user, source.pass)

		if i+1 < userNum {
			fmt.Println("Sleeping for 60 seconds to decrease server load.")
			time.Sleep(60 * time.Second)
		}
	}
}

func main() {

	access, target, source := getInfo()

	switch access {
	case "single":
		user := []string{source.user}
		reqBackups(user, target, source)
	case "reseller":
		// here's were we'll get the user list
		var users []string
		reqBackups(users, target, source)
	}

	// dumpInfo(&target, &source)
}
