package internal

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// HomeHandler affiche la page d'accueil
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())

	data := PageData{
		IsLoggedIn: user != nil,
		User:       user,
	}

	renderTemplate(w, "layout.html", "index.html", data)
}

// RegisterPageHandler affiche la page d'inscription
func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())

	// Si déjà connecté, rediriger vers le forum
	if user != nil {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	data := PageData{
		IsLoggedIn: false,
	}

	renderTemplate(w, "layout.html", "register.html", data)
}

// RegisterHandler traite l'inscription d'un nouvel utilisateur
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Erreur de traitement du formulaire", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")

	// Valider les données
	if username == "" || email == "" || password == "" {
		data := PageData{
			Message: "Tous les champs sont obligatoires",
		}
		renderTemplate(w, "layout.html", "register.html", data)
		return
	}

	// Validation des longueurs
	if len(username) < 3 {
		data := PageData{
			Message: "Le nom d'utilisateur doit contenir au moins 3 caractères",
		}
		renderTemplate(w, "layout.html", "register.html", data)
		return
	}

	if len(password) < 6 {
		data := PageData{
			Message: "Le mot de passe doit contenir au moins 6 caractères",
		}
		renderTemplate(w, "layout.html", "register.html", data)
		return
	}

	if password != passwordConfirm {
		data := PageData{
			Message: "Les mots de passe ne correspondent pas",
		}
		renderTemplate(w, "layout.html", "register.html", data)
		return
	}

	// Vérifier si le username existe déjà
	existingUser, err := GetUserByUsername(username)
	if err != nil {
		log.Printf("Erreur BD: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		data := PageData{
			Message: "Ce nom d'utilisateur est déjà pris",
		}
		renderTemplate(w, "layout.html", "register.html", data)
		return
	}

	// Vérifier si l'email existe déjà
	existingEmail, err := GetUserByEmail(email)
	if err != nil {
		log.Printf("Erreur BD: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	if existingEmail != nil {
		data := PageData{
			Message: "Cet email est déjà utilisé",
		}
		renderTemplate(w, "layout.html", "register.html", data)
		return
	}

	// Hasher le mot de passe
	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Printf("Erreur hashage: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	// Créer le nouvel utilisateur
	newUser := &User{
		ID:       uuid.New().String(),
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}

	err = CreateUser(newUser)
	if err != nil {
		log.Printf("Erreur création utilisateur: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	// Créer une session automatique et rediriger
	token, err := CreateUserSession(newUser.ID)
	if err != nil {
		log.Printf("Erreur création session: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	SetSessionCookie(w, token)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

// LoginPageHandler affiche la page de connexion
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)

	// Si déjà connecté, rediriger vers le forum
	if user != nil {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	data := PageData{
		IsLoggedIn: false,
	}

	renderTemplate(w, "layout.html", "login.html", data)
}

// LoginHandler traite la connexion d'un utilisateur
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Erreur de traitement du formulaire", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	// Valider les données
	if username == "" || password == "" {
		data := PageData{
			Message: "Tous les champs sont obligatoires",
		}
		renderTemplate(w, "layout.html", "login.html", data)
		return
	}

	// Récupérer l'utilisateur
	user, err := GetUserByUsername(username)
	if err != nil {
		log.Printf("Erreur BD: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	if user == nil {
		data := PageData{
			Message: "Identifiant ou mot de passe incorrect",
		}
		renderTemplate(w, "layout.html", "login.html", data)
		return
	}

	// Vérifier le mot de passe
	if !VerifyPassword(user.Password, password) {
		data := PageData{
			Message: "Identifiant ou mot de passe incorrect",
		}
		renderTemplate(w, "layout.html", "login.html", data)
		return
	}

	// Créer une session
	token, err := CreateUserSession(user.ID)
	if err != nil {
		log.Printf("Erreur création session: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	SetSessionCookie(w, token)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

// LogoutHandler déconnecte l'utilisateur
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearSessionCookie(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ForumHandler affiche la page du forum
func ForumHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)

	// Récupérer tous les posts
	posts, err := GetAllPosts()
	if err != nil {
		log.Printf("Erreur récupération posts: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	// Charger les réponses pour chaque post
	for i := range posts {
		replies, err := GetRepliesByPostID(posts[i].ID)
		if err != nil {
			log.Printf("Erreur récupération réponses: %v", err)
			continue
		}
		posts[i].Replies = replies
	}

	data := PageData{
		IsLoggedIn: true,
		User:       user,
		Posts:      posts,
	}

	renderTemplate(w, "layout.html", "forum.html", data)
}

// CreatePostHandler crée un nouveau post
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Erreur de traitement du formulaire", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))

	// Valider les données
	if title == "" || content == "" {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	// Créer le post
	post := &Post{
		ID:      uuid.New().String(),
		UserID:  user.ID,
		Title:   title,
		Content: content,
	}

	err = CreatePost(post)
	if err != nil {
		log.Printf("Erreur création post: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

// CreateReplyHandler crée une réponse à un post
func CreateReplyHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Erreur de traitement du formulaire", http.StatusBadRequest)
		return
	}

	// Récupérer l'ID du post depuis l'URL
	postID := r.PathValue("id")
	if postID == "" {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))

	// Valider le contenu
	if content == "" {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	// Créer la réponse
	reply := &Reply{
		ID:      uuid.New().String(),
		PostID:  postID,
		UserID:  user.ID,
		Content: content,
	}

	err = CreateReply(reply)
	if err != nil {
		log.Printf("Erreur création réponse: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

// renderTemplate charge et exécute les templates HTML
func renderTemplate(w http.ResponseWriter, layoutName, pageName string, data interface{}) {
	layoutPath := filepath.Join("templates", layoutName)
	pagePath := filepath.Join("templates", pageName)

	t, err := template.ParseFiles(layoutPath, pagePath)
	if err != nil {
		log.Printf("Erreur parsing template: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, layoutName, data)
	if err != nil {
		log.Printf("Erreur exécution template: %v", err)
		http.Error(w, "Erreur serveur", http.StatusInternalServerError)
		return
	}
}
