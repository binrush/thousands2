package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
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
	Datadir      string
	AuthConfig   AuthProviders
	ItemsPerPage int
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
	DB     *Database
}

type Summits struct {
	Summits []Summit `json:"summits"`
}

func (h *Api) HandleSummits(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	summits, err := LoadSummits(h.Config.Datadir)
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
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, string(resp))
}

func (h *Api) HandleAuth(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	provider := h.Config.AuthConfig[head]
	if provider != nil {
		http.Redirect(
			w, r, provider.AuthCodeURL("FIXME"), http.StatusTemporaryRedirect)
	}
}

func (h *Api) HandleTop(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	var page int
	var err error
	pageParam := r.URL.Query()["page"]
	if pageParam == nil {
		page = 1
	} else if len(pageParam) != 1 {
		http.Error(w, "Invalid page parameger provided", http.StatusBadRequest)
		return
	} else {
		page, err = strconv.Atoi(pageParam[0])
		if (err != nil) || (page <= 0) {
			http.Error(w, "Invalid page parameter provided", http.StatusBadRequest)
			return
		}
	}
	top, err := FetchTop(h.DB, page, h.Config.ItemsPerPage)
	if err != nil {
		log.Printf("Error fetching top: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(top)
	if err != nil {
		log.Printf("Error marshalling top: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, string(resp))
}

func (h *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)
	switch head {
	case "summits":
		h.HandleSummits(w, r)
	case "auth":
		h.HandleAuth(w, r)
	case "top":
		h.HandleTop(w, r)
	default:
		http.NotFound(w, r)
	}
}

func main() {
	if len(os.Args) != 4 {
		log.Fatal("Usage: api <datadir> <ui_dir> <db_path>")
	}
	conf := &RuntimeConfig{
		Datadir:      path.Clean(os.Args[1]),
		AuthConfig:   GetAuthProviders(),
		ItemsPerPage: 20,
	}
	db, err := NewDatabase(path.Clean(os.Args[3]))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting migrations...")
	err = db.Migrate()
	if err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}
	log.Printf("Migrations completed")
	app := &App{Api: &Api{Config: conf, DB: db}, UIDir: os.Args[2]}
	log.Fatal(http.ListenAndServe(":5000", app))
}
