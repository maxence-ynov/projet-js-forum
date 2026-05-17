package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"forum/internal"

	"github.com/google/uuid"
)

const (
	seedPrefix = "forum-seed"
	password   = "password123"
)

type seedUser struct {
	ID       string
	Username string
	Email    string
}

type seedCategory struct {
	ID          string
	Name        string
	Description string
}

type seedTopic struct {
	ID         string
	CategoryID string
	UserID     string
	Title      string
	Content    string
	CreatedAt  time.Time
}

type seedComment struct {
	ID        string
	TopicID   string
	UserID    string
	Content   string
	CreatedAt time.Time
}

type seedVote struct {
	UserID     string
	TargetType string
	TargetID   string
	VoteType   string
}

func main() {
	dbPath := flag.String("db", "forum.db", "chemin vers la base SQLite")
	reset := flag.Bool("reset", false, "supprime puis recrée les données fake")
	flag.Parse()

	if err := internal.InitDB(*dbPath); err != nil {
		log.Fatalf("Erreur initialisation BD: %v", err)
	}
	defer internal.CloseDB()

	data, err := buildSeedData()
	if err != nil {
		log.Fatalf("Erreur préparation des données: %v", err)
	}

	if *reset {
		if err := resetSeedData(data); err != nil {
			log.Fatalf("Erreur suppression données fake: %v", err)
		}
	}

	if err := insertSeedData(data); err != nil {
		log.Fatalf("Erreur insertion données fake: %v", err)
	}

	fmt.Printf("Données fake prêtes dans %s\n", *dbPath)
	fmt.Printf("- %d utilisateurs fake\n", len(data.users))
	fmt.Printf("- %d catégories\n", len(data.categories))
	fmt.Printf("- %d sujets\n", len(data.topics))
	fmt.Printf("- %d commentaires\n", len(data.comments))
	fmt.Printf("- %d votes\n", len(data.votes))
	fmt.Printf("Connexion de test: alice / %s\n", password)
}

type seedData struct {
	users      []seedUser
	categories []seedCategory
	topics     []seedTopic
	comments   []seedComment
	votes      []seedVote
}

