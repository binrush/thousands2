package main

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type RuntimeConfig struct {
	Datadir      string
	ItemsPerPage int
}

type App struct {
	Api        *Api
	AuthServer *AuthServer
	UIDir      string
	SM         *scs.SessionManager
	router     *chi.Mux
}

func NewAppServer(conf *RuntimeConfig, storage *Storage, sm *scs.SessionManager, uiDir string) *App {
	app := &App{
		Api:        NewApi(conf, storage, sm),
		AuthServer: NewAuthServer(GetAuthProviders(os.Getenv("BASE_URL")), storage, sm),
		UIDir:      uiDir,
		SM:         sm,
		router:     chi.NewRouter(),
	}

	// Set up routes
	app.router.Use(sm.LoadAndSave)
	app.router.Mount("/api", app.Api.router)
	app.router.Mount("/auth", app.AuthServer.router)

	// Serve static files for all other routes
	app.router.NotFound(http.FileServer(http.Dir(app.UIDir)).ServeHTTP)

	return app
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

	storage := NewStorage(db)

	log.Printf("Loading summits data to database...")
	err = storage.LoadSummits(conf.Datadir)
	if err != nil {
		log.Fatalf("Failed to load summits: %v", err)
	}
	log.Printf("Summits data loaded")

	sm := scs.New()
	app := NewAppServer(conf, storage, sm, os.Args[2])
	log.Fatal(http.ListenAndServe(":5000", app.router))
}
