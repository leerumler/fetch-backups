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
	ip, user, pass, dir, port, email string
}

type sourceinfo struct {
	ip, user, pass string
}

type optinfo struct {
	access, email, proto string
}

func prompt(question string) *string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(question)
	input, _ := reader.ReadString('\n')
	answer := strings.TrimSuffix(input, "\n")
	return &answer
}

// func parsePrompt(flagname, usage, question string) string {
func reqFlags(flagMap map[string][2]string) map[string]string {

	//
	answers := make(map[string]string)

	//
	for flagname, helps := range flagMap {
		var response string
		flag.StringVar(&response, flagname, "", helps[0])
		answers[flagname] = response
	}
	flag.Parse()

	//
	for flagname, helps := range flagMap {
		if answers[flagname] == "" {
			answers[flagname] = *prompt(helps[1])
		}
	}
	return answers
}

func checkPrivilege() *string {

	// Read access level from Args.
	access := flag.Arg(0)

	// Validate responses.
	for access != "single" && access != "reseller" {
		access = *prompt("Level of access (single|reseller): ")
	}

	// Return access
	return &access
}

func getInfo() (*optinfo, *targetinfo, *sourceinfo) {

	// Create map of required flags.
	var flagMap map[string][2]string
	flagMap["sip"] = [2]string{"source ip address", "Source IP: "}
	flagMap["suser"] = [2]string{"source username", "Source User:"}
	flagMap["spass"] = [2]string{"source password", "Source Pass: "}
	flagMap["tip"] = [2]string{"target ip address", "Target IP: "}
	flagMap["tuser"] = [2]string{"target username", "Target User: "}
	flagMap["tpass"] = [2]string{"target password", "Target Pass: "}
	flagMap["tdir"] = [2]string{"target directory", "Target Directory: "}
	flagMap["tport"] = [2]string{"target port", "Target Port: "}
	flagMap["proto"] = [2]string{"target protocol", "Transport Protocol (scp|ftp): "}
	flagMap["email"] = [2]string{"email address (for notifications)", "Email (for notifications): "}

	flagVals := reqFlags(flagMap)

	// Set opts.
	var opts optinfo
	opts.access = *checkPrivilege()
	opts.proto = flagVals["proto"]
	opts.email = flagVals["email"]

	// Set target info.
	var target targetinfo
	target.ip = flagVals["tip"]
	target.user = flagVals["tuser"]
	target.pass = flagVals["tpass"]
	target.dir = flagVals["tdir"]
	target.port = flagVals["tport"]

	// Collect source info.
	var source sourceinfo
	source.ip = flagVals["sip"]
	source.user = flagVals["suser"]
	source.pass = flagVals["spass"]

	return &opts, &target, &source
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

func reqBackups(users []string, target *targetinfo, source *sourceinfo, opts *optinfo) {
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
		data.Set("arg-0", opts.proto)
		data.Set("arg-1", target.ip)
		data.Set("arg-2", target.user)
		data.Set("arg-3", target.pass)
		data.Set("arg-4", opts.email)
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

	opts, target, source := getInfo()

	switch opts.access {
	case "single":
		user := []string{source.user}
		reqBackups(user, target, source, opts)
	case "reseller":
		// here's were we'll get the user list
		var users []string
		reqBackups(users, target, source, opts)
	}

	// dumpInfo(&target, &source)
}
