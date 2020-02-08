package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dineshgowda24/lb/backendserver/config"
)

type AppServer struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (a *AppServer) IsAlive() bool {
	a.mux.RLock()
	val := a.Alive
	a.mux.RUnlock()
	return val
}

func (a *AppServer) SetAlive(isAlive bool) {
	a.mux.Lock()
	a.Alive = isAlive
	a.mux.Unlock()
}

type AppCluster struct {
	servers []*AppServer
	current uint64
}

func (c *AppCluster) AddAppServer(appServer *AppServer) {
	c.servers = append(c.servers, appServer)
}

func (c *AppCluster) NextServerIndex() uint64 {
	return uint64(atomic.AddUint64(&c.current, uint64(1)) % uint64(len(c.servers)))
}

func (c *AppCluster) GetNextServer() *AppServer {
	next := c.NextServerIndex()
	l := uint64(len(c.servers)) + next
	for i := next; i < l; i++ {
		//need to do mod
		if c.servers[i%uint64(len(c.servers))].IsAlive() {
			if i != next {
				atomic.StoreUint64(&c.current, i)
			}
			return c.servers[i]
		}
	}
	return nil
}

func isAppServerAlive(u *url.URL) bool {
	tout := 2 * time.Second
	_, err := http.Get(u.String() + "/")

	if err != nil {
		log.Printf("Server Unreachable, timeout after  %s seconds. Error : %s ", tout.String(), err)
		return false
	} else {
		log.Printf("Server reachable %s", u.String())
	}
	return true
}

func (c *AppCluster) HealthCheck() {
	for _, server := range c.servers {
		alive := isAppServerAlive(server.URL)
		server.SetAlive(alive)
		log.Println("Server ", server.URL, " Status ", alive)
	}
}

func (c *AppCluster) MarkBackendServerStatus(u url.URL, flag bool) {
	for _, s := range c.servers {
		if s.URL.String() == u.String() {
			s.SetAlive(flag)
			break
		}
	}
}

const (
	Attempts int = iota
	Retry
)

func getAttemptsFromRequest(r *http.Request) int {
	if a, ok := r.Context().Value(Attempts).(int); ok {
		return a
	}
	return 1
}

func getRetryFromRequest(r *http.Request) int {
	if r, ok := r.Context().Value(Retry).(int); ok {
		return r
	}
	return 0
}

func peroidicHealthCheck() {
	t := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-t.C:
			log.Println("Start Cluster Health Check")
			//Need to call HealthCheck of Cluster
			cluster.HealthCheck()
			log.Println("Stopping Cluster Health Check")
		}
	}
}

func loadbalance(w http.ResponseWriter, r *http.Request) {
	attemps := getAttemptsFromRequest(r)
	if attemps > 3 {
		//we have reached the max attempts so dont server Request
		log.Fatal("Max attempts reached for request, teminating. ", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	bserver := cluster.GetNextServer()
	if bserver != nil {
		bserver.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service Unavailabe", http.StatusServiceUnavailable)
}

var cluster AppCluster

func main() {
	c := flag.String("c", "backendserver/config/config.json", "Please specify conf.json")
	flag.Parse()
	file, err := os.Open(*c)
	if err != nil {
		log.Fatal("Unable to open config file")
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := bconfig.BackendConfiguration{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Unable to decode conf.json file")
	}

	for _, bserver := range config.Servers {
		serverURL := url.URL{
			Scheme: "http",
			Host:   bserver.Host + ":" + bserver.Port,
		}

		proxy := httputil.NewSingleHostReverseProxy(&serverURL)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, er error) {
			log.Printf("[%s] %s\n", serverURL.Host, er.Error())
			retries := getRetryFromRequest(r)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(r.Context(), Retry, retries+1)
					proxy.ServeHTTP(w, r.WithContext(ctx))
				}
				return
			}

			// after 3 retries, mark this backend as down
			cluster.MarkBackendServerStatus(serverURL, false)

			// if the same request routing for few attempts with different backends, increase the count
			attempts := getAttemptsFromRequest(r)
			log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
			ctx := context.WithValue(r.Context(), Attempts, attempts+1)
			loadbalance(w, r.WithContext(ctx))
		}
		cluster.AddAppServer(&AppServer{
			URL:          &serverURL,
			Alive:        true,
			ReverseProxy: proxy,
		})

		log.Printf("Configured backend server with host : %s and port : %s", bserver.Host, bserver.Port)
	}

	go peroidicHealthCheck()

	http.HandleFunc("/", loadbalance)
	if err := http.ListenAndServe("localhost:8800", nil); err != nil {
		log.Fatal("Unable to start loadbalancer")
	}
}
