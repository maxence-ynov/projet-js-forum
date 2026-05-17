package internal

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var errNotFound = errors.New("ressource introuvable")

type topicListMode int

const (
	allTopics topicListMode = iota
	latestTopics
)

// HomeHandler affiche la page d'accueil.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	selectedCategoryID := strings.TrimSpace(r.URL.Query().Get("category"))

	categories, topics, selectedCategoryName, err := loadCategoryTopics(selectedCategoryID, latestTopics, 10)
	if err != nil {
		if errors.Is(err, errNotFound) {
			NotFoundHandler(w, r)
			return
		}
		log.Printf("Erreur récupération accueil: %v", err)
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

// RegisterPageHandler affiche la page d'inscription.
func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())

	if user != nil {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	data := PageData{
		IsLoggedIn: false,
	}

	renderTemplate(w, "layout.html", "register.html", data)
}

// RegisterHandler traite l'inscription d'un nouvel utilisateur.
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
		renderAuthFormError(w, "register.html", message, map[string]string{"username": username, "email": email})
		return
	}

	if message := validateEmail(email); message != "" {
		renderAuthFormError(w, "register.html", message, map[string]string{"username": username, "email": email})
		return
	}

	if message := validatePassword(password); message != "" {
		renderAuthFormError(w, "register.html", message, map[string]string{"username": username, "email": email})
		return
	}

	if password != passwordConfirm {
		renderAuthFormError(w, "register.html", "Les mots de passe ne correspondent pas", map[string]string{"username": username, "email": email})
		return
	}

	existingUser, err := GetUserByUsername(username)
	if err != nil {
		log.Printf("Erreur BD: %v", err)
		RenderServerError(w, r)
		return
	}

	if existingUser != nil {
		renderAuthFormError(w, "register.html", "Ce nom d'utilisateur est déjà pris", map[string]string{"username": username, "email": email})
		return
	}

	existingEmail, err := GetUserByEmail(email)
	if err != nil {
		log.Printf("Erreur BD: %v", err)
		RenderServerError(w, r)
		return
	}

	if existingEmail != nil {
		renderAuthFormError(w, "register.html", "Cet email est déjà utilisé", map[string]string{"username": username, "email": email})
		return
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Printf("Erreur hashage: %v", err)
		RenderServerError(w, r)
		return
	}

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

	// Après inscription, l'utilisateur est connecté directement.
	token, err := CreateUserSession(newUser.ID)
	if err != nil {
		log.Printf("Erreur création session: %v", err)
		RenderServerError(w, r)
		return
	}

	SetSessionCookie(w, r, token)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

// LoginPageHandler affiche la page de connexion.
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromRequest(r)

	if user != nil {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	data := PageData{
		IsLoggedIn: false,
	}

	renderTemplate(w, "layout.html", "login.html", data)
}

// LoginHandler traite la connexion d'un utilisateur.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := parseLimitedForm(w, r)
	if err != nil {
		RenderBadRequest(w, r, "Le formulaire envoyé est invalide.")
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	if username == "" || password == "" {
		renderAuthFormError(w, "login.html", "Tous les champs sont obligatoires", map[string]string{"username": username})
		return
	}
	if utf8.RuneCountInString(username) > maxUsernameLength || len(password) > maxPasswordLength {
		renderAuthFormError(w, "login.html", "Identifiant ou mot de passe incorrect", map[string]string{"username": username})
		return
	}

	user, err := GetUserByUsername(username)
	if err != nil {
		log.Printf("Erreur BD: %v", err)
		RenderServerError(w, r)
		return
	}

	if user == nil {
		renderAuthFormError(w, "login.html", "Identifiant ou mot de passe incorrect", map[string]string{"username": username})
		return
	}

	if !VerifyPassword(user.Password, password) {
		renderAuthFormError(w, "login.html", "Identifiant ou mot de passe incorrect", map[string]string{"username": username})
		return
	}

	token, err := CreateUserSession(user.ID)
	if err != nil {
		log.Printf("Erreur création session: %v", err)
		RenderServerError(w, r)
		return
	}

	SetSessionCookie(w, r, token)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

// LogoutHandler déconnecte l'utilisateur.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearSessionCookie(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ForumHandler affiche la page du forum.
func ForumHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	selectedCategoryID := strings.TrimSpace(r.URL.Query().Get("category"))
	activeFilter := strings.TrimSpace(r.URL.Query().Get("filter"))
	if activeFilter == "" {
		activeFilter = "all"
	}
	if activeFilter != "all" {
		selectedCategoryID = ""
	}

	categories, topics, selectedCategoryName, err := loadForumTopics(user, selectedCategoryID, activeFilter)
	if err != nil {
		if errors.Is(err, errNotFound) {
			NotFoundHandler(w, r)
			return
		}
		if errors.Is(err, errAuthRequired) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		log.Printf("Erreur récupération forum: %v", err)
		RenderServerError(w, r)
		return
	}

	if err := fillTopicListDetails(topics, user); err != nil {
		log.Printf("Erreur préparation liste forum: %v", err)
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
		ActiveFilter:         activeFilter,
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

// CreatePostHandler crée un nouveau sujet.
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

	comments, err := fillTopicDetails(topic, user)
	if err != nil {
		log.Printf("Erreur préparation page sujet: %v", err)
		RenderServerError(w, r)
		return
	}

	data := PageData{
		IsLoggedIn: user != nil,
		User:       user,
		Topic:      topic,
		Comments:   comments,
	}

	renderTemplate(w, "layout.html", "topic.html", data)
}

// CreateReplyHandler crée une réponse sur un sujet.
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

// VoteTopicHandler enregistre le vote d'un utilisateur sur un sujet.
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

	voteType, ok := readVoteType(r)
	if !ok {
		RenderBadRequest(w, r, "Le vote envoyé est invalide.")
		return
	}

	// SetVote remplace l'ancien vote au lieu d'ajouter une deuxième ligne.
	err = SetVote(user.ID, "topic", topic.ID, voteType)
	if err != nil {
		log.Printf("Erreur enregistrement vote sujet: %v", err)
		RenderServerError(w, r)
		return
	}

	http.Redirect(w, r, "/forum/post/"+topic.ID, http.StatusSeeOther)
}

