# 🗄️ Guide d'Utilisation du Database.go

## 📖 Documentation complète des fonctions

Ce fichier contient toutes les fonctions pour gérer SQLite sans ORM.

---

## 🔌 Initialisation

### `InitDB(nomFichier string) error`
Ouvre la base de données et crée toutes les tables automatiquement.

```go
// Appeler une seule fois au démarrage
err := InitDB("forum.db")
if err != nil {
    log.Fatal(err)
}
defer CloseDB()
```

### `CloseDB() error`
Ferme la connexion à la base de données.

```go
defer CloseDB()  // À la fin du main()
```

---

## 👥 Gestion des Utilisateurs

### `CreateUser(user *User) error`
Crée un nouvel utilisateur.

```go
newUser := &User{
    ID:       uuid.New().String(),
    Username: "alice",
    Email:    "alice@example.com",
    Password: hashedPassword, // Utiliser HashPassword()
}
err := CreateUser(newUser)
```

### `GetUserByUsername(username string) (*User, error)`
Récupère un utilisateur par son nom.

```go
user, err := GetUserByUsername("alice")
if user == nil {
    // Utilisateur inexistant
}
```

### `GetUserByID(id string) (*User, error)`
Récupère un utilisateur par son ID.

```go
user, err := GetUserByID("550e8400-e29b-41d4-a716-446655440000")
if user != nil {
    fmt.Println(user.Username)
}
```

---

## 🔐 Gestion des Sessions

### `CreateSession(session *Session) error`
Crée une nouvelle session utilisateur.

```go
session := &Session{
    ID:        uuid.New().String(),
    UserID:    user.ID,
    Token:     token, // Token généré aléatoirement
    ExpiresAt: time.Now().Add(24 * time.Hour),
}
err := CreateSession(session)
```

