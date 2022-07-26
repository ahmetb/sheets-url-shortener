package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	port, addr := os.Getenv("PORT"), os.Getenv("LISTEN_ADDR")
	if port == "" {
		port = "8080"
	}

	googleSheetsID := os.Getenv("GOOGLE_SHEET_ID")
	sheetName := os.Getenv("SHEET_NAME")
	homeRedirect := os.Getenv("HOME_REDIRECT")

	ttlVal := os.Getenv("CACHE_TTL")
	ttl := time.Second * 5
	if ttlVal != "" {
		v, err := time.ParseDuration(ttlVal)
		if err != nil {
			log.Fatalf("failed to parse CACHE_TTL as duration: %v", err)
		}
		ttl = v
	}

	srv := &server{
		db: &cachedURLMap{
			ttl: ttl,
			sheet: &sheetsProvider{
				googleSheetsID: googleSheetsID,
				sheetName:      sheetName,
			},
		},
		homeRedirect: homeRedirect,
	}

	http.HandleFunc("/", srv.handler)

	listenAddr := net.JoinHostPort(addr, port)
	log.Printf("starting server at %s; ttl=%v", listenAddr, ttl)
	err := http.ListenAndServe(listenAddr, nil)
	log.Fatal(err)
}

type server struct {
	db           *cachedURLMap
	homeRedirect string
}

type URLMap map[string]*url.URL

type cachedURLMap struct {
	sync.RWMutex
	v          URLMap
	lastUpdate time.Time

	ttl   time.Duration
	sheet *sheetsProvider
}

func (c *cachedURLMap) Get(query string) (*url.URL, error) {
	if err := c.refresh(); err != nil {
		return nil, err
	}
	c.RLock()
	defer c.RUnlock()
	return c.v[query], nil
}

func (c *cachedURLMap) refresh() error {
	c.Lock()
	defer c.Unlock()
	if time.Since(c.lastUpdate) <= c.ttl {
		return nil
	}

	rows, err := c.sheet.Query()
	if err != nil {
		return err
	}
	c.v = urlMap(rows)
	c.lastUpdate = time.Now()
	return nil
}

func (s *server) handler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		s.home(w, req)
		return
	}
	s.redirect(w, req)
}

func (s *server) home(w http.ResponseWriter, req *http.Request) {
	if s.homeRedirect != "" {
		http.Redirect(w, req, s.homeRedirect, http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `<!DOCTYPE html>
	<html><head><title>Not found</title></head><body><h1>Not found :(</h1>
	<p>This is home page for a URL redirector service.</p>
	<p>The URL is missing the shortcut in the path.</p>
	</body></html>`)
}

func (s *server) redirect(w http.ResponseWriter, req *http.Request) {

	if req.Body != nil {
		defer req.Body.Close()
	}

	redirTo, err := s.findRedirect(req.URL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to find redirect: %v", err)
		return
	}
	if redirTo == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 not found")
		return
	}

	log.Printf("redirecting=%q to=%q", req.URL, redirTo.String())
	http.Redirect(w, req, redirTo.String(), http.StatusMovedPermanently)
}

func (s *server) findRedirect(req *url.URL) (*url.URL, error) {
	path := strings.TrimPrefix(req.Path, "/")

	segments := strings.Split(path, "/")
	var discard []string
	for len(segments) > 0 {
		query := strings.Join(segments, "/")
		v, err := s.db.Get(query)
		if err != nil {
			return nil, err
		}
		if v != nil {
			return prepRedirect(v, strings.Join(discard, "/"), req.Query()), nil
		}
		discard = append([]string{segments[len(segments)-1]}, discard...)
		segments = segments[:len(segments)-1]
	}
	return nil, nil
}

func prepRedirect(base *url.URL, addPath string, query url.Values) *url.URL {
	if addPath != "" {
		if !strings.HasSuffix(base.Path, "/") {
			base.Path += "/"
		}
		base.Path += addPath
	}

	qs := base.Query()
	for k := range query {
		qs.Add(k, query.Get(k))
	}
	base.RawQuery = qs.Encode()
	return base
}

func urlMap(in [][]interface{}) URLMap {
	out := make(URLMap)
	for _, row := range in {
		if len(row) < 2 {
			continue
		}
		k, ok := row[0].(string)
		if !ok || k == "" {
			continue
		}
		v, ok := row[1].(string)
		if !ok || v == "" {
			continue
		}

		k = strings.ToLower(k)
		u, err := url.Parse(v)
		if err != nil {
			log.Printf("warn: %s=%s url invalid", k, v)
			continue
		}

		_, exists := out[k]
		if exists {
			log.Printf("warn: shortcut %q redeclared, overwriting", k)
		}
		out[k] = u
	}
	return out
}

func writeError(w http.ResponseWriter, code int, msg string, vals ...interface{}) {
	w.WriteHeader(code)
	fmt.Fprintf(w, msg, vals...)
}
