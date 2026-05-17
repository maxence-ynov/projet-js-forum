package internal

import (
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxFormBodySize = 1 << 20

	minUsernameLength = 3
	maxUsernameLength = 30
	maxEmailLength    = 254
	minPasswordLength = 6
	maxPasswordLength = 72

	minTitleLength   = 3
	maxTitleLength   = 120
	minPostLength    = 10
	maxPostLength    = 5000
	maxCommentLength = 2000
)

var (
	usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	categoryPattern = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)
	htmlTagPattern  = regexp.MustCompile(`<[^>]+>`)
)

func parseLimitedForm(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodySize)
	return r.ParseForm()
}

func validateUsername(username string) string {
	switch {
	case username == "":
		return "Le nom d'utilisateur est obligatoire"
	case utf8.RuneCountInString(username) < minUsernameLength:
		return "Le nom d'utilisateur doit contenir au moins 3 caractères"
	case utf8.RuneCountInString(username) > maxUsernameLength:
		return "Le nom d'utilisateur ne doit pas dépasser 30 caractères"
	case containsHTML(username):
		return "Le nom d'utilisateur ne doit pas contenir de HTML"
	case !usernamePattern.MatchString(username):
		return "Le nom d'utilisateur ne peut contenir que des lettres, chiffres, tirets et underscores"
	default:
		return ""
	}
}

func validateEmail(email string) string {
	switch {
	case email == "":
		return "L'email est obligatoire"
	case utf8.RuneCountInString(email) > maxEmailLength:
		return "L'email est trop long"
	case containsHTML(email):
		return "L'email ne doit pas contenir de HTML"
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "L'email est invalide"
	}

	return ""
}

func validatePassword(password string) string {
	switch {
	case password == "":
		return "Le mot de passe est obligatoire"
	case len(password) < minPasswordLength:
		return "Le mot de passe doit contenir au moins 6 caractères"
	case len(password) > maxPasswordLength:
		return "Le mot de passe ne doit pas dépasser 72 caractères"
	default:
		return ""
	}
}

func validateTopicInput(title, content, categoryID string) string {
	switch {
	case title == "" || content == "" || categoryID == "":
		return "Tous les champs sont obligatoires"
	case utf8.RuneCountInString(title) < minTitleLength:
		return "Le titre doit contenir au moins 3 caractères"
	case utf8.RuneCountInString(title) > maxTitleLength:
		return "Le titre ne doit pas dépasser 120 caractères"
	case containsHTML(title):
		return "Le titre ne doit pas contenir de HTML"
	case utf8.RuneCountInString(content) < minPostLength:
		return "Le contenu doit contenir au moins 10 caractères"
	case utf8.RuneCountInString(content) > maxPostLength:
		return "Le contenu ne doit pas dépasser 5000 caractères"
	case containsHTML(content):
		return "Le contenu ne doit pas contenir de HTML"
	case !isValidCategoryID(categoryID):
		return "La catégorie sélectionnée est invalide"
	default:
		return ""
	}
}

func validateCommentInput(content string) string {
	switch {
	case content == "":
		return "Le commentaire ne peut pas être vide"
	case utf8.RuneCountInString(content) > maxCommentLength:
		return "Le commentaire ne doit pas dépasser 2000 caractères"
	case containsHTML(content):
		return "Le commentaire ne doit pas contenir de HTML ou de script"
	default:
		return ""
	}
}

func isValidCategoryID(categoryID string) bool {
	return categoryPattern.MatchString(categoryID)
}

func isValidUUID(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}

func containsHTML(value string) bool {
	value = strings.TrimSpace(value)
	return htmlTagPattern.MatchString(value)
}
