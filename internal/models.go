package internal

import "time"

// ============= MODELS =============

// User représente un utilisateur du forum
type User struct {
	ID        string    `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
}

// Session représente une session utilisateur
type Session struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Token     string    `db:"token"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}

// Category représente une catégorie du forum
type Category struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

// Topic représente un sujet du forum
type Topic struct {
	ID           string    `db:"id"`
	CategoryID   string    `db:"category_id"`
	CategoryName string    // Pour l'affichage du nom de catégorie
	UserID       string    `db:"user_id"`
	Username     string    // Pour l'affichage du nom d'auteur
	Title        string    `db:"title"`
	Content      string    `db:"content"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	Comments     []Comment // Les commentaires du topic
	CommentCount int       // Nombre de commentaires
	Likes        int       // Nombre de likes
	Dislikes     int       // Nombre de dislikes
}

// Comment représente un commentaire sur un topic
type Comment struct {
	ID        string    `db:"id"`
	TopicID   string    `db:"topic_id"`
	UserID    string    `db:"user_id"`
	Username  string    // Pour l'affichage du nom d'auteur
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Likes     int       // Nombre de likes
	Dislikes  int       // Nombre de dislikes
}

// Like représente un like pour un topic ou commentaire
type Like struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	TopicID   *string   `db:"topic_id"`   // NULL si like sur un commentaire
	CommentID *string   `db:"comment_id"` // NULL si like sur un topic
	CreatedAt time.Time `db:"created_at"`
}

// PageData est utilisé pour passer les données aux templates
type PageData struct {
	IsLoggedIn           bool
	User                 *User
	Categories           []Category
	Topics               []Topic
	Topic                *Topic
	Comments             []Comment
	SelectedCategoryID   string
	SelectedCategoryName string
	Message              string // Pour afficher des messages de succès/erreur
}
