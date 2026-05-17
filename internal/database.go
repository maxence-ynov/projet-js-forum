package internal

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// DB est la variable globale pour la connexion à la base de données
var DB *sql.DB

// InitDB initialise la base de données SQLite
// Ouvre forum.db et crée les tables automatiquement
func InitDB(nomFichier string) error {
	var err error

	// Ouvrir la connexion à SQLite
	DB, err = sql.Open("sqlite3", nomFichier)
	if err != nil {
		return fmt.Errorf("erreur ouverture BD: %w", err)
	}

	// Vérifier que la connexion fonctionne
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("erreur connexion BD: %w", err)
	}

	log.Printf("✓ Base de données '%s' connectée", nomFichier)

	// Créer toutes les tables
	err = createAllTables()
	if err != nil {
		return fmt.Errorf("erreur création tables: %w", err)
	}

	log.Println("✓ Tables créées/vérifiées")

	err = seedDefaultCategories()
	if err != nil {
		return fmt.Errorf("erreur création catégories par défaut: %w", err)
	}

	return nil
}

// CloseDB ferme la connexion à la base de données
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// createAllTables crée toutes les tables nécessaires
func createAllTables() error {
	// Table users
	if err := createUsersTable(); err != nil {
		return err
	}

	// Table categories
	if err := createCategoriesTable(); err != nil {
		return err
	}

	// Table topics
	if err := createTopicsTable(); err != nil {
		return err
	}

	// Table comments
	if err := createCommentsTable(); err != nil {
		return err
	}

	// Table likes
	if err := createLikesTable(); err != nil {
		return err
	}

	// Table dislikes
	if err := createDislikesTable(); err != nil {
		return err
	}

	// Table votes
	if err := createVotesTable(); err != nil {
		return err
	}

	// Table sessions
	if err := createSessionsTable(); err != nil {
		return err
	}

	return nil
}

func seedDefaultCategories() error {
	categories := []Category{
		{ID: "general", Name: "Général", Description: "Discussions générales autour du forum."},
		{ID: "questions", Name: "Questions", Description: "Questions, aide et réponses de la communauté."},
		{ID: "annonces", Name: "Annonces", Description: "Informations importantes et nouveautés."},
	}

	query := "INSERT OR IGNORE INTO categories (id, name, description) VALUES (?, ?, ?)"
	for _, category := range categories {
		_, err := DB.Exec(query, category.ID, category.Name, category.Description)
		if err != nil {
			return err
		}
	}

	return nil
}

// createUsersTable crée la table des utilisateurs
func createUsersTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := DB.Exec(query)
	return err
}

// createCategoriesTable crée la table des catégories
func createCategoriesTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS categories (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := DB.Exec(query)
	return err
}

// createTopicsTable crée la table des sujets/topics
func createTopicsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS topics (
		id TEXT PRIMARY KEY,
		category_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	_, err := DB.Exec(query)
	return err
}

// createCommentsTable crée la table des commentaires
func createCommentsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS comments (
		id TEXT PRIMARY KEY,
		topic_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	_, err := DB.Exec(query)
	return err
}

// createLikesTable crée la table des likes
func createLikesTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS likes (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		topic_id TEXT,
		comment_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
		FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
		UNIQUE(user_id, topic_id, comment_id)
	);
	`
	_, err := DB.Exec(query)
	return err
}

// createDislikesTable crée la table des dislikes.
func createDislikesTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS dislikes (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		topic_id TEXT,
		comment_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
		FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
		UNIQUE(user_id, topic_id, comment_id)
	);
	`
	_, err := DB.Exec(query)
	return err
}

