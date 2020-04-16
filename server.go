package main

import (
	"crypto/tls"
	"encoding/json"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func httpServer() {
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
	}
	server := &http.Server{
		Addr:    ":http",
		Handler: logRequest(certManager.HTTPHandler(nil)),

	}
	server.ListenAndServe()
}

func httpsServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", mainProcess)
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
	}
	dir := cacheDir()
	if dir != "" {
		certManager.Cache = autocert.DirCache(dir)
	}
	server := http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
		Handler: logRequest(mux),
	}
	server.TLSConfig.NextProtos = append(server.TLSConfig.NextProtos, acme.ALPNProto)
	server.ListenAndServeTLS("", "")
}

func mainProcess(w http.ResponseWriter, req *http.Request) {
	config := loadConfig(req.Host, req.RequestURI)
	if config.Upstream != "" {
		urlParsed, _ := url.Parse(config.Upstream)
		proxy := httputil.NewSingleHostReverseProxy(urlParsed)
		req.URL.Host = urlParsed.Host
		req.URL.Scheme = urlParsed.Scheme
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Host = urlParsed.Host
		proxy.ServeHTTP(w, req)
	} else {
		w.Header().Set("Strict-Transport-Security", "max-age=15768000 ; includeSubDomains")
		for _, value := range config.Headers {
			w.Header().Set(value.Name, value.Value)
		}
		w.WriteHeader(int(config.Status))
		w.Write([]byte(config.Body))
	}
}

func cacheDir() string {
	dir := CertsDir
	if err := os.MkdirAll(dir, 0700); err == nil {
		return dir
	}
	return ""
}

func loadConfig(host string, uri string) RedirectResponse {
	dir := ConfigDir

	defaultResponse := RedirectResponse{uri, 200, nil, "Redirect system", ""}
	notFoundResponse := RedirectResponse{uri, 404, nil, "", ""}

	filename := uri
	if uri == "/" {
		filename = "/index.html"
	}
	f, err := os.Open(dir + host + "/static" + filename)
	if err == nil {
		fstat, _ := f.Stat()
		if !fstat.IsDir() {
			body, err := ioutil.ReadFile(dir + host + "/static" + filename)
			if err != nil {
				return notFoundResponse
			}
			ct := http.DetectContentType(body)
			if strings.HasPrefix(ct, "text/plain") {
				if strings.HasSuffix(strings.ToLower(filename), ".css") {
					ct = strings.Replace(ct, "text/plain", "text/css", 1)
				}
				if strings.HasSuffix(strings.ToLower(filename), ".js") {
					ct = strings.Replace(ct, "text/plain", "application/javascript", 1)
				}
				if strings.HasSuffix(strings.ToLower(filename), ".html") || strings.HasSuffix(strings.ToLower(filename), ".htm") {
					ct = strings.Replace(ct, "text/plain", "text/html", 1)
				}
			}
			return RedirectResponse{
				uri,
				200,
				[]RedirectHeader{
					{"Content-Type", ct},
					{"Cache-Control", "max-age=31536000"},
				},
				string(body),
				"",
			}
		}
	}

	file, err := ioutil.ReadFile(dir + host + "/config.json")
	if err != nil {
		file, err = ioutil.ReadFile(dir + "default" + "/config.json")
		if err != nil {
			return defaultResponse
		}
	}
	var redirects []RedirectResponse
	err = json.Unmarshal(file, &redirects)
	if err != nil {
		return defaultResponse
	}

	index1 := -1
	index2 := -1
	for index, value := range redirects {
		if value.Uri == uri {
			index1 = index
			break
		}
		if value.Uri == "*" {
			index2 = index
		}
	}

	if index1 == -1 {
		if index2 == -1 {
			return defaultResponse
		} else {
			return redirects[index2]
		}
	} else {
		return redirects[index1]
	}
}