package main

import (
	"bufio"
	"fmt"
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

func checkPrivelege() string {
	access := prompt("Level of access (single|reseller): ")
	if access != "single" && access != "reseller" {
		checkPrivelege()
	}
	return access
}

func prompts() (targetinfo, sourceinfo) {

	// Collect target info.
	var target targetinfo
	target.ip = prompt("Target IP: ")
	target.user = prompt("Target User: ")
	target.pass = prompt("Target Pass: ")
	target.dir = prompt("Target Directory: ")
	target.port = prompt("Target Port: ")
	target.proto = prompt("Transport Protocol (scp|ftp): ")
	target.email = prompt("Email (for notifications): ")

	// Collect source info.
	var source sourceinfo
	source.ip = prompt("Source IP: ")
	source.user = prompt("Source User: ")
	source.pass = prompt("Source Pass: ")

	return target, source
}

func genPOST(user *[]string, target *targetinfo, source *sourceinfo) {

}

func main() {
	access := checkPrivelege()
	target, source := prompts()

	switch access {
	case "single":
		user := []string{source.user}
		genPOST(&user, &target, &source)
	case "reseller":
		// here's were we'll get the user list
		var users []string
		genPOST(&users, &target, &source)
	}

	// dumpInfo(&target, &source)
}