// createVotesTable crée la table des votes avec un seul vote par utilisateur et cible.
func createVotesTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS votes (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		target_type TEXT NOT NULL CHECK(target_type IN ('topic', 'comment')),
		target_id TEXT NOT NULL,
		vote_type TEXT NOT NULL CHECK(vote_type IN ('like', 'dislike')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, target_type, target_id)
	);
	`
	_, err := DB.Exec(query)
	return err
}

// createSessionsTable crée la table des sessions
func createSessionsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	_, err := DB.Exec(query)
	return err
}

// ============= USERS =============

// GetUserByUsername récupère un utilisateur par son nom d'utilisateur
func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := "SELECT id, username, email, password, created_at FROM users WHERE username = ?"

	err := DB.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur requête utilisateur: %w", err)
	}

	return user, nil
}

// GetUserByEmail récupère un utilisateur par son email
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := "SELECT id, username, email, password, created_at FROM users WHERE email = ?"

	err := DB.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur requête utilisateur: %w", err)
	}

	return user, nil
}

// GetUserByID récupère un utilisateur par son ID
func GetUserByID(id string) (*User, error) {
	user := &User{}
	query := "SELECT id, username, email, password, created_at FROM users WHERE id = ?"

	err := DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur requête utilisateur: %w", err)
	}

	return user, nil
}

// CreateUser crée un nouvel utilisateur
func CreateUser(user *User) error {
	query := "INSERT INTO users (id, username, email, password) VALUES (?, ?, ?, ?)"

	_, err := DB.Exec(query, user.ID, user.Username, user.Email, user.Password)
	if err != nil {
		return fmt.Errorf("erreur création utilisateur: %w", err)
	}

	return nil
}

// ============= SESSIONS =============

// CreateSession crée une nouvelle session
func CreateSession(session *Session) error {
	query := "INSERT INTO sessions (id, user_id, token, expires_at) VALUES (?, ?, ?, ?)"

	_, err := DB.Exec(query, session.ID, session.UserID, session.Token, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("erreur création session: %w", err)
	}

	return nil
}

// GetSessionByToken récupère une session par token (si elle n'a pas expiré)
func GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	query := `
		SELECT id, user_id, token, created_at, expires_at 
		FROM sessions 
		WHERE token = ? AND expires_at > CURRENT_TIMESTAMP
	`

	err := DB.QueryRow(query, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.CreatedAt, &session.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur requête session: %w", err)
	}

	return session, nil
}

// DeleteSession supprime une session par token
func DeleteSession(token string) error {
	query := "DELETE FROM sessions WHERE token = ?"

	_, err := DB.Exec(query, token)
	if err != nil {
		return fmt.Errorf("erreur suppression session: %w", err)
	}

	return nil
}

// ============= CATEGORIES =============

// CreateCategory crée une nouvelle catégorie
func CreateCategory(category *Category) error {
	query := "INSERT INTO categories (id, name, description) VALUES (?, ?, ?)"

	_, err := DB.Exec(query, category.ID, category.Name, category.Description)
	if err != nil {
		return fmt.Errorf("erreur création catégorie: %w", err)
	}

	return nil
}

// GetAllCategories récupère toutes les catégories
func GetAllCategories() ([]Category, error) {
	query := "SELECT id, name, description, created_at FROM categories ORDER BY name ASC"

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur requête catégories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("erreur scan catégorie: %w", err)
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

// GetCategoryByID récupère une catégorie par son ID
func GetCategoryByID(id string) (*Category, error) {
	category := &Category{}
	query := "SELECT id, name, description, created_at FROM categories WHERE id = ?"

	err := DB.QueryRow(query, id).Scan(
		&category.ID, &category.Name, &category.Description, &category.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur requête catégorie: %w", err)
	}

	return category, nil
}

// ============= TOPICS =============

// CreateTopic crée un nouveau topic/sujet
func CreateTopic(topic *Topic) error {
	query := `
		INSERT INTO topics (id, category_id, user_id, title, content) 
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := DB.Exec(query, topic.ID, topic.CategoryID, topic.UserID, topic.Title, topic.Content)
	if err != nil {
		return fmt.Errorf("erreur création topic: %w", err)
	}

	return nil
}

