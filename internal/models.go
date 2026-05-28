package internal

import "time"

// User représente un utilisateur du forum.
type User struct {
	ID        string    `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
}

// Session représente une session utilisateur.
type Session struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Token     string    `db:"token"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}

// Category représente une catégorie du forum.
type Category struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

// Topic représente un sujet du forum.
type Topic struct {
	ID           string    `db:"id"`
	CategoryID   string    `db:"category_id"`
	CategoryName string    // Nom de catégorie affiché dans les templates
	UserID       string    `db:"user_id"`
	Username     string    // Nom d'auteur affiché dans les templates
	Title        string    `db:"title"`
	Content      string    `db:"content"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	Comments     []Comment // Commentaires chargés pour la page du sujet
	CommentCount int       // Nombre de commentaires
	Likes        int       // Nombre de likes
	Dislikes     int       // Nombre de dislikes
	UserVote     string    // Vote de l'utilisateur connecté: "like", "dislike" ou vide
}

// Comment représente un commentaire sur un sujet.
type Comment struct {
	ID        string    `db:"id"`
	TopicID   string    `db:"topic_id"`
	UserID    string    `db:"user_id"`
	Username  string    // Nom d'auteur affiché dans les templates
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Likes     int       // Nombre de likes
	Dislikes  int       // Nombre de dislikes
	UserVote  string    // Vote de l'utilisateur connecté: "like", "dislike" ou vide
}

// Vote représente un vote unique pour un sujet ou comentaire.
type Vote struct {
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	TargetType string    `db:"target_type"`
	TargetID   string    `db:"target_id"`
	VoteType   string    `db:"vote_type"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// PageData regroupe les données envoyées aux templates.
type PageData struct {
	IsLoggedIn           bool
	User                 *User
	Categories           []Category
	Topics               []Topic
	CreatedTopics        []Topic
	LikedTopics          []Topic
	Topic                *Topic
	Comments             []Comment
	SelectedCategoryID   string
	SelectedCategoryName string
	ActiveFilter         string
	Message              string // Message affiché après une erreur de formulaire
	FormValues           map[string]string
	StatusCode           int
	ErrorTitle           string
	ErrorMessage         string
}