func buildSeedData() (seedData, error) {
	now := time.Now().Add(-72 * time.Hour)

	users := []seedUser{
		{ID: deterministicID("user:alice"), Username: "alice", Email: "alice@example.test"},
		{ID: deterministicID("user:mehdi"), Username: "mehdi", Email: "mehdi@example.test"},
		{ID: deterministicID("user:clara"), Username: "clara", Email: "clara@example.test"},
		{ID: deterministicID("user:lucas"), Username: "lucas", Email: "lucas@example.test"},
	}

	categories := []seedCategory{
		{ID: "dev-web", Name: "Dev web", Description: "HTML, CSS, JavaScript, Go et architecture web."},
		{ID: "entraide", Name: "Entraide", Description: "Questions pratiques et blocages techniques."},
		{ID: "projets", Name: "Projets", Description: "Présentations, idées et retours sur des projets."},
	}

	topics := []seedTopic{
		{
			ID:         deterministicID("topic:routes-go"),
			CategoryID: "dev-web",
			UserID:     users[0].ID,
			Title:      "Organiser les routes HTTP en Go",
			Content:    "Je cherche une façon simple de garder les routes lisibles sans ajouter un framework complet. Vous découpez plutôt par handlers ou par domaine ?",
			CreatedAt:  now.Add(2 * time.Hour),
		},
		{
			ID:         deterministicID("topic:sqlite-tests"),
			CategoryID: "entraide",
			UserID:     users[1].ID,
			Title:      "Préparer une base SQLite pour les tests",
			Content:    "Pour tester rapidement le forum, une base remplie avec des utilisateurs et des sujets serait très pratique. Vous utilisez plutôt un script Go ou un fichier SQL ?",
			CreatedAt:  now.Add(8 * time.Hour),
		},
		{
			ID:         deterministicID("topic:css-forum"),
			CategoryID: "projets",
			UserID:     users[2].ID,
			Title:      "Retour sur le style de la page forum",
			Content:    "J'ai commencé à améliorer la page forum. L'objectif est de rendre les sujets plus faciles à parcourir sur mobile et desktop.",
			CreatedAt:  now.Add(20 * time.Hour),
		},
		{
			ID:         deterministicID("topic:votes"),
			CategoryID: "dev-web",
			UserID:     users[3].ID,
			Title:      "Likes et dislikes sur les commentaires",
			Content:    "Le système de vote fonctionne sur les sujets et les commentaires. Il reste à vérifier que chaque utilisateur ne peut voter qu'une seule fois par cible.",
			CreatedAt:  now.Add(30 * time.Hour),
		},
	}

	comments := []seedComment{
		{ID: deterministicID("comment:routes-go:1"), TopicID: topics[0].ID, UserID: users[1].ID, Content: "Je préfère regrouper les handlers par domaine. Le main reste lisible et chaque fichier garde une responsabilité claire.", CreatedAt: topics[0].CreatedAt.Add(45 * time.Minute)},
		{ID: deterministicID("comment:routes-go:2"), TopicID: topics[0].ID, UserID: users[2].ID, Content: "Avec le routeur standard de Go récent, les méthodes dans les patterns rendent déjà le découpage assez propre.", CreatedAt: topics[0].CreatedAt.Add(2 * time.Hour)},
		{ID: deterministicID("comment:sqlite-tests:1"), TopicID: topics[1].ID, UserID: users[0].ID, Content: "Un script Go a l'avantage de réutiliser le même schéma et le même hashage de mot de passe que l'application.", CreatedAt: topics[1].CreatedAt.Add(30 * time.Minute)},
		{ID: deterministicID("comment:sqlite-tests:2"), TopicID: topics[1].ID, UserID: users[3].ID, Content: "Pense aussi à rendre le script idempotent pour pouvoir le lancer plusieurs fois sans créer de doublons.", CreatedAt: topics[1].CreatedAt.Add(90 * time.Minute)},
		{ID: deterministicID("comment:css-forum:1"), TopicID: topics[2].ID, UserID: users[0].ID, Content: "Les cartes de sujets sont plus faciles à scanner si la catégorie et le nombre de réponses restent visibles dès la liste.", CreatedAt: topics[2].CreatedAt.Add(75 * time.Minute)},
		{ID: deterministicID("comment:votes:1"), TopicID: topics[3].ID, UserID: users[2].ID, Content: "La contrainte unique sur user, target_type et target_id devrait suffire pour éviter les votes multiples.", CreatedAt: topics[3].CreatedAt.Add(40 * time.Minute)},
	}

	votes := []seedVote{
		{UserID: users[1].ID, TargetType: "topic", TargetID: topics[0].ID, VoteType: "like"},
		{UserID: users[2].ID, TargetType: "topic", TargetID: topics[0].ID, VoteType: "like"},
		{UserID: users[0].ID, TargetType: "topic", TargetID: topics[1].ID, VoteType: "like"},
		{UserID: users[3].ID, TargetType: "topic", TargetID: topics[1].ID, VoteType: "like"},
		{UserID: users[0].ID, TargetType: "comment", TargetID: comments[0].ID, VoteType: "like"},
		{UserID: users[2].ID, TargetType: "comment", TargetID: comments[2].ID, VoteType: "like"},
		{UserID: users[1].ID, TargetType: "comment", TargetID: comments[5].ID, VoteType: "like"},
	}

	return seedData{
		users:      users,
		categories: categories,
		topics:     topics,
		comments:   comments,
		votes:      votes,
	}, nil
}