// GetAllTopics récupère tous les topics avec les infos utilisateur
func GetAllTopics() ([]Topic, error) {
	query := `
		SELECT t.id, t.category_id, t.user_id, t.title, t.content,
		       t.created_at, t.updated_at, u.username, c.name,
		       (SELECT COUNT(*) FROM comments co WHERE co.topic_id = t.id),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'like'),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'dislike')
		FROM topics t
		JOIN users u ON t.user_id = u.id
		JOIN categories c ON t.category_id = c.id
		ORDER BY t.updated_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur requête topics: %w", err)
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		var topic Topic
		err := rows.Scan(
			&topic.ID, &topic.CategoryID, &topic.UserID, &topic.Title, &topic.Content,
			&topic.CreatedAt, &topic.UpdatedAt, &topic.Username, &topic.CategoryName,
			&topic.CommentCount, &topic.Likes, &topic.Dislikes,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan topic: %w", err)
		}
		topics = append(topics, topic)
	}

	return topics, rows.Err()
}

// GetLatestTopics récupère les derniers topics avec leur nombre de commentaires.
func GetLatestTopics(limit int) ([]Topic, error) {
	query := `
		SELECT t.id, t.category_id, t.user_id, t.title, t.content,
		       t.created_at, t.updated_at, u.username, c.name,
		       (SELECT COUNT(*) FROM comments co WHERE co.topic_id = t.id),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'like'),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'dislike')
		FROM topics t
		JOIN users u ON t.user_id = u.id
		JOIN categories c ON t.category_id = c.id
		ORDER BY t.created_at DESC
		LIMIT ?
	`

	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("erreur requête derniers topics: %w", err)
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		var topic Topic
		err := rows.Scan(
			&topic.ID, &topic.CategoryID, &topic.UserID, &topic.Title, &topic.Content,
			&topic.CreatedAt, &topic.UpdatedAt, &topic.Username, &topic.CategoryName,
			&topic.CommentCount, &topic.Likes, &topic.Dislikes,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan dernier topic: %w", err)
		}
		topics = append(topics, topic)
	}

	return topics, rows.Err()
}

// GetTopicsByCategory récupère tous les topics d'une catégorie
func GetTopicsByCategory(categoryID string) ([]Topic, error) {
	query := `
		SELECT t.id, t.category_id, t.user_id, t.title, t.content,
		       t.created_at, t.updated_at, u.username, c.name,
		       (SELECT COUNT(*) FROM comments co WHERE co.topic_id = t.id),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'like'),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'dislike')
		FROM topics t
		JOIN users u ON t.user_id = u.id
		JOIN categories c ON t.category_id = c.id
		WHERE t.category_id = ?
		ORDER BY t.updated_at DESC
	`

	rows, err := DB.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("erreur requête topics: %w", err)
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		var topic Topic
		err := rows.Scan(
			&topic.ID, &topic.CategoryID, &topic.UserID, &topic.Title, &topic.Content,
			&topic.CreatedAt, &topic.UpdatedAt, &topic.Username, &topic.CategoryName,
			&topic.CommentCount, &topic.Likes, &topic.Dislikes,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan topic: %w", err)
		}
		topics = append(topics, topic)
	}

	return topics, rows.Err()
}

// GetLatestTopicsByCategory récupère les derniers topics d'une catégorie.
func GetLatestTopicsByCategory(categoryID string, limit int) ([]Topic, error) {
	query := `
		SELECT t.id, t.category_id, t.user_id, t.title, t.content,
		       t.created_at, t.updated_at, u.username, c.name,
		       (SELECT COUNT(*) FROM comments co WHERE co.topic_id = t.id),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'like'),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'dislike')
		FROM topics t
		JOIN users u ON t.user_id = u.id
		JOIN categories c ON t.category_id = c.id
		WHERE t.category_id = ?
		ORDER BY t.created_at DESC
		LIMIT ?
	`

	rows, err := DB.Query(query, categoryID, limit)
	if err != nil {
		return nil, fmt.Errorf("erreur requête derniers topics par catégorie: %w", err)
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		var topic Topic
		err := rows.Scan(
			&topic.ID, &topic.CategoryID, &topic.UserID, &topic.Title, &topic.Content,
			&topic.CreatedAt, &topic.UpdatedAt, &topic.Username, &topic.CategoryName,
			&topic.CommentCount, &topic.Likes, &topic.Dislikes,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan dernier topic par catégorie: %w", err)
		}
		topics = append(topics, topic)
	}

	return topics, rows.Err()
}

// GetTopicByID récupère un topic par son ID
func GetTopicByID(id string) (*Topic, error) {
	topic := &Topic{}
	query := `
		SELECT t.id, t.category_id, t.user_id, t.title, t.content, 
		       t.created_at, t.updated_at, u.username, c.name,
		       (SELECT COUNT(*) FROM comments co WHERE co.topic_id = t.id),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'like'),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'dislike')
		FROM topics t
		JOIN users u ON t.user_id = u.id
		JOIN categories c ON t.category_id = c.id
		WHERE t.id = ?
	`

	err := DB.QueryRow(query, id).Scan(
		&topic.ID, &topic.CategoryID, &topic.UserID, &topic.Title, &topic.Content,
		&topic.CreatedAt, &topic.UpdatedAt, &topic.Username, &topic.CategoryName,
		&topic.CommentCount, &topic.Likes, &topic.Dislikes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur requête topic: %w", err)
	}

	return topic, nil
}

// ============= COMMENTS =============

// CreateComment crée un nouveau commentaire
func CreateComment(comment *Comment) error {
	query := "INSERT INTO comments (id, topic_id, user_id, content) VALUES (?, ?, ?, ?)"

	_, err := DB.Exec(query, comment.ID, comment.TopicID, comment.UserID, comment.Content)
	if err != nil {
		return fmt.Errorf("erreur création commentaire: %w", err)
	}

	return nil
}

// GetCommentsByTopic récupère tous les commentaires d'un topic
func GetCommentsByTopic(topicID string) ([]Comment, error) {
	query := `
		SELECT c.id, c.topic_id, c.user_id, c.content, c.created_at, c.updated_at, u.username,
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'comment' AND v.target_id = c.id AND v.vote_type = 'like'),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'comment' AND v.target_id = c.id AND v.vote_type = 'dislike')
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.topic_id = ?
		ORDER BY c.created_at ASC
	`

	rows, err := DB.Query(query, topicID)
	if err != nil {
		return nil, fmt.Errorf("erreur requête commentaires: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&comment.ID, &comment.TopicID, &comment.UserID, &comment.Content,
			&comment.CreatedAt, &comment.UpdatedAt, &comment.Username,
			&comment.Likes, &comment.Dislikes,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan commentaire: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

// GetCommentByID récupère un commentaire par son ID
func GetCommentByID(id string) (*Comment, error) {
	comment := &Comment{}
	query := `
		SELECT c.id, c.topic_id, c.user_id, c.content, c.created_at, c.updated_at, u.username,
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'comment' AND v.target_id = c.id AND v.vote_type = 'like'),
		       (SELECT COUNT(*) FROM votes v WHERE v.target_type = 'comment' AND v.target_id = c.id AND v.vote_type = 'dislike')
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?
	`

	err := DB.QueryRow(query, id).Scan(
		&comment.ID, &comment.TopicID, &comment.UserID, &comment.Content,
		&comment.CreatedAt, &comment.UpdatedAt, &comment.Username,
		&comment.Likes, &comment.Dislikes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur requête commentaire: %w", err)
	}

	return comment, nil
}

// ============= LIKES =============

// CreateLike crée un like pour un topic ou commentaire
func CreateLike(like *Like) error {
	query := "INSERT INTO likes (id, user_id, topic_id, comment_id) VALUES (?, ?, ?, ?)"

	_, err := DB.Exec(query, like.ID, like.UserID, like.TopicID, like.CommentID)
	if err != nil {
		return fmt.Errorf("erreur création like: %w", err)
	}

	return nil
}

// DeleteLike supprime un like
func DeleteLike(id string) error {
	query := "DELETE FROM likes WHERE id = ?"

	_, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression like: %w", err)
	}

	return nil
}

// GetLikesByTopic récupère le nombre de likes pour un topic
func GetLikesByTopic(topicID string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM votes WHERE target_type = 'topic' AND target_id = ? AND vote_type = 'like'"

	err := DB.QueryRow(query, topicID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage likes: %w", err)
	}

	return count, nil
}

// GetLikesByComment récupère le nombre de likes pour un commentaire
func GetLikesByComment(commentID string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM votes WHERE target_type = 'comment' AND target_id = ? AND vote_type = 'like'"

	err := DB.QueryRow(query, commentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage likes: %w", err)
	}

	return count, nil
}

// GetDislikesByTopic récupère le nombre de dislikes pour un topic.
func GetDislikesByTopic(topicID string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM votes WHERE target_type = 'topic' AND target_id = ? AND vote_type = 'dislike'"

	err := DB.QueryRow(query, topicID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage dislikes: %w", err)
	}

	return count, nil
}

// GetDislikesByComment récupère le nombre de dislikes pour un commentaire.
func GetDislikesByComment(commentID string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM votes WHERE target_type = 'comment' AND target_id = ? AND vote_type = 'dislike'"

	err := DB.QueryRow(query, commentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage dislikes: %w", err)
	}

	return count, nil
}

// UserLikedTopic vérifie si un utilisateur a aimé un topic
func UserLikedTopic(userID, topicID string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM votes WHERE user_id = ? AND target_type = 'topic' AND target_id = ? AND vote_type = 'like'"

	err := DB.QueryRow(query, userID, topicID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("erreur vérification like: %w", err)
	}

	return count > 0, nil
}

// UserLikedComment vérifie si un utilisateur a aimé un commentaire
func UserLikedComment(userID, commentID string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM votes WHERE user_id = ? AND target_type = 'comment' AND target_id = ? AND vote_type = 'like'"

	err := DB.QueryRow(query, userID, commentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("erreur vérification like: %w", err)
	}

	return count > 0, nil
}

// GetUserVote récupère le vote actuel d'un utilisateur pour une cible.
func GetUserVote(userID, targetType, targetID string) (string, error) {
	var voteType string
	query := `
		SELECT vote_type
		FROM votes
		WHERE user_id = ? AND target_type = ? AND target_id = ?
	`

	err := DB.QueryRow(query, userID, targetType, targetID).Scan(&voteType)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("erreur récupération vote utilisateur: %w", err)
	}

	return voteType, nil
}

// SetVote crée ou remplace le vote d'un utilisateur pour une cible.
func SetVote(userID, targetType, targetID, voteType string) error {
	if targetType != "topic" && targetType != "comment" {
		return fmt.Errorf("type de cible invalide")
	}
	if voteType != "like" && voteType != "dislike" {
		return fmt.Errorf("type de vote invalide")
	}

	query := `
		INSERT INTO votes (id, user_id, target_type, target_id, vote_type)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id, target_type, target_id)
		DO UPDATE SET vote_type = excluded.vote_type, updated_at = CURRENT_TIMESTAMP
	`

	_, err := DB.Exec(query, uuid.New().String(), userID, targetType, targetID, voteType)
	if err != nil {
		return fmt.Errorf("erreur enregistrement vote: %w", err)
	}

	return nil
}