// VoteCommentHandler enregistre le vote d'un utilisateur sur un commentaire.
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

	voteType, ok := readVoteType(r)
	if !ok {
		RenderBadRequest(w, r, "Le vote envoyé est invalide.")
		return
	}

	// Même logique que pour les sujets: un seul vote par utilisateur et commentaire.
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

	comments, err := fillTopicDetails(topic, user)
	if err != nil {
		log.Printf("Erreur préparation page sujet: %v", err)
		RenderServerError(w, r)
		return
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

	categories, topics, _, err := loadCategoryTopics("", allTopics, 0)
	if err != nil {
		log.Printf("Erreur récupération formulaire forum: %v", err)
		RenderServerError(w, r)
		return
	}

	data := PageData{
		IsLoggedIn:         true,
		User:               user,
		Categories:         categories,
		Topics:             topics,
		SelectedCategoryID: selectedCategoryID,
		ActiveFilter:       "all",
		Message:            message,
		FormValues: formValues(map[string]string{
			"title":   r.FormValue("title"),
			"content": r.FormValue("content"),
		}),
	}

	renderFormError(w, http.StatusUnprocessableEntity, "forum.html", data)
}

var errAuthRequired = errors.New("authentification requise")

func loadForumTopics(user *User, selectedCategoryID, activeFilter string) ([]Category, []Topic, string, error) {
	categories, err := GetAllCategories()
	if err != nil {
		return nil, nil, "", err
	}

	selectedCategoryName := ""
	if selectedCategoryID != "" {
		if !isValidCategoryID(selectedCategoryID) {
			return nil, nil, "", errNotFound
		}

		category, err := GetCategoryByID(selectedCategoryID)
		if err != nil {
			return nil, nil, "", err
		}
		if category == nil {
			return nil, nil, "", errNotFound
		}
		selectedCategoryName = category.Name
	}

	switch activeFilter {
	case "all":
		if selectedCategoryID != "" {
			topics, err := GetTopicsByCategory(selectedCategoryID)
			return categories, topics, selectedCategoryName, err
		}
		topics, err := GetAllTopics()
		return categories, topics, selectedCategoryName, err
	case "mine":
		if user == nil {
			return nil, nil, "", errAuthRequired
		}
		topics, err := GetTopicsByUser(user.ID)
		return categories, topics, selectedCategoryName, err
	case "liked":
		if user == nil {
			return nil, nil, "", errAuthRequired
		}
		topics, err := GetLikedTopicsByUser(user.ID)
		return categories, topics, selectedCategoryName, err
	default:
		return nil, nil, "", errNotFound
	}
}

func loadCategoryTopics(selectedCategoryID string, mode topicListMode, limit int) ([]Category, []Topic, string, error) {
	categories, err := GetAllCategories()
	if err != nil {
		return nil, nil, "", err
	}

	if selectedCategoryID != "" {
		if !isValidCategoryID(selectedCategoryID) {
			return nil, nil, "", errNotFound
		}

		category, err := GetCategoryByID(selectedCategoryID)
		if err != nil {
			return nil, nil, "", err
		}
		if category == nil {
			return nil, nil, "", errNotFound
		}

		topics, err := topicsForCategory(selectedCategoryID, mode, limit)
		return categories, topics, category.Name, err
	}

	topics, err := topicsForCategory("", mode, limit)
	return categories, topics, "", err
}

func topicsForCategory(categoryID string, mode topicListMode, limit int) ([]Topic, error) {
	if categoryID != "" {
		if mode == latestTopics {
			return GetLatestTopicsByCategory(categoryID, limit)
		}
		return GetTopicsByCategory(categoryID)
	}

	if mode == latestTopics {
		return GetLatestTopics(limit)
	}
	return GetAllTopics()
}

func fillTopicListDetails(topics []Topic, user *User) error {
	for i := range topics {
		comments, err := GetCommentsByTopic(topics[i].ID)
		if err != nil {
			return err
		}
		topics[i].Comments = comments
		topics[i].CommentCount = len(comments)

		if user != nil {
			vote, err := GetUserVote(user.ID, "topic", topics[i].ID)
			if err != nil {
				return err
			}
			topics[i].UserVote = vote
		}
	}

	return nil
}

func fillTopicDetails(topic *Topic, user *User) ([]Comment, error) {
	comments, err := GetCommentsByTopic(topic.ID)
	if err != nil {
		return nil, err
	}

	topic.Comments = comments
	topic.CommentCount = len(comments)

	if user == nil {
		return comments, nil
	}

	topic.UserVote, err = GetUserVote(user.ID, "topic", topic.ID)
	if err != nil {
		return nil, err
	}

	for i := range comments {
		comments[i].UserVote, err = GetUserVote(user.ID, "comment", comments[i].ID)
		if err != nil {
			return nil, err
		}
	}
	topic.Comments = comments

	return comments, nil
}

func renderAuthFormError(w http.ResponseWriter, templateName, message string, values map[string]string) {
	data := PageData{
		Message:    message,
		FormValues: formValues(values),
	}
	renderFormError(w, http.StatusUnprocessableEntity, templateName, data)
}

func readVoteType(r *http.Request) (string, bool) {
	voteType := strings.TrimSpace(r.FormValue("vote"))
	return voteType, voteType == "like" || voteType == "dislike"
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
