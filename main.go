package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"forum/internal"
	"github.com/go-chi/chi/v5"
)

func main() {
	dbPath := os.Getenv("FORUM_DB_PATH")
	if dbPath == "" {
		dbPath = "forum.db"
	}

	err := internal.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Erreur initialisation BD: %v", err)
	}
	defer internal.CloseDB()

	r := chi.NewRouter()

	// Les pages publiques chargent quand même l'utilisateur pour adapter l'affichage.
	r.Get("/", internal.LoadUserIfAuthenticated(internal.HomeHandler))
	r.Get("/register", internal.LoadUserIfAuthenticated(internal.RegisterPageHandler))
	r.Post("/register", internal.RegisterHandler)
	r.Get("/login", internal.LoginPageHandler)
	r.Post("/login", internal.LoginHandler)
	r.Get("/logout", internal.LogoutHandler)

	// Le forum se lit sans compte, mais les actions d'écriture restent protégées.
	r.Get("/forum", internal.LoadUserIfAuthenticated(internal.ForumHandler))
	r.Get("/profile", internal.RequireAuth(internal.ProfileHandler))
	r.Post("/forum/post", internal.RequireAuth(internal.CreatePostHandler))
	r.Get("/forum/post/{id}", internal.LoadUserIfAuthenticated(internal.TopicHandler))

	r.Post("/forum/post/{id}/reply", internal.RequireAuth(internal.CreateReplyHandler))
	r.Post("/forum/post/{id}/vote", internal.RequireAuth(internal.VoteTopicHandler))
	r.Post("/forum/post/{id}/comments/{commentID}/vote", internal.RequireAuth(internal.VoteCommentHandler))

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.NotFound(http.HandlerFunc(internal.NotFoundHandler))

	adresse := ":8080"
	fmt.Println("🚀 Serveur démarré sur http://localhost:8080")
	fmt.Println("Appuyez sur Ctrl+C pour arrêter")

	err = http.ListenAndServe(adresse, internal.RecoveryMiddleware(r))
	if err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}
