package main

import (
	"log"
	"net/http"
	"regexp"
	"strings"
)

var cache prefixCache
var conf config

func init() {
	cache.init()
	loadConfig(&conf)
	go cache.purgeEvery(conf.cacheTime)
}

func main() {
	http.HandleFunc("/", handle)

	http.HandleFunc("/health", handleHealthcheck)

	// Wrap the default mux with logging middleware so all requests are logged
	handler := loggingMiddleware(http.DefaultServeMux)
	log.Fatal(http.ListenAndServe(conf.listen, handler))

}

func handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	path := strings.Split(r.URL.Path, "/")
	q := r.URL.Query()

	// path has 3 segments:
	// /vendor/addressfamily/AS1234:AS-SET
	if len(path) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if cache bypass is allowed
	if !conf.allowCacheBypass && q.Get("bypassCache") != "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vendor := path[1]
	addrFamily := path[2]
	asnOrAsSet := strings.ReplaceAll(strings.ToUpper(path[3]), "_", ":")

	if vendor == "" || addrFamily == "" || !strings.HasPrefix((asnOrAsSet), "AS") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Make sure asnOrAsSet is correct format
	// AS\d{1,5} or AS-SET

	isASN, _ := regexp.MatchString("^AS\\d{1,6}$", asnOrAsSet)
	isAsSet, _ := regexp.MatchString("^AS[A-Z0-9:-]{1,48}$", asnOrAsSet)

	if !isASN && !isAsSet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the prefix list is already cached
	// If it isn't, look it up using bgpq4 and cache the result
	// Optional cache bypass
	output := cache.get(vendor, addrFamily, asnOrAsSet)

	if output == "" || q.Get("bypassCache") == "1" || q.Get("bypassCache") == "true" {
		output = queryBgpq4(vendor, addrFamily, asnOrAsSet)

		if output == "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}

		cache.set(vendor, addrFamily, asnOrAsSet, output)
	}

	if q.Has("name") {
		output = strings.ReplaceAll(output, "NN", q.Get("name"))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}
