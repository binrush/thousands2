package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

//go:embed ui/dist
var uiFS embed.FS

type RuntimeConfig struct {
	Datadir      string
	ItemsPerPage int
}

type App struct {
	Api        *Api
	AuthServer *AuthServer
	SM         *scs.SessionManager
	router     *chi.Mux
}

func NewAppServer(conf *RuntimeConfig, storage *Storage, sm *scs.SessionManager) *App {
	app := &App{
		Api:        NewApi(conf, storage, sm),
		AuthServer: NewAuthServer(GetAuthProviders(os.Getenv("BASE_URL")), storage, sm),
		SM:         sm,
		router:     chi.NewRouter(),
	}

	// Set up routes
	app.router.Use(sm.LoadAndSave)

	// Get the embedded filesystem
	fsys, err := fs.Sub(uiFS, "ui/dist")
	if err != nil {
		log.Fatalf("Failed to get UI subdirectory: %v", err)
	}

	// Create a file server for static assets
	fileServer := http.FileServer(http.FS(fsys))

	// Serve static files for /assets/ and other static resources
	app.router.Get("/assets/{path}", fileServer.ServeHTTP)
	app.router.Get("/favicon.ico", fileServer.ServeHTTP)
	app.router.Get("/logo.svg", fileServer.ServeHTTP)
	app.router.Get("/logo_notext.svg", fileServer.ServeHTTP)
	app.router.Get("/climber_no_photo.svg", fileServer.ServeHTTP)
	app.router.Get("/vklogo.svg", fileServer.ServeHTTP)

	// Mount API and auth routes
	app.router.Mount("/api", app.Api.router)
	app.router.Mount("/auth", app.AuthServer.router)

	// Serve index.html for the root path
	app.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		serveIndexHTML(w, fsys)
	})

	// Serve index.html for all other routes to support Vue Router
	app.router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		serveIndexHTML(w, fsys)
	})

	return app
}

func serveIndexHTML(w http.ResponseWriter, fsys fs.FS) {
	// Serve index.html
	index, err := fsys.Open("index.html")
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	defer index.Close()

	// Read the entire file
	content, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set content type and serve
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: api <datadir> <db_path>")
	}

	conf := &RuntimeConfig{
		Datadir:      path.Clean(os.Args[1]),
		ItemsPerPage: 20,
	}

	db, err := NewDatabase(path.Clean(os.Args[2]))
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	log.Printf("Starting migrations...")
	err = Migrate(db)
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
	sm.Store = sqlite3store.New(db)

	app := NewAppServer(conf, storage, sm)

	log.Println("Server starting on :5000")
	log.Fatal(http.ListenAndServe(":5000", app.router))
}
