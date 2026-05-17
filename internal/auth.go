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

// Type privé pour éviter les collisions avec d'autres valeurs du contexte.
type userContextKeyType string

const userContextKey userContextKeyType = "user"

// HashPassword génère un hash bcrypt du mot de passe.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// VerifyPassword compare le mot de passe saisi avec son hash.
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateToken génère un token aléatoire utilisé pour les sessions.
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateUserSession crée une session côté base et renvoie le token à mettre en cookie.
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

// GetUserFromRequest retrouve l'utilisateur grâce au cookie de session.
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

// GetUserFromContext récupère l'utilisateur déjà chargé par un middleware.
func GetUserFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// SetSessionCookie stocke seulement le token dans le navigateur.
// Les vraies informations de session restent en base de données.
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

// ClearSessionCookie supprime la session en base puis expire le cookie côté navigateur.
func ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		DeleteSession(cookie.Value)
	}

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

// RequireAuth bloque la page si l'utilisateur n'est pas connecté.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromRequest(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// LoadUserIfAuthenticated charge l'utilisateur si possible, sans forcer la connexion.
func LoadUserIfAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromRequest(r)

		ctx := context.WithValue(r.Context(), userContextKey, user)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
