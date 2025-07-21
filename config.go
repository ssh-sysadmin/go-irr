package main

import (
	"os"
	"regexp"
	"strings"
)

type config struct {
	sources     []string
	matchParent bool
	listen      string
}

func loadConfig(cfg *config) {
	cfg.sources = parseEnv("SOURCES", []string{"AFRINIC", "APNIC", "ARIN", "LACNIC", "RIPE"}, func(s string) []string {
		return strings.Split(strings.ReplaceAll(s, ", ", ","), ",")
	})

	cfg.matchParent = parseEnv("MATCH_PARENT", true, func(s string) bool {
		matched, _ := regexp.MatchString("true|1|y(es)?", s)
		return matched
	})

	cfg.listen = fetchEnv("LISTEN", "[::]:8080")
}

type envParser[T any] func(string) T

func parseEnv[T any](key string, defaultValue T, parse envParser[T]) T {
	value, found := os.LookupEnv(key)
	if found {
		return parse(value)
	} else {
		return defaultValue
	}
}

func fetchEnv(key string, defaultValue string) string {
	return parseEnv(key, defaultValue, func(s string) string { return s })
}
