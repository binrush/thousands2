package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/alexedwards/scs/v2"
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
	ItemsPerPage int
}

type App struct {
	Api        *Api
	AuthServer *AuthServer
	UIDir      string
	SM         *scs.SessionManager
}

func (h *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	head, path := ShiftPath(r.URL.Path)
	switch head {
	case "api":
		r.URL.Path = path
		h.Api.ServeHTTP(w, r)
	case "auth":
		r.URL.Path = path
		h.AuthServer.ServeHTTP(w, r)
	default:
		fs := http.FileServer(http.Dir(h.UIDir))
		fs.ServeHTTP(w, r)
	}
}

func main() {
	if len(os.Args) != 4 {
		log.Fatal("Usage: api <datadir> <ui_dir> <db_path>")
	}
	conf := &RuntimeConfig{
		Datadir:      path.Clean(os.Args[1]),
		ItemsPerPage: 20,
	}
	db, err := NewDatabase(path.Clean(os.Args[3]))
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	log.Printf("Starting migrations...")
	err = db.Migrate()
	if err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}
	log.Printf("Migrations completed")

	log.Printf("Loading summits data to database...")
	err = LoadSummits(conf.Datadir, db)
	if err != nil {
		log.Fatalf("Failed to load summits: %v", err)
	}
	log.Printf("Summits data loaded")

	baseUrl := os.Getenv("BASE_URL")

	sm := scs.New()
	app := &App{
		Api:        NewApi(conf, db, sm),
		AuthServer: NewAuthServer(GetAuthProviders(baseUrl), db, sm),
		UIDir:      os.Args[2],
		SM:         sm,
	}
	log.Fatal(http.ListenAndServe(":5000", sm.LoadAndSave(app)))
}
