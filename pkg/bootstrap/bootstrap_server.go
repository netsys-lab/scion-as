package bootstrap

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

func RunBootstrapServer(configDir string, url string) error {
	// TODO: Configure logging
	// TODO: Source IP Validation
	//accessLog := log.New(os.Stdout, "ACCESS: ", log.LstdFlags)
	//errorLog := log.New(os.Stderr, "ERROR: ", log.LstdFlags)

	// Set up file server
	fileServer := http.FileServer(http.Dir(configDir))

	// Define routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Expires", "0")
		// accessLog.Printf("%s - %s \"%s\" %s \"%s\" \"%s\" \"%s\"",
		//	r.RemoteAddr, "-", r.Method+" "+r.URL.Path+" "+r.Proto, http.StatusOK, r.Referer(),
		//	r.UserAgent(), r.Header.Get("X-Forwarded-For"))

		if strings.HasPrefix(r.URL.Path, "/topology") {
			log.Println("[Bootstrap Server] Serving topology.json: ", filepath.Join(configDir, "topology.json"))
			http.ServeFile(w, r, filepath.Join(configDir, "topology.json"))
		} else if strings.HasPrefix(r.URL.Path, "/trcs") {
			if match, _ := regexp.MatchString(`^/trcs/isd\d+-b\d+-s\d+$`, r.URL.Path); match {
				isd, base, serial := parseISDBSerial(r.URL.Path)
				trc := fmt.Sprintf("isd%s-b%s-s%s.json", isd, base, serial)
				trcFile := filepath.Join(configDir, "certs", trc)
				log.Println("[Bootstrap Server] Serving TRC file: ", trcFile)
				http.ServeFile(w, r, trcFile)
			} else if match, _ := regexp.MatchString(`^/trcs/isd\d+-b\d+-s\d+/blob$`, r.URL.Path); match {
				isd, base, serial := parseISDBSerialBlob(r.URL.Path)
				trc := fmt.Sprintf("ISD%s-B%s-S%s.trc", isd, base, serial)
				trcFile := filepath.Join(configDir, "certs", trc)
				log.Println("[Bootstrap Server] Serving TRC file blob: ", trcFile)
				http.ServeFile(w, r, trcFile)
			} else {
				log.Println("[Bootstrap Server] Serving trcs.json: ", filepath.Join(configDir, "trcs.json"))
				http.ServeFile(w, r, filepath.Join(configDir, "trcs.json"))
			}
		} else {
			fileServer.ServeHTTP(w, r)
		}
	})

	listenUrl := ":8041"
	if url != "" {
		listenUrl = url
	}

	// Start server
	// errorLog.Println("Starting server on :8041")
	log.Println("[Bootstrap Server] Starting Bootstrap server on ", listenUrl)
	if err := http.ListenAndServe(listenUrl, nil); err != nil {
		return err
	}

	return nil
}

// parseISDBSerial extracts ISD, base, and serial from the URL
func parseISDBSerial(path string) (string, string, string) {
	re := regexp.MustCompile(`^/trcs/isd(\d+)-b(\d+)-s(\d+)$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) == 4 {
		return matches[1], matches[2], matches[3]
	}
	return "", "", ""
}

// parseISDBSerialBlob extracts ISD, base, and serial from the URL for blob
func parseISDBSerialBlob(path string) (string, string, string) {
	re := regexp.MustCompile(`^/trcs/isd(\d+)-b(\d+)-s(\d+)/blob$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) == 4 {
		return matches[1], matches[2], matches[3]
	}
	return "", "", ""
}
