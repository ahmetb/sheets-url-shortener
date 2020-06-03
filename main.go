package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"google.golang.org/api/sheets/v4"
)

func main() {
	log.Println("--------------------")
	port, addr := os.Getenv("PORT"), os.Getenv("LISTEN_ADDR")
	if port == "" {
		port = "8080"
	}
	if addr == "" {
		addr = "localhost"

	}

	googleSheetsID := os.Getenv("GOOGLE_SHEET_ID")
	sheetName := os.Getenv("SHEET_NAME")

	srv := &server{
		googleSheetsID: googleSheetsID,
		sheetName:      sheetName,
	}

	http.HandleFunc("/", srv.redirect)

	listenAddr := net.JoinHostPort(addr, port)
	log.Printf("starting server at %s", listenAddr)
	err := http.ListenAndServe(listenAddr, nil)
	log.Fatal(err)
}

type server struct {
	googleSheetsID string
	sheetName      string
}

func (s *server) redirect(w http.ResponseWriter, req *http.Request) {
	if s.googleSheetsID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "GOOGLE_SHEET_ID not set")
		return
	}

	ctx := req.Context()
	srv, err := sheets.NewService(ctx)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	log.Println("querying sheet")
	readRange := "A:B"
	if s.sheetName != "" {
		readRange = s.sheetName + "!" + readRange
	}
	resp, err := srv.Spreadsheets.Values.Get(s.googleSheetsID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	shortcuts := urlMap(resp.Values)
	log.Printf("parsed %d shortcuts", len(shortcuts))

	redirTo := findRedirect(shortcuts, req.URL)
	if redirTo == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "shortcut not found")
		return
	}

	log.Printf("redirecting=%q to=%q", req.URL, redirTo.String())
	http.Redirect(w, req, redirTo.String(), http.StatusMovedPermanently)
}

func findRedirect(m map[string]*url.URL, req *url.URL) *url.URL {
	path := strings.TrimPrefix(req.Path, "/")

	segments := strings.Split(path, "/")
	var discard []string
	for len(segments) > 0 {
		query := strings.Join(segments, "/")
		v, ok := m[query]
		if ok {
			return prepRedirect(v, strings.Join(discard, "/"), req.Query())
		}
		discard = append([]string{segments[len(segments)-1]}, discard...)
		segments = segments[:len(segments)-1]
	}
	return nil
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

func urlMap(in [][]interface{}) map[string]*url.URL {
	out := make(map[string]*url.URL)
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
