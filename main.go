package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/internal"
)

func main() {
	err := internal.InitDB("forum.db")
	if err != nil {
		log.Fatalf("Erreur initialisation BD: %v", err)
	}
	defer internal.CloseDB()

	mux := http.NewServeMux()

	// Les pages publiques chargent quand même l'utilisateur pour adapter l'affichage.
	mux.HandleFunc("GET /{$}", internal.LoadUserIfAuthenticated(internal.HomeHandler))
	mux.HandleFunc("GET /register", internal.LoadUserIfAuthenticated(internal.RegisterPageHandler))
	mux.HandleFunc("POST /register", internal.RegisterHandler)
	mux.HandleFunc("GET /login", internal.LoginPageHandler)
	mux.HandleFunc("POST /login", internal.LoginHandler)
	mux.HandleFunc("GET /logout", internal.LogoutHandler)

	// Le forum se lit sans compte, mais les actions d'écriture restent protégées.
	mux.HandleFunc("GET /forum", internal.LoadUserIfAuthenticated(internal.ForumHandler))
	mux.HandleFunc("GET /profile", internal.RequireAuth(internal.ProfileHandler))
	mux.HandleFunc("POST /forum/post", internal.RequireAuth(internal.CreatePostHandler))
	mux.HandleFunc("GET /forum/post/{id}", internal.LoadUserIfAuthenticated(internal.TopicHandler))

	mux.HandleFunc("POST /forum/post/{id}/reply", internal.RequireAuth(internal.CreateReplyHandler))
	mux.HandleFunc("POST /forum/post/{id}/vote", internal.RequireAuth(internal.VoteTopicHandler))
	mux.HandleFunc("POST /forum/post/{id}/comments/{commentID}/vote", internal.RequireAuth(internal.VoteCommentHandler))

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/", internal.NotFoundHandler)

	adresse := ":8080"
	fmt.Println("🚀 Serveur démarré sur http://localhost:8080")
	fmt.Println("Appuyez sur Ctrl+C pour arrêter")

	err = http.ListenAndServe(adresse, internal.RecoveryMiddleware(mux))
	if err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}
