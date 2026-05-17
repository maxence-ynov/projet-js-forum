package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// RecoveryMiddleware transforme les panics inattendues en vraie page 500.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("Panic récupérée: %v\n%s", recovered, debug.Stack())
				RenderServerError(w, r)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	renderErrorPage(w, r, http.StatusNotFound, "Page introuvable", "La page demandée n'existe pas ou a été déplacée.")
}

func RenderServerError(w http.ResponseWriter, r *http.Request) {
	renderErrorPage(w, r, http.StatusInternalServerError, "Erreur serveur", "Une erreur interne est survenue. Réessayez dans quelques instants.")
}

func RenderBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	renderErrorPage(w, r, http.StatusBadRequest, "Requête invalide", message)
}

func renderErrorPage(w http.ResponseWriter, r *http.Request, statusCode int, title, message string) {
	user := GetUserFromRequest(r)
	data := PageData{
		IsLoggedIn:   user != nil,
		User:         user,
		StatusCode:   statusCode,
		ErrorTitle:   title,
		ErrorMessage: message,
	}

	renderTemplateWithStatus(w, statusCode, "layout.html", "error.html", data)
}

func renderTemplate(w http.ResponseWriter, layoutName, pageName string, data interface{}) {
	renderTemplateWithStatus(w, http.StatusOK, layoutName, pageName, data)
}

// renderTemplateWithStatus charge le rendu dans un buffer avant d'écrire la réponse.
// Cela permet d'afficher une page 500 propre si un template échoue.
func renderTemplateWithStatus(w http.ResponseWriter, statusCode int, layoutName, pageName string, data interface{}) {
	layoutPath := filepath.Join("templates", layoutName)
	pagePath := filepath.Join("templates", pageName)

	t, err := template.New(layoutName).Funcs(template.FuncMap{
		"contains": containsAny,
	}).ParseFiles(layoutPath, pagePath)
	if err != nil {
		log.Printf("Erreur parsing template: %v", err)
		writeFallbackError(w, http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, layoutName, data)
	if err != nil {
		log.Printf("Erreur exécution template: %v", err)
		writeFallbackError(w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Printf("Erreur écriture réponse: %v", err)
	}
}

func writeFallbackError(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, http.StatusText(statusCode))
}

func formValues(values map[string]string) map[string]string {
	cleaned := make(map[string]string, len(values))
	for key, value := range values {
		cleaned[key] = strings.TrimSpace(value)
	}
	return cleaned
}

func renderFormError(w http.ResponseWriter, statusCode int, templateName string, data PageData) {
	renderTemplateWithStatus(w, statusCode, "layout.html", templateName, data)
}
