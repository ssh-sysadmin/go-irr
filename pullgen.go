package main

import (
	"net/http"
	"strings"
	"sync"
)

type tomlConfig struct {
	Routers map[string]Router `toml:"routers"`
	Peers   PeersSection      `toml:"peers"`
}

type Router struct {
	SourceIP string   `toml:"source-ip"`
	Hostname string   `toml:"hostname"`
	Peers    []string `toml:"peers"`
	Vendor   string   `toml:"vendor"`
}

type PeersSection struct {
	ASSet map[string]string `toml:"as-sets"`
}

func pgPeers(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	path := strings.Split(r.URL.Path, "/")
	rtr := path[len(path)-1]
	peers := pullGenConfig.Routers[rtr].Peers
	var asSets []string
	for _, peer := range peers {
		asSets = append(asSets, pullGenConfig.Peers.ASSet[peer])
	}
	out := strings.Join(asSets, "\n")
	w.Write([]byte(out))

}

//func pgGenRouterConfig(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
//	path := strings.Split(r.URL.Path, "/")
//	rtr := pullGenConfig.Routers[path[len(path)-1]]
//	peers := rtr.Peers
//	var response string
//	for _, peer := range peers {
//		asSet := pullGenConfig.Peers.ASSet[peer]
//		v4raw := queryBgpq4(rtr.Vendor, "v4", asSet)
//		v6raw := queryBgpq4(rtr.Vendor, "v6", asSet)
//		v4named := strings.Replace(v4raw, "NN", "as"+peer+"-import-ipv4", 1)
//		v6named := strings.Replace(v6raw, "NN", "as"+peer+"-import-ipv6", 1)
//		response = strings.Join([]string{response, v4named, v6named}, "\n")
//	}
//	w.Write([]byte(response))
//}

func pgGenRouterConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	path := strings.Split(r.URL.Path, "/")
	rtr := pullGenConfig.Routers[path[len(path)-1]]
	peers := rtr.Peers

	type result struct {
		content string
		err     error
	}

	results := make(chan result, len(peers))
	var wg sync.WaitGroup

	for _, peer := range peers {
		wg.Add(1)
		go func(peer string) {
			defer wg.Done()

			asSet := pullGenConfig.Peers.ASSet[peer]

			// Do both IPv4 and IPv6 in parallel, too (optional optimization)
			var innerWg sync.WaitGroup
			innerWg.Add(2)

			var v4raw, v6raw string

			go func() {
				defer innerWg.Done()
				v4raw = queryBgpq4(rtr.Vendor, "v4", asSet)
			}()

			go func() {
				defer innerWg.Done()
				v6raw = queryBgpq4(rtr.Vendor, "v6", asSet)
			}()

			innerWg.Wait()

			v4named := strings.Replace(v4raw, "NN", "as"+peer+"-import-ipv4", 1)
			v6named := strings.Replace(v6raw, "NN", "as"+peer+"-import-ipv6", 1)

			results <- result{content: strings.Join([]string{v4named, v6named}, "\n")}
		}(peer)
	}

	// Close results channel once all goroutines finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var responseBuilder strings.Builder
	for res := range results {
		if res.err != nil {
			continue // optionally handle/log error
		}
		responseBuilder.WriteString(res.content)
		responseBuilder.WriteString("\n")
	}

	w.Write([]byte(responseBuilder.String()))
}