### `GetSessionByToken(token string) (*Session, error)`
Récupère une session par son token (si elle n'a pas expiré).

```go
session, err := GetSessionByToken(cookieToken)
if session != nil {
    // Session valide et non expirée
    user, _ := GetUserByID(session.UserID)
}
```

### `DeleteSession(token string) error`
Supprime une session (déconnexion).

```go
err := DeleteSession(token)
```

---

## 📂 Gestion des Catégories

### `CreateCategory(category *Category) error`
Crée une nouvelle catégorie.

```go
category := &Category{
    ID:          uuid.New().String(),
    Name:        "Programmation",
    Description: "Discussions sur la programmation",
}
err := CreateCategory(category)
```

### `GetAllCategories() ([]Category, error)`
Récupère toutes les catégories.

```go
categories, err := GetAllCategories()
for _, cat := range categories {
    fmt.Printf("%s: %s\n", cat.Name, cat.Description)
}
```

### `GetCategoryByID(id string) (*Category, error)`
Récupère une catégorie par son ID.

```go
category, err := GetCategoryByID(categoryID)
```

---

## 📌 Gestion des Topics (Sujets)

### `CreateTopic(topic *Topic) error`
Crée un nouveau topic.

```go
topic := &Topic{
    ID:         uuid.New().String(),
    CategoryID: categoryID,
    UserID:     user.ID,
    Title:      "Comment utiliser Go ?",
    Content:    "Je veux apprendre Go...",
}
err := CreateTopic(topic)
```

### `GetAllTopics() ([]Topic, error)`
Récupère tous les topics avec infos auteur et catégorie.

```go
topics, err := GetAllTopics()
for _, topic := range topics {
    fmt.Printf("[%s] %s - Par %s\n", 
        topic.CategoryName, topic.Title, topic.Username)
}
```

### `GetTopicsByCategory(categoryID string) ([]Topic, error)`
Récupère les topics d'une catégorie.

```go
topics, err := GetTopicsByCategory(categoryID)
```

### `GetTopicByID(id string) (*Topic, error)`
Récupère un topic par son ID.

```go
topic, err := GetTopicByID(topicID)
if topic != nil {
    fmt.Println(topic.Title)
}
```

---

## 💬 Gestion des Commentaires

### `CreateComment(comment *Comment) error`
Ajoute un commentaire à un topic.

```go
comment := &Comment{
    ID:      uuid.New().String(),
    TopicID: topicID,
    UserID:  user.ID,
    Content: "Excellente question !",
}
err := CreateComment(comment)
```

### `GetCommentsByTopic(topicID string) ([]Comment, error)`
Récupère tous les commentaires d'un topic.

```go
comments, err := GetCommentsByTopic(topicID)
for _, comment := range comments {
    fmt.Printf("%s: %s\n", comment.Username, comment.Content)
}
```

### `GetCommentByID(id string) (*Comment, error)`
Récupère un commentaire par son ID.

```go
comment, err := GetCommentByID(commentID)
```

---

## 👍 Gestion des Likes

### `CreateLike(like *Like) error`
Crée un like pour un topic ou commentaire.

```go
// Like sur un topic
like := &Like{
    ID:     uuid.New().String(),
    UserID: user.ID,
    TopicID: &topicID,
    // CommentID reste nil
}
err := CreateLike(like)

// Ou like sur un commentaire
like2 := &Like{
    ID:        uuid.New().String(),
    UserID:    user.ID,
    CommentID: &commentID,
    // TopicID reste nil
}
err = CreateLike(like2)
```

### `DeleteLike(id string) error`
Supprime un like (unlike).

```go
err := DeleteLike(likeID)
```

### `GetLikesByTopic(topicID string) (int, error)`
Compte les likes d'un topic.

```go
count, err := GetLikesByTopic(topicID)
fmt.Printf("Ce topic a %d likes\n", count)
```

### `GetLikesByComment(commentID string) (int, error)`
Compte les likes d'un commentaire.

```go
count, err := GetLikesByComment(commentID)
```

### `UserLikedTopic(userID, topicID string) (bool, error)`
Vérifie si un utilisateur a liké un topic.

```go
liked, err := UserLikedTopic(user.ID, topicID)
if liked {
    fmt.Println("Vous avez aimé ce topic")
}
```

### `UserLikedComment(userID, commentID string) (bool, error)`
Vérifie si un utilisateur a liké un commentaire.

```go
liked, err := UserLikedComment(user.ID, commentID)
```

---

## 🛠️ Exemple Complet d'Utilisation

```go
package main

import (
	"fmt"
	"log"
	"forum/internal"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. Initialiser la BD
	err := internal.InitDB("forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer internal.CloseDB()

	// 2. Créer un utilisateur
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	newUser := &internal.User{
		ID:       uuid.New().String(),
		Username: "alice",
		Email:    "alice@example.com",
		Password: string(hashedPwd),
	}
	internal.CreateUser(newUser)
	fmt.Println("✓ Utilisateur créé")

	// 3. Créer une catégorie
	category := &internal.Category{
		ID:          uuid.New().String(),
		Name:        "Programmation",
		Description: "Discussions sur le code",
	}
	internal.CreateCategory(category)
	fmt.Println("✓ Catégorie créée")

	// 4. Créer un topic
	topic := &internal.Topic{
		ID:         uuid.New().String(),
		CategoryID: category.ID,
		UserID:     newUser.ID,
		Title:      "Comment débuter en Golang ?",
		Content:    "Je suis intéressé par Go...",
	}
	internal.CreateTopic(topic)
	fmt.Println("✓ Topic créé")

	// 5. Ajouter un commentaire
	comment := &internal.Comment{
		ID:      uuid.New().String(),
		TopicID: topic.ID,
		UserID:  newUser.ID,
		Content: "Go est un excellent langage !",
	}
	internal.CreateComment(comment)
	fmt.Println("✓ Commentaire ajouté")

	// 6. Ajouter un like
	like := &internal.Like{
		ID:      uuid.New().String(),
		UserID:  newUser.ID,
		TopicID: &topic.ID,
	}
	internal.CreateLike(like)
	fmt.Println("✓ Like ajouté")

	// 7. Récupérer les données
	allTopics, _ := internal.GetAllTopics()
	fmt.Printf("\n📌 Topics: %d\n", len(allTopics))
	for _, t := range allTopics {
		fmt.Printf("  - %s (par %s)\n", t.Title, t.Username)
	}

	comments, _ := internal.GetCommentsByTopic(topic.ID)
	fmt.Printf("\n💬 Commentaires du topic: %d\n", len(comments))

	likeCount, _ := internal.GetLikesByTopic(topic.ID)
	fmt.Printf("\n👍 Likes du topic: %d\n", likeCount)
}
```

---

## ⚠️ Gestion des Erreurs

Toutes les fonctions retournent une `error`. Toujours les vérifier :

```go
user, err := GetUserByUsername("alice")
if err != nil {
    log.Printf("Erreur BD: %v", err)
    return
}

if user == nil {
    fmt.Println("Utilisateur non trouvé")
    return
}

// Utiliser user...
```

---

## 🔗 Relations CASCADE DELETE

Quand vous supprimez des données :

- Supprimer un `user` → Supprime ses topics, commentaires, likes, sessions
- Supprimer une `category` → Supprime ses topics et leurs commentaires/likes
- Supprimer un `topic` → Supprime ses commentaires et likes
- Supprimer un `comment` → Supprime ses likes

Cela maintient l'intégrité de la BD automatiquement !

---

## 📊 Verrous et Concurrence

SQLite fonctionne bien avec une concurrence modérée. Pour plus de détails, voir le schéma complet dans `DATABASE_SCHEMA.md`.
