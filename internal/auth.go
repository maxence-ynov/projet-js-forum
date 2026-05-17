package internal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "forum_session"
	SessionDuration   = 24 * time.Hour
)

// Clé privée pour stocker l'utilisateur dans le contexte (évite les collisions)
type userContextKeyType string

const userContextKey userContextKeyType = "user"

// HashPassword génère un hash bcrypt du mot de passe
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// VerifyPassword vérifie un mot de passe contre son hash
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateToken génère un token aléatoire
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateUserSession crée une nouvelle session pour un utilisateur
func CreateUserSession(userID string) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}

	session := &Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(SessionDuration),
	}

	err = CreateSession(session)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetUserFromRequest récupère l'utilisateur à partir de la requête (cookie de session)
func GetUserFromRequest(r *http.Request) *User {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil
	}

	session, err := GetSessionByToken(cookie.Value)
	if err != nil || session == nil {
		return nil
	}

	user, err := GetUserByID(session.UserID)
	if err != nil || user == nil {
		return nil
	}

	return user
}

// GetUserFromContext récupère l'utilisateur depuis le contexte de la requête
func GetUserFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// SetSessionCookie ajoute le cookie de session dans la réponse
func SetSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   isHTTPSRequest(r),
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie supprime le cookie de session
func ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		// Supprimer la session de la base de données
		DeleteSession(cookie.Value)
	}

	// Supprimer le cookie
	clearCookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isHTTPSRequest(r),
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, clearCookie)
}

func isHTTPSRequest(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

// RequireAuth est un middleware qui vérifie l'authentification
// et ajoute l'utilisateur dans le contexte de la requête
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromRequest(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Ajouter l'utilisateur dans le contexte
		ctx := context.WithValue(r.Context(), userContextKey, user)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// LoadUserIfAuthenticated est un middleware optionnel qui charge l'utilisateur
// s'il est connecté, sans redirection. Utile pour les pages publiques.
func LoadUserIfAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromRequest(r)
		// user peut être nil, ce qui est acceptable ici

		// Ajouter l'utilisateur dans le contexte (nil ou non)
		ctx := context.WithValue(r.Context(), userContextKey, user)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
