package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/internal"
)

func main() {
	// Initialiser la base de données
	err := internal.InitDB("forum.db")
	if err != nil {
		log.Fatalf("Erreur initialisation BD: %v", err)
	}
	defer internal.CloseDB()

	// Créer le multiplexeur de routes
	mux := http.NewServeMux()

	// Routes publiques avec middleware optionnel (charge l'utilisateur s'il est connecté)
	mux.HandleFunc("GET /", internal.LoadUserIfAuthenticated(internal.HomeHandler))
	mux.HandleFunc("GET /register", internal.LoadUserIfAuthenticated(internal.RegisterPageHandler))
	mux.HandleFunc("POST /register", internal.RegisterHandler)
	mux.HandleFunc("GET /login", internal.LoginPageHandler)
	mux.HandleFunc("POST /login", internal.LoginHandler)
	mux.HandleFunc("GET /logout", internal.LogoutHandler)

	// Routes protégées (forum)
	mux.HandleFunc("GET /forum", internal.RequireAuth(internal.ForumHandler))
	mux.HandleFunc("POST /forum/post", internal.RequireAuth(internal.CreatePostHandler))
	mux.HandleFunc("GET /forum/post/{id}", internal.LoadUserIfAuthenticated(internal.TopicHandler))
	mux.HandleFunc("POST /forum/post/{id}/reply", internal.RequireAuth(internal.CreateReplyHandler))
	mux.HandleFunc("POST /forum/post/{id}/vote", internal.RequireAuth(internal.VoteTopicHandler))
	mux.HandleFunc("POST /forum/post/{id}/comments/{commentID}/vote", internal.RequireAuth(internal.VoteCommentHandler))

	// Servir les fichiers statiques
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Démarrer le serveur
	adresse := "localhost:8080"
	fmt.Printf("🚀 Serveur démarré sur http://%s\n", adresse)
	fmt.Println("Appuyez sur Ctrl+C pour arrêter")

	err = http.ListenAndServe(adresse, mux)
	if err != nil {
		log.Fatalf("Erreur serveur: %v", err)
	}
}
