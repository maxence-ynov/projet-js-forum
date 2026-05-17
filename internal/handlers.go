package internal

import (
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// HomeHandler affiche la page d'accueil
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	selectedCategoryID := strings.TrimSpace(r.URL.Query().Get("category"))
	selectedCategoryName := ""

	categories, err := GetAllCategories()
	if err != nil {
		log.Printf("Erreur récupération catégories: %v", err)
		RenderServerError(w, r)
		return
	}

	var topics []Topic
	if selectedCategoryID != "" {
		if !isValidCategoryID(selectedCategoryID) {
			NotFoundHandler(w, r)
			return
		}
		category, err := GetCategoryByID(selectedCategoryID)
		if err != nil {
			log.Printf("Erreur récupération catégorie: %v", err)
			RenderServerError(w, r)
			return
		}
		if category == nil {
			NotFoundHandler(w, r)
			return
		}
		selectedCategoryName = category.Name
		topics, err = GetLatestTopicsByCategory(selectedCategoryID, 10)
	} else {
		topics, err = GetLatestTopics(10)
	}
	if err != nil {
		log.Printf("Erreur récupération derniers sujets: %v", err)
		RenderServerError(w, r)
		return
	}

	data := PageData{
		IsLoggedIn:           user != nil,
		User:                 user,
		Categories:           categories,
		Topics:               topics,
		SelectedCategoryID:   selectedCategoryID,
		SelectedCategoryName: selectedCategoryName,
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
	err := parseLimitedForm(w, r)
	if err != nil {
		RenderBadRequest(w, r, "Le formulaire envoyé est invalide.")
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")

	if message := validateUsername(username); message != "" {
		data := PageData{
			Message:    message,
			FormValues: formValues(map[string]string{"username": username, "email": email}),
		}
		renderFormError(w, http.StatusUnprocessableEntity, "register.html", data)
		return
	}

	if message := validateEmail(email); message != "" {
		data := PageData{
			Message:    message,
			FormValues: formValues(map[string]string{"username": username, "email": email}),
		}
		renderFormError(w, http.StatusUnprocessableEntity, "register.html", data)
		return
	}

	if message := validatePassword(password); message != "" {
		data := PageData{
			Message:    message,
			FormValues: formValues(map[string]string{"username": username, "email": email}),
		}
		renderFormError(w, http.StatusUnprocessableEntity, "register.html", data)
		return
	}

	if password != passwordConfirm {
		data := PageData{
			Message:    "Les mots de passe ne correspondent pas",
			FormValues: formValues(map[string]string{"username": username, "email": email}),
		}
		renderFormError(w, http.StatusUnprocessableEntity, "register.html", data)
		return
	}

	// Vérifier si le username existe déjà
	existingUser, err := GetUserByUsername(username)
	if err != nil {
		log.Printf("Erreur BD: %v", err)
		RenderServerError(w, r)
		return
	}

	if existingUser != nil {
		data := PageData{
			Message:    "Ce nom d'utilisateur est déjà pris",
			FormValues: formValues(map[string]string{"username": username, "email": email}),
		}
		renderFormError(w, http.StatusUnprocessableEntity, "register.html", data)
		return
	}

	// Vérifier si l'email existe déjà
	existingEmail, err := GetUserByEmail(email)
	if err != nil {
		log.Printf("Erreur BD: %v", err)
		RenderServerError(w, r)
		return
	}

	if existingEmail != nil {
		data := PageData{
			Message:    "Cet email est déjà utilisé",
			FormValues: formValues(map[string]string{"username": username, "email": email}),
		}
		renderFormError(w, http.StatusUnprocessableEntity, "register.html", data)
		return
	}

	// Hasher le mot de passe
	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Printf("Erreur hashage: %v", err)
		RenderServerError(w, r)
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
		RenderServerError(w, r)
		return
	}

	// Créer une session automatique et rediriger
	token, err := CreateUserSession(newUser.ID)
	if err != nil {
		log.Printf("Erreur création session: %v", err)
		RenderServerError(w, r)
		return
	}

	SetSessionCookie(w, r, token)
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
	err := parseLimitedForm(w, r)
	if err != nil {
		RenderBadRequest(w, r, "Le formulaire envoyé est invalide.")
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	// Valider les données
	if username == "" || password == "" {
		data := PageData{
			Message:    "Tous les champs sont obligatoires",
			FormValues: formValues(map[string]string{"username": username}),
		}
		renderFormError(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}
	if utf8.RuneCountInString(username) > maxUsernameLength || len(password) > maxPasswordLength {
		data := PageData{
			Message:    "Identifiant ou mot de passe incorrect",
			FormValues: formValues(map[string]string{"username": username}),
		}
		renderFormError(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	// Récupérer l'utilisateur
	user, err := GetUserByUsername(username)
	if err != nil {
		log.Printf("Erreur BD: %v", err)
		RenderServerError(w, r)
		return
	}

	if user == nil {
		data := PageData{
			Message:    "Identifiant ou mot de passe incorrect",
			FormValues: formValues(map[string]string{"username": username}),
		}
		renderFormError(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	// Vérifier le mot de passe
	if !VerifyPassword(user.Password, password) {
		data := PageData{
			Message:    "Identifiant ou mot de passe incorrect",
			FormValues: formValues(map[string]string{"username": username}),
		}
		renderFormError(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	// Créer une session
	token, err := CreateUserSession(user.ID)
	if err != nil {
		log.Printf("Erreur création session: %v", err)
		RenderServerError(w, r)
		return
	}

	SetSessionCookie(w, r, token)
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
	selectedCategoryID := strings.TrimSpace(r.URL.Query().Get("category"))
	selectedCategoryName := ""

	categories, err := GetAllCategories()
	if err != nil {
		log.Printf("Erreur récupération catégories: %v", err)
		RenderServerError(w, r)
		return
	}

	var topics []Topic
	if selectedCategoryID != "" {
		if !isValidCategoryID(selectedCategoryID) {
			NotFoundHandler(w, r)
			return
		}
		category, err := GetCategoryByID(selectedCategoryID)
		if err != nil {
			log.Printf("Erreur récupération catégorie: %v", err)
			RenderServerError(w, r)
			return
		}
		if category == nil {
			NotFoundHandler(w, r)
			return
		}
		selectedCategoryName = category.Name
		topics, err = GetTopicsByCategory(selectedCategoryID)
	} else {
		topics, err = GetAllTopics()
	}
	if err != nil {
		log.Printf("Erreur récupération sujets: %v", err)
		RenderServerError(w, r)
		return
	}

	// Charger les commentaires pour chaque sujet
	for i := range topics {
		comments, err := GetCommentsByTopic(topics[i].ID)
		if err != nil {
			log.Printf("Erreur récupération commentaires: %v", err)
			continue
		}
		topics[i].Comments = comments
		topics[i].CommentCount = len(comments)
	}

	data := PageData{
		IsLoggedIn:           true,
		User:                 user,
		Categories:           categories,
		Topics:               topics,
		SelectedCategoryID:   selectedCategoryID,
		SelectedCategoryName: selectedCategoryName,
	}

	renderTemplate(w, "layout.html", "forum.html", data)
}

// ProfileHandler affiche les sujets de l'utilisateur connecté et ceux qu'il a likés.
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	createdTopics, err := GetTopicsByUser(user.ID)
	if err != nil {
		log.Printf("Erreur récupération sujets utilisateur: %v", err)
		RenderServerError(w, r)
		return
	}

	likedTopics, err := GetLikedTopicsByUser(user.ID)
	if err != nil {
		log.Printf("Erreur récupération sujets likés: %v", err)
		RenderServerError(w, r)
		return
	}

	data := PageData{
		IsLoggedIn:    true,
		User:          user,
		CreatedTopics: createdTopics,
		LikedTopics:   likedTopics,
	}

	renderTemplate(w, "layout.html", "profile.html", data)
}

// CreatePostHandler crée un nouveau post
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := parseLimitedForm(w, r)
	if err != nil {
		RenderBadRequest(w, r, "Le formulaire envoyé est invalide.")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	categoryID := strings.TrimSpace(r.FormValue("category_id"))

	if message := validateTopicInput(title, content, categoryID); message != "" {
		renderForumFormWithMessage(w, r, message, categoryID)
		return
	}

	category, err := GetCategoryByID(categoryID)
	if err != nil {
		log.Printf("Erreur récupération catégorie: %v", err)
		RenderServerError(w, r)
		return
	}
	if category == nil {
		renderForumFormWithMessage(w, r, "La catégorie sélectionnée est invalide", categoryID)
		return
	}

	// Créer le sujet
	topic := &Topic{
		ID:         uuid.New().String(),
		CategoryID: categoryID,
		UserID:     user.ID,
		Title:      title,
		Content:    content,
	}

	err = CreateTopic(topic)
	if err != nil {
		log.Printf("Erreur création sujet: %v", err)
		RenderServerError(w, r)
		return
	}

	http.Redirect(w, r, "/forum/post/"+topic.ID, http.StatusSeeOther)
}

// TopicHandler affiche un sujet et ses commentaires.
func TopicHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	topicID := r.PathValue("id")
	if !isValidUUID(topicID) {
		NotFoundHandler(w, r)
		return
	}

	topic, err := GetTopicByID(topicID)
	if err != nil {
		log.Printf("Erreur récupération sujet: %v", err)
		RenderServerError(w, r)
		return
	}
	if topic == nil {
		NotFoundHandler(w, r)
		return
	}

	comments, err := GetCommentsByTopic(topic.ID)
	if err != nil {
		log.Printf("Erreur récupération commentaires: %v", err)
		RenderServerError(w, r)
		return
	}

	likes, err := GetLikesByTopic(topic.ID)
	if err != nil {
		log.Printf("Erreur récupération likes sujet: %v", err)
		RenderServerError(w, r)
		return
	}

	dislikes, err := GetDislikesByTopic(topic.ID)
	if err != nil {
		log.Printf("Erreur récupération dislikes sujet: %v", err)
		RenderServerError(w, r)
		return
	}

	for i := range comments {
		comments[i].Likes, err = GetLikesByComment(comments[i].ID)
		if err != nil {
			log.Printf("Erreur récupération likes commentaire: %v", err)
			RenderServerError(w, r)
			return
		}

		comments[i].Dislikes, err = GetDislikesByComment(comments[i].ID)
		if err != nil {
			log.Printf("Erreur récupération dislikes commentaire: %v", err)
			RenderServerError(w, r)
			return
		}
	}

	topic.Comments = comments
	topic.CommentCount = len(comments)
	topic.Likes = likes
	topic.Dislikes = dislikes
	if user != nil {
		topic.UserVote, err = GetUserVote(user.ID, "topic", topic.ID)
		if err != nil {
			log.Printf("Erreur récupération vote utilisateur sujet: %v", err)
			RenderServerError(w, r)
			return
		}

		for i := range comments {
			comments[i].UserVote, err = GetUserVote(user.ID, "comment", comments[i].ID)
			if err != nil {
				log.Printf("Erreur récupération vote utilisateur commentaire: %v", err)
				RenderServerError(w, r)
				return
			}
		}
	}

	data := PageData{
		IsLoggedIn: user != nil,
		User:       user,
		Topic:      topic,
		Comments:   comments,
	}

	renderTemplate(w, "layout.html", "topic.html", data)
}

// CreateReplyHandler crée une réponse à un post
func CreateReplyHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := parseLimitedForm(w, r)
	if err != nil {
		RenderBadRequest(w, r, "Le formulaire envoyé est invalide.")
		return
	}

	// Récupérer l'ID du sujet depuis l'URL
	topicID := r.PathValue("id")
	if !isValidUUID(topicID) {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))

	topic, err := GetTopicByID(topicID)
	if err != nil {
		log.Printf("Erreur récupération sujet: %v", err)
		RenderServerError(w, r)
		return
	}
	if topic == nil {
		NotFoundHandler(w, r)
		return
	}

	if message := validateCommentInput(content); message != "" {
		renderTopicWithMessage(w, r, topic, message)
		return
	}

	// Créer le commentaire
	comment := &Comment{
		ID:      uuid.New().String(),
		TopicID: topicID,
		UserID:  user.ID,
		Content: content,
	}

	err = CreateComment(comment)
	if err != nil {
		log.Printf("Erreur création commentaire: %v", err)
		RenderServerError(w, r)
		return
	}

	http.Redirect(w, r, "/forum/post/"+topicID, http.StatusSeeOther)
}

// VoteTopicHandler enregistre un like ou dislike sur un sujet.
func VoteTopicHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	topicID := r.PathValue("id")
	if !isValidUUID(topicID) {
		NotFoundHandler(w, r)
		return
	}

	topic, err := GetTopicByID(topicID)
	if err != nil {
		log.Printf("Erreur récupération sujet: %v", err)
		RenderServerError(w, r)
		return
	}
	if topic == nil {
		NotFoundHandler(w, r)
		return
	}

	err = parseLimitedForm(w, r)
	if err != nil {
		RenderBadRequest(w, r, "Le formulaire envoyé est invalide.")
		return
	}

	voteType := strings.TrimSpace(r.FormValue("vote"))
	if voteType != "like" && voteType != "dislike" {
		RenderBadRequest(w, r, "Le vote envoyé est invalide.")
		return
	}

	err = SetVote(user.ID, "topic", topic.ID, voteType)
	if err != nil {
		log.Printf("Erreur enregistrement vote sujet: %v", err)
		RenderServerError(w, r)
		return
	}

	http.Redirect(w, r, "/forum/post/"+topic.ID, http.StatusSeeOther)
}

// VoteCommentHandler enregistre un like ou dislike sur un commentaire.
func VoteCommentHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	topicID := r.PathValue("id")
	commentID := r.PathValue("commentID")
	if !isValidUUID(topicID) || !isValidUUID(commentID) {
		NotFoundHandler(w, r)
		return
	}

	topic, err := GetTopicByID(topicID)
	if err != nil {
		log.Printf("Erreur récupération sujet: %v", err)
		RenderServerError(w, r)
		return
	}
	if topic == nil {
		NotFoundHandler(w, r)
		return
	}

	comment, err := GetCommentByID(commentID)
	if err != nil {
		log.Printf("Erreur récupération commentaire: %v", err)
		RenderServerError(w, r)
		return
	}
	if comment == nil || comment.TopicID != topic.ID {
		NotFoundHandler(w, r)
		return
	}

	err = parseLimitedForm(w, r)
	if err != nil {
		RenderBadRequest(w, r, "Le formulaire envoyé est invalide.")
		return
	}

	voteType := strings.TrimSpace(r.FormValue("vote"))
	if voteType != "like" && voteType != "dislike" {
		RenderBadRequest(w, r, "Le vote envoyé est invalide.")
		return
	}

	err = SetVote(user.ID, "comment", comment.ID, voteType)
	if err != nil {
		log.Printf("Erreur enregistrement vote commentaire: %v", err)
		RenderServerError(w, r)
		return
	}

	http.Redirect(w, r, "/forum/post/"+topic.ID, http.StatusSeeOther)
}

func renderTopicWithMessage(w http.ResponseWriter, r *http.Request, topic *Topic, message string) {
	user := GetUserFromRequest(r)

	comments, err := GetCommentsByTopic(topic.ID)
	if err != nil {
		log.Printf("Erreur récupération commentaires: %v", err)
		RenderServerError(w, r)
		return
	}

	likes, err := GetLikesByTopic(topic.ID)
	if err != nil {
		log.Printf("Erreur récupération likes sujet: %v", err)
		RenderServerError(w, r)
		return
	}

	dislikes, err := GetDislikesByTopic(topic.ID)
	if err != nil {
		log.Printf("Erreur récupération dislikes sujet: %v", err)
		RenderServerError(w, r)
		return
	}

	for i := range comments {
		comments[i].Likes, err = GetLikesByComment(comments[i].ID)
		if err != nil {
			log.Printf("Erreur récupération likes commentaire: %v", err)
			RenderServerError(w, r)
			return
		}

		comments[i].Dislikes, err = GetDislikesByComment(comments[i].ID)
		if err != nil {
			log.Printf("Erreur récupération dislikes commentaire: %v", err)
			RenderServerError(w, r)
			return
		}
	}

	topic.Comments = comments
	topic.CommentCount = len(comments)
	topic.Likes = likes
	topic.Dislikes = dislikes
	if user != nil {
		topic.UserVote, err = GetUserVote(user.ID, "topic", topic.ID)
		if err != nil {
			log.Printf("Erreur récupération vote utilisateur sujet: %v", err)
			RenderServerError(w, r)
			return
		}

		for i := range comments {
			comments[i].UserVote, err = GetUserVote(user.ID, "comment", comments[i].ID)
			if err != nil {
				log.Printf("Erreur récupération vote utilisateur commentaire: %v", err)
				RenderServerError(w, r)
				return
			}
		}
	}

	data := PageData{
		IsLoggedIn: user != nil,
		User:       user,
		Topic:      topic,
		Comments:   comments,
		Message:    message,
	}

	renderFormError(w, http.StatusUnprocessableEntity, "topic.html", data)
}

func renderForumFormWithMessage(w http.ResponseWriter, r *http.Request, message, selectedCategoryID string) {
	user := GetUserFromRequest(r)

	categories, err := GetAllCategories()
	if err != nil {
		log.Printf("Erreur récupération catégories: %v", err)
		RenderServerError(w, r)
		return
	}

	topics, err := GetAllTopics()
	if err != nil {
		log.Printf("Erreur récupération sujets: %v", err)
		RenderServerError(w, r)
		return
	}

	data := PageData{
		IsLoggedIn:         true,
		User:               user,
		Categories:         categories,
		Topics:             topics,
		SelectedCategoryID: selectedCategoryID,
		Message:            message,
		FormValues: formValues(map[string]string{
			"title":   r.FormValue("title"),
			"content": r.FormValue("content"),
		}),
	}

	renderFormError(w, http.StatusUnprocessableEntity, "forum.html", data)
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