func insertSeedData(data seedData) error {
	tx, err := internal.DB.Begin()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)

	hashedPassword, err := internal.HashPassword(password)
	if err != nil {
		return err
	}

	for _, user := range data.users {
		if _, err := tx.Exec(
			"INSERT OR IGNORE INTO users (id, username, email, password) VALUES (?, ?, ?, ?)",
			user.ID, user.Username, user.Email, hashedPassword,
		); err != nil {
			return fmt.Errorf("utilisateur %s: %w", user.Username, err)
		}
	}

	for _, category := range data.categories {
		if _, err := tx.Exec(
			"INSERT OR IGNORE INTO categories (id, name, description) VALUES (?, ?, ?)",
			category.ID, category.Name, category.Description,
		); err != nil {
			return fmt.Errorf("catégorie %s: %w", category.ID, err)
		}
	}

	for _, topic := range data.topics {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO topics
				(id, category_id, user_id, title, content, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?)`,
			topic.ID, topic.CategoryID, topic.UserID, topic.Title, topic.Content, topic.CreatedAt, topic.CreatedAt,
		); err != nil {
			return fmt.Errorf("sujet %s: %w", topic.Title, err)
		}
	}

	for _, comment := range data.comments {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO comments
				(id, topic_id, user_id, content, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?)`,
			comment.ID, comment.TopicID, comment.UserID, comment.Content, comment.CreatedAt, comment.CreatedAt,
		); err != nil {
			return fmt.Errorf("commentaire %s: %w", comment.ID, err)
		}
	}

	for _, vote := range data.votes {
		if _, err := tx.Exec(
			`INSERT INTO votes (id, user_id, target_type, target_id, vote_type)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT(user_id, target_type, target_id)
				DO UPDATE SET vote_type = excluded.vote_type, updated_at = CURRENT_TIMESTAMP`,
			deterministicID(fmt.Sprintf("vote:%s:%s:%s", vote.UserID, vote.TargetType, vote.TargetID)),
			vote.UserID,
			vote.TargetType,
			vote.TargetID,
			vote.VoteType,
		); err != nil {
			return fmt.Errorf("vote %s/%s: %w", vote.TargetType, vote.TargetID, err)
		}
	}

	return tx.Commit()
}

func resetSeedData(data seedData) error {
	tx, err := internal.DB.Begin()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)

	if err := deleteByIDs(tx, "votes", "target_id", append(topicIDs(data.topics), commentIDs(data.comments)...)); err != nil {
		return err
	}
	if err := deleteByIDs(tx, "votes", "user_id", userIDs(data.users)); err != nil {
		return err
	}
	if err := deleteByIDs(tx, "comments", "id", commentIDs(data.comments)); err != nil {
		return err
	}
	if err := deleteByIDs(tx, "topics", "id", topicIDs(data.topics)); err != nil {
		return err
	}
	if err := deleteByIDs(tx, "categories", "id", categoryIDs(data.categories)); err != nil {
		return err
	}
	if err := deleteByIDs(tx, "sessions", "user_id", userIDs(data.users)); err != nil {
		return err
	}
	if err := deleteByIDs(tx, "users", "id", userIDs(data.users)); err != nil {
		return err
	}

	return tx.Commit()
}

func deleteByIDs(tx *sql.Tx, table, column string, ids []string) error {
	for _, id := range ids {
		if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table, column), id); err != nil {
			return fmt.Errorf("suppression %s.%s=%s: %w", table, column, id, err)
		}
	}
	return nil
}

func deterministicID(name string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(seedPrefix+":"+name)).String()
}

func rollbackUnlessCommitted(tx *sql.Tx) {
	_ = tx.Rollback()
}

func userIDs(users []seedUser) []string {
	ids := make([]string, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}

func categoryIDs(categories []seedCategory) []string {
	ids := make([]string, 0, len(categories))
	for _, category := range categories {
		ids = append(ids, category.ID)
	}
	return ids
}

func topicIDs(topics []seedTopic) []string {
	ids := make([]string, 0, len(topics))
	for _, topic := range topics {
		ids = append(ids, topic.ID)
	}
	return ids
}

func commentIDs(comments []seedComment) []string {
	ids := make([]string, 0, len(comments))
	for _, comment := range comments {
		ids = append(ids, comment.ID)
	}
	return ids
}
