package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}
type Server interface {
	Address() string
	IsAlive() bool
	Server(w http.ResponseWriter, r *http.Request)
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)
	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type loadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func NewLoadBalancer(port string, servers []Server) *loadBalancer {
	return &loadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
func (s *simpleServer) Address() string {
	return s.addr
}
func (s *simpleServer) IsAlive() bool {
	return true
}
func (s *simpleServer) Serve(rw http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(rw, r)
}

func (lb *loadBalancer) getNextAvailableSever() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server
}
func (lb *loadBalancer) serveProxy(rw http.ResponseWriter, r *http.Request) {
	targetServer := lb.getNextAvailableSever()
	fmt.Printf("forwarding request to address %q\n", targetServer.Address())
	targetServer.Serve(rw, r)
}

func main() {
	servers := []Server{
		newSimpleServer("https://www.youtube.com"),
		newSimpleServer("https://www.twitter.com"),
		newSimpleServer("https://www.facebook.com"),
	}
	lb := NewLoadBalancer("8003", servers)
	handleRedirect := func(rw http.ResponseWriter, r *http.Request) {
		lb.serveProxy(rw, r)
	}
	http.HandleFunc("/", handleRedirect)
	fmt.Printf("server is serving at 'localhost:%s'\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
