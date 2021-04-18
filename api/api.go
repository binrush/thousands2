package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

type RuntimeConfig struct {
	datadir string
}

type App struct {
	Api   *Api
	UIDir string
}

func (h *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	head, path := ShiftPath(r.URL.Path)
	if head == "api" {
		r.URL.Path = path
		h.Api.ServeHTTP(w, r)
		return
	}
	fs := http.FileServer(http.Dir(h.UIDir))
	fs.ServeHTTP(w, r)
}

type Api struct {
	Config *RuntimeConfig
}

type Summits struct {
	Summits []Summit `json:"summits"`
}

func (h *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	if head == "summits" {
		if r.Method != "GET" {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/" {
			http.NotFound(w, r)
		}
		summits, err := LoadSummits(h.Config.datadir)
		if err != nil {
			log.Printf("Error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		resp, err := json.Marshal(Summits{summits})
		if err != nil {
			log.Printf("Error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(resp))
		return
	}
	http.NotFound(w, r)
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: api <datadir> <ui_dir>")
	}
	conf := &RuntimeConfig{datadir: path.Clean(os.Args[1])}
	app := &App{Api: &Api{Config: conf}, UIDir: os.Args[2]}
	log.Fatal(http.ListenAndServe(":5000", app))
}
