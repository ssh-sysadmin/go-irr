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

func vendorShorthand(s string) string {
	return vendorShorthands[strings.ToLower(s)]
}

var addrFamilyShorthands = map[string]string{
	"v4": "4",
	"v6": "6",
}

func addrFamilyShorthand(s string) string {
	return addrFamilyShorthands[strings.ToLower(s)]
}

func queryBgpq4(addrFamily string, routerOs string, asnOrAsSet string, isEosAsSet bool) string {
	if isEosAsSet {
		asnOrAsSet = strings.ReplaceAll(asnOrAsSet, "_", ":")
	}

	var args []string

	args = append(args, "-S"+strings.Join(conf.sources, ","), "-"+addrFamily, "-"+routerOs)

	if routerOs == "J" {
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

	return stdout.String()
}
