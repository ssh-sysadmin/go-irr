package main

import (
	"bytes"
	"os/exec"
	"strings"
)

var vendorShorthands = map[string]string{
	"arista":    "e",
	"eos":       "e",
	"juniper":   "J",
	"bird":      "b",
	"routeros6": "K",
	"routeros7": "K7",
}

var addrFamilyShorthands = map[string]string{
	"v4": "4",
	"v6": "6",
}

func queryBgpq4(vendorName string, addrFamily string, asnOrAsSet string) string {
	var args []string

	vendor := vendorShorthands[strings.ToLower(vendorName)]
	addrFamily = addrFamilyShorthands[strings.ToLower(addrFamily)]

	args = append(args, "-S"+strings.Join(conf.sources, ","), "-"+addrFamily, "-"+vendor)

	if vendor == "J" {
		args = append(args, "-3")
	} else {
		args = append(args, "-A")
	}

	maxLen := "24"
	if addrFamily == "6" {
		maxLen = "48"
	}
	args = append(args, "-m "+maxLen)
	if conf.matchParent {
		args = append(args, "-R "+maxLen)
	}

	args = append(args, asnOrAsSet)
	cmd := exec.Command("bgpq4", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return ""
	}

	output := stdout.String()

	if vendorName == "eos" {
		output = stripHeadersForEos(output)
	}

	return output
}

func stripHeadersForEos(prefixList string) string {
	output := skipLines(prefixList, 2)

	if strings.Contains(output, "deny") {
		output = skipLines(prefixList, 1)
	}

	return output
}

func skipLines(s string, n int) string {
	return strings.SplitN(s, "\n", n+1)[n]
}
