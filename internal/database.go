package internal

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

const topicSelectColumns = `
	t.id, t.category_id, t.user_id, t.title, t.content,
	t.created_at, t.updated_at, u.username, c.name,
	(SELECT COUNT(*) FROM comments co WHERE co.topic_id = t.id),
	(SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'like'),
	(SELECT COUNT(*) FROM votes v WHERE v.target_type = 'topic' AND v.target_id = t.id AND v.vote_type = 'dislike')
`

const topicJoins = `
	FROM topics t
	JOIN users u ON t.user_id = u.id
	JOIN categories c ON t.category_id = c.id
`

const commentSelectColumns = `
	c.id, c.topic_id, c.user_id, c.content, c.created_at, c.updated_at, u.username,
	(SELECT COUNT(*) FROM votes v WHERE v.target_type = 'comment' AND v.target_id = c.id AND v.vote_type = 'like'),
	(SELECT COUNT(*) FROM votes v WHERE v.target_type = 'comment' AND v.target_id = c.id AND v.vote_type = 'dislike')
`

// DB garde la connexion SQLite partagée par les handlers.
var DB *sql.DB

// InitDB ouvre la base SQLite et vérifie que les tables existent.
func InitDB(nomFichier string) error {
	var err error

	DB, err = sql.Open("sqlite3", nomFichier)
	if err != nil {
		return fmt.Errorf("erreur ouverture BD: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("erreur connexion BD: %w", err)
	}

	if err := enableForeignKeys(); err != nil {
		return fmt.Errorf("erreur activation clés étrangères: %w", err)
	}

	log.Printf("✓ Base de données '%s' connectée", nomFichier)

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

// CloseDB ferme la connexion à la base de données.
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func enableForeignKeys() error {
	_, err := DB.Exec("PRAGMA foreign_keys = ON")
	return err
}

// createAllTables garde le démarrage simple: l'application crée son schéma si besoin.
func createAllTables() error {
	if err := createUsersTable(); err != nil {
		return err
	}

	if err := createCategoriesTable(); err != nil {
		return err
	}

	if err := createTopicsTable(); err != nil {
		return err
	}

	if err := createCommentsTable(); err != nil {
		return err
	}

	if err := createVotesTable(); err != nil {
		return err
	}

	if err := dropOldVoteTables(); err != nil {
		return err
	}

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

// createUsersTable crée la table des utilisateurs.
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

// createCategoriesTable crée la table des catégories.
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

// createTopicsTable crée la table des sujets.
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

// createCommentsTable crée la table des commentaires.
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

// createVotesTable limite chaque utilisateur à un vote par sujet ou commentaire.
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

func dropOldVoteTables() error {
	if _, err := DB.Exec("DROP TABLE IF EXISTS likes"); err != nil {
		return err
	}
	_, err := DB.Exec("DROP TABLE IF EXISTS dislikes")
	return err
}

// createSessionsTable crée la table des sessions.
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

// GetUserByUsername récupère un utilisateur par son nom d'utilisateur.
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

// GetUserByEmail récupère un utilisateur par son email.
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

// GetUserByID récupère un utilisateur par son ID.
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

// CreateUser crée un nouvel utilisateur.
func CreateUser(user *User) error {
	query := "INSERT INTO users (id, username, email, password) VALUES (?, ?, ?, ?)"

	_, err := DB.Exec(query, user.ID, user.Username, user.Email, user.Password)
	if err != nil {
		return fmt.Errorf("erreur création utilisateur: %w", err)
	}

	return nil
}

// CreateSession enregistre une session côté serveur.
func CreateSession(session *Session) error {
	query := "INSERT INTO sessions (id, user_id, token, expires_at) VALUES (?, ?, ?, ?)"

	_, err := DB.Exec(query, session.ID, session.UserID, session.Token, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("erreur création session: %w", err)
	}

	return nil
}

// GetSessionByToken ignore les sessions expirées.
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

// DeleteSession supprime une session par token.
func DeleteSession(token string) error {
	query := "DELETE FROM sessions WHERE token = ?"

	_, err := DB.Exec(query, token)
	if err != nil {
		return fmt.Errorf("erreur suppression session: %w", err)
	}

	return nil
}

// CreateCategory crée une nouvelle catégorie.
func CreateCategory(category *Category) error {
	query := "INSERT INTO categories (id, name, description) VALUES (?, ?, ?)"

	_, err := DB.Exec(query, category.ID, category.Name, category.Description)
	if err != nil {
		return fmt.Errorf("erreur création catégorie: %w", err)
	}

	return nil
}

// GetAllCategories récupère toutes les catégories.
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

// GetCategoryByID récupère une catégorie par son ID.
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

func scanTopic(scanner interface {
	Scan(dest ...any) error
}) (Topic, error) {
	var topic Topic
	err := scanner.Scan(
		&topic.ID, &topic.CategoryID, &topic.UserID, &topic.Title, &topic.Content,
		&topic.CreatedAt, &topic.UpdatedAt, &topic.Username, &topic.CategoryName,
		&topic.CommentCount, &topic.Likes, &topic.Dislikes,
	)
	return topic, err
}

func queryTopics(where, orderBy, limit string, args ...any) ([]Topic, error) {
	// Les colonnes communes évitent de dupliquer les compteurs de commentaires et de votes.
	query := "SELECT " + topicSelectColumns + topicJoins
	if strings.TrimSpace(where) != "" {
		query += " WHERE " + where
	}
	if strings.TrimSpace(orderBy) != "" {
		query += " ORDER BY " + orderBy
	}
	if strings.TrimSpace(limit) != "" {
		query += " LIMIT " + limit
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("erreur requête topics: %w", err)
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		topic, err := scanTopic(rows)
		if err != nil {
			return nil, fmt.Errorf("erreur scan topic: %w", err)
		}
		topics = append(topics, topic)
	}

	return topics, rows.Err()
}

// CreateTopic crée un nouveau sujet.
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

// GetAllTopics récupère tous les sujets avec les infos d'auteur.
func GetAllTopics() ([]Topic, error) {
	return queryTopics("", "t.updated_at DESC", "")
}

// GetLatestTopics récupère les derniers sujets avec leur nombre de commentaires.
func GetLatestTopics(limit int) ([]Topic, error) {
	return queryTopics("", "t.created_at DESC", "?", limit)
}

// GetTopicsByCategory récupère tous les sujets d'une catégorie.
func GetTopicsByCategory(categoryID string) ([]Topic, error) {
	return queryTopics("t.category_id = ?", "t.updated_at DESC", "", categoryID)
}

// GetTopicsByUser récupère les sujets créés par un utilisateur.
func GetTopicsByUser(userID string) ([]Topic, error) {
	return queryTopics("t.user_id = ?", "t.created_at DESC", "", userID)
}

// GetLikedTopicsByUser récupère les sujets likés par un utilisateur.
func GetLikedTopicsByUser(userID string) ([]Topic, error) {
	query := "SELECT " + topicSelectColumns + topicJoins + `
		JOIN votes uv ON uv.target_type = 'topic'
			AND uv.target_id = t.id
			AND uv.user_id = ?
			AND uv.vote_type = 'like'
		ORDER BY uv.updated_at DESC
	`

	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur requête topics likés: %w", err)
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		topic, err := scanTopic(rows)
		if err != nil {
			return nil, fmt.Errorf("erreur scan topic liké: %w", err)
		}
		topics = append(topics, topic)
	}

	return topics, rows.Err()
}

// GetLatestTopicsByCategory récupère les derniers sujets d'une catégorie.
func GetLatestTopicsByCategory(categoryID string, limit int) ([]Topic, error) {
	return queryTopics("t.category_id = ?", "t.created_at DESC", "?", categoryID, limit)
}

// GetTopicByID récupère un sujet par son ID.
func GetTopicByID(id string) (*Topic, error) {
	query := "SELECT " + topicSelectColumns + topicJoins + " WHERE t.id = ?"
	topic, err := scanTopic(DB.QueryRow(query, id))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur requête topic: %w", err)
	}

	return &topic, nil
}

func scanComment(scanner interface {
	Scan(dest ...any) error
}) (Comment, error) {
	var comment Comment
	err := scanner.Scan(
		&comment.ID, &comment.TopicID, &comment.UserID, &comment.Content,
		&comment.CreatedAt, &comment.UpdatedAt, &comment.Username,
		&comment.Likes, &comment.Dislikes,
	)
	return comment, err
}

// CreateComment crée un nouveau commentaire.
func CreateComment(comment *Comment) error {
	query := "INSERT INTO comments (id, topic_id, user_id, content) VALUES (?, ?, ?, ?)"

	_, err := DB.Exec(query, comment.ID, comment.TopicID, comment.UserID, comment.Content)
	if err != nil {
		return fmt.Errorf("erreur création commentaire: %w", err)
	}

	return nil
}

// GetCommentsByTopic récupère tous les commentaires d'un sujet.
func GetCommentsByTopic(topicID string) ([]Comment, error) {
	query := "SELECT " + commentSelectColumns + `
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
		comment, err := scanComment(rows)
		if err != nil {
			return nil, fmt.Errorf("erreur scan commentaire: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

// GetCommentByID récupère un commentaire par son ID.
func GetCommentByID(id string) (*Comment, error) {
	query := "SELECT " + commentSelectColumns + `
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?
	`

	comment, err := scanComment(DB.QueryRow(query, id))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur requête commentaire: %w", err)
	}

	return &comment, nil
}

func countVotes(targetType, targetID, voteType string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM votes WHERE target_type = ? AND target_id = ? AND vote_type = ?"

	err := DB.QueryRow(query, targetType, targetID, voteType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage votes: %w", err)
	}

	return count, nil
}

// UserLikedTopic vérifie si un utilisateur a liké un sujet.
func UserLikedTopic(userID, topicID string) (bool, error) {
	return userHasVote(userID, "topic", topicID, "like")
}

// UserLikedComment vérifie si un utilisateur a liké un commentaire.
func UserLikedComment(userID, commentID string) (bool, error) {
	return userHasVote(userID, "comment", commentID, "like")
}

func userHasVote(userID, targetType, targetID, voteType string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM votes WHERE user_id = ? AND target_type = ? AND target_id = ? AND vote_type = ?"

	err := DB.QueryRow(query, userID, targetType, targetID, voteType).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("erreur vérification vote: %w", err)
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

// SetVote crée le vote ou le remplace si l'utilisateur avait déjà voté.
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
