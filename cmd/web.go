package cmd

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initWebCmd())
}

func initWebCmd() *cobra.Command {
	var args webArgs
	cmd := cobra.Command{
		Use:   "web",
		Short: "Start a web server for repo management with API proxy",
		Example: `  linglong-tools web -r https://repo.linglong.dev
  linglong-tools web -r https://repo.linglong.dev -a :9090`,
		Run: func(cmd *cobra.Command, _ []string) {
			runWeb(args)
		},
	}
	cmd.Flags().StringVarP(&args.RepoURL, "repo", "r", DefaultRepoUrl, "remote repo api url")
	cmd.Flags().StringVarP(&args.Addr, "addr", "a", ":8080", "listen address")
	return &cmd
}

type webArgs struct {
	RepoURL string
	Addr    string
}

func runWeb(args webArgs) {
	repoURL := strings.TrimRight(args.RepoURL, "/")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/", makeProxyHandler(repoURL))
	mux.HandleFunc("/", makeStaticHandler())

	srv := &http.Server{
		Addr:         args.Addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	log.Printf("linglong web server listening on %s, proxy to %s", args.Addr, repoURL)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func makeProxyHandler(repoURL string) http.HandlerFunc {
	client := &http.Client{
		Timeout: 300 * time.Second,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL := repoURL + r.URL.Path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			http.Error(w, "create proxy request failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		for key, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}
		proxyReq.Header.Del("Origin")
		proxyReq.Header.Del("Referer")

		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "proxy request failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

func makeStaticHandler() http.HandlerFunc {
	webDir := findWebDir()
	fs := http.FileServer(http.Dir(webDir))
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		if _, err := os.Stat(filepath.Join(webDir, path)); err != nil {
			r.URL.Path = "/"
		}
		fs.ServeHTTP(w, r)
	}
}

func findWebDir() string {
	candidates := []string{
		"web",
		filepath.Join("..", "web"),
	}
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates, filepath.Join(exeDir, "web"))
	}
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
			return dir
		}
	}
	return "web"
}
