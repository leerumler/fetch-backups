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
	ip, user, pass, dir, port *string
}

type sourceinfo struct {
	ip, user, pass *string
}

type optinfo struct {
	access, email, proto *string
}

func prompt(question string) *string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(question)
	input, _ := reader.ReadString('\n')
	answer := strings.TrimSuffix(input, "\n")
	return &answer
}

func validate(answer *string, question string) *string {
	for *answer == "" {
		answer = prompt(question)
	}
	return answer
}

func checkPrivilege(access *string) *string {

	// Validate responses.
	for *access != "single" && *access != "reseller" {
		access = prompt("Level of access (single|reseller): ")
	}

	// Return access
	return access
}

func getInfo() (*optinfo, *targetinfo, *sourceinfo) {

	var opts optinfo
	var target targetinfo
	var source sourceinfo

	opts.access = flag.String("access", "", "level of access (single|reseller)")
	source.ip = flag.String("sip", "", "source ip address")
	source.user = flag.String("suser", "", "source username")
	source.pass = flag.String("spass", "", "source password")
	target.ip = flag.String("tip", "", "target ip address")
	target.user = flag.String("tuser", "", "target username")
	target.pass = flag.String("tpass", "", "target password")
	target.dir = flag.String("tdir", "", "target directory")
	target.port = flag.String("tport", "", "target port")
	opts.email = flag.String("email", "", "email address (for notifications)")
	opts.proto = flag.String("proto", "", "transport protocol")

	flag.Parse()

	opts.access = checkPrivilege(opts.access)
	source.ip = validate(source.ip, "Source IP: ")
	source.user = validate(source.user, "Source User: ")
	source.pass = validate(source.pass, "Source Pass: ")
	target.ip = validate(target.ip, "Target IP: ")
	target.user = validate(target.user, "Target User: ")
	target.pass = validate(target.pass, "Target Pass: ")
	target.port = validate(target.port, "Target Port: ")
	target.dir = validate(target.dir, "Target Directory: ")
	opts.proto = validate(opts.proto, "Transport Protocol: ")
	opts.email = validate(opts.email, "Email (for notifications): ")

	return &opts, &target, &source
}

func getAddr(source *sourceinfo) string {
	urlPieces := []string{"https://", *source.ip, ":2083/json-api/cpanel"}
	return strings.Join(urlPieces, "")
}

func getUsers(source *sourceinfo) []string {
	cPurl := getAddr(source)
	data := url.Values{}
	data.Add("cpanel_jsonapi_user", *source.user)
	data.Add("cpanel_jsonapi_apiversion", "2")
	data.Add("cpanel_jsonapi_module", "resellers")
	data.Add("cpanel_jsonapi_func", "get_sub_accounts")
	body := data.Encode()

	response := sendPOST(cPurl, body, *source.user, *source.pass)

	cPjsonbuff := new(bytes.Buffer)
	if _, err := io.Copy(cPjsonbuff, response.Body); err != nil {
		log.Fatal(err)
	}

	// cPjsonbytes := make([]byte, cPjsonbuff.Len())

	// 	cPjson
	//	{
	//		"cpanelresult":{
	//			"apiversion":2,
	// 			"func":"resellers",
	//			"data":[
	//				{
	//					"domain":"example.com",
	//					"user":"example",
	//					"select":"1"
	//				},
	//				{
	//					"domain":"example1.com",
	//					"user":"example1",
	//					"select":""
	//				}
	//			],
	//			"event":{
	//				"result":1
	//			},
	//			"module":"Reseller"
	//		}
	//	}

	return make([]string, 0)
}

func sendPOST(cPurl, body, user, pass string) *http.Response {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

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

	return response
}

func fetchBackups(users []string, target *targetinfo, source *sourceinfo, opts *optinfo) {
	userNum := len(users)
	fmt.Println("Found", userNum, "remote users.")

	cPurl := getAddr(source)

	for i, user := range users {
		data := url.Values{}
		data.Set("api.version", "1")
		data.Set("cpanel_jsonapi_user", user)
		data.Set("cpanel_jsonapi_module", "Fileman")
		data.Set("cpanel_jsonapi_func", "fullbackup")
		data.Set("cpanel_jsonapi_apiversion", "1")
		data.Set("arg-0", *opts.proto)
		data.Set("arg-1", *target.ip)
		data.Set("arg-2", *target.user)
		data.Set("arg-3", *target.pass)
		data.Set("arg-4", *opts.email)
		data.Set("arg-5", *target.port)
		data.Set("arg-6", *target.dir)
		body := data.Encode()

		// debugging info
		fmt.Printf("Sending request for user %v of %v: %v\n", i+1, userNum, user)
		fmt.Println("to", cPurl)
		fmt.Println("Post Body:", body)

		//
		response := sendPOST(cPurl, body, user, *source.pass)

		// Display response status and body.
		fmt.Println("Response Status:", response.Status)
		// if _, err := io.Copy(os.Stdout, response.Body); err != nil {
		// 	log.Fatal(err)
		// }

		if i+1 < userNum {
			fmt.Println("Sleeping for 60 seconds to decrease server load.")
			time.Sleep(60 * time.Second)
		}

	}
}

func main() {

	opts, target, source := getInfo()

	switch *opts.access {
	case "single":
		user := []string{*source.user}
		fetchBackups(user, target, source, opts)
	case "reseller":
		// here's were we'll get the user list
		users := getUsers(source)
		fetchBackups(users, target, source, opts)
	}
}
