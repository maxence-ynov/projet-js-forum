# Forum Golang - Guide de démarrage

## 📋 Description

Projet de forum simple en **Golang** sans framework, utilisant :
- **net/http** - Serveur HTTP natif
- **html/template** - Rendu des templates
- **SQLite** - Base de données
- **bcrypt** - Hashage des mots de passe
- **uuid** - Génération d'identifiants uniques
- **Cookies de session** - Gestion des sessions utilisateur

## 🏗️ Architecture du projet

```
projet-js-forum/
├── main.go                 # Point d'entrée principal
├── go.mod                  # Dépendances du projet
├── forum.db               # Base de données SQLite (généré)
│
├── /internal              # Code interne du projet
│   ├── models.go         # Structures de données (User, Post, Reply, etc.)
│   ├── database.go       # Fonctions d'accès à la base de données
│   ├── auth.go           # Authentification et gestion des sessions
│   └── handlers.go       # Gestionnaires HTTP (routes)
│
├── /templates            # Templates HTML
│   ├── layout.html      # Layout principal (header, footer)
│   ├── index.html       # Page d'accueil
│   ├── login.html       # Page de connexion
│   ├── register.html    # Page d'inscription
│   └── forum.html       # Page du forum
│
└── /static              # Fichiers statiques (CSS, JS, images)
```

## 🚀 Installation et lancement

### 1. Installer les dépendances

```bash
cd c:\Users\Pirat\Documents\GitHub\projet-js-forum
go mod download
```

### 2. Lancer le serveur

```bash
go run main.go
```

Le serveur démarre sur **http://localhost:8080**

### 3. Accéder au forum

Ouvrez votre navigateur et allez à : `http://localhost:8080`

## 📚 Fonctionnalités

### ✅ Authentification
- 📝 Inscription avec validation
- 🔐 Connexion sécurisée (bcrypt)
- 🍪 Sessions avec cookies
- 🔓 Déconnexion

### 💬 Forum
- 📌 Créer des posts (sujets)
- 💭 Ajouter des réponses aux posts
- 👤 Voir le nom d'auteur de chaque post/réponse
- 📅 Dates de publication

### 🔒 Sécurité
- Mots de passe hashés avec bcrypt
- Sessions sécurisées avec tokens
- Middleware d'authentification
- Protection CSRF via POST

## 📝 Routes principales

| Méthode | Route | Description |
|---------|-------|-------------|
| GET | `/` | Accueil |
| GET | `/register` | Page d'inscription |
| POST | `/register` | Traiter l'inscription |
| GET | `/login` | Page de connexion |
| POST | `/login` | Traiter la connexion |
| GET | `/logout` | Déconnexion |
| GET | `/forum` | Consulter le forum (protégé) |
| POST | `/forum/post` | Créer un post (protégé) |
| POST | `/forum/post/{id}/reply` | Ajouter une réponse (protégé) |

## 🗄️ Base de données

La base de données SQLite contient 4 tables :

### `users`
```sql
- id (TEXT PRIMARY KEY)
- username (TEXT UNIQUE)
- email (TEXT UNIQUE)
- password (TEXT) -- hash bcrypt
- created_at (DATETIME)
```

### `sessions`
```sql
- id (TEXT PRIMARY KEY)
- user_id (TEXT FOREIGN KEY)
- token (TEXT UNIQUE)
- created_at (DATETIME)
- expires_at (DATETIME)
```

### `posts`
```sql
- id (TEXT PRIMARY KEY)
- user_id (TEXT FOREIGN KEY)
- title (TEXT)
- content (TEXT)
- created_at (DATETIME)
```

### `replies`
```sql
- id (TEXT PRIMARY KEY)
- post_id (TEXT FOREIGN KEY)
- user_id (TEXT FOREIGN KEY)
- content (TEXT)
- created_at (DATETIME)
```

## 🔧 Code exemple

### Créer un utilisateur
```go
user := &User{
    ID:       uuid.New().String(),
    Username: "alice",
    Email:    "alice@example.com",
    Password: hashedPassword, // hashage avec bcrypt
}
err := CreateUser(user)
```

### Vérifier un mot de passe
```go
isValid := VerifyPassword(hashedPassword, plainPassword)
```

### Récupérer l'utilisateur connecté
```go
user := GetUserFromRequest(r)
if user != nil {
    // Utilisateur authentifié
}
```

### Créer un post
```go
post := &Post{
    ID:      uuid.New().String(),
    UserID:  user.ID,
    Title:   "Mon premier post",
    Content: "Contenu du post...",
}
err := CreatePost(post)
```

## 📖 Structure du code

### `main.go`
- Initialisation de la base de données
- Configuration des routes
- Lancement du serveur

### `internal/models.go`
- Structures de données (User, Post, Reply, Session)
- Struct PageData pour les templates

### `internal/database.go`
- Initialisation et création des tables
- Fonctions CRUD pour tous les modèles
- Requêtes SQL préparées

### `internal/auth.go`
- Hashage bcrypt
- Génération de tokens
- Gestion des sessions
- Middleware d'authentification

### `internal/handlers.go`
- Gestionnaires pour chaque route
- Rendu des templates
- Validation des données
- Redirection post-action

## 🎯 Bonnes pratiques appliquées

✅ Code lisible et commenté  
✅ Séparation des responsabilités (models, database, handlers)  
✅ Pas de framework complexe  
✅ Gestion d'erreurs simple mais complète  
✅ Validation des données  
✅ Sécurité de base (bcrypt, sessions)  
✅ Templates HTML propres et stylisés  

## 💡 Améliorations possibles

- Pagination des posts
- Édition/suppression de posts
- Système de réactions (likes)
- Recherche de posts
- Catégories/tags
- Notifications
- Système d'administration
- Tests unitaires
- Documentation API

## ❓ Dépannage

### Base de données vide
```bash
# Supprimer et recréer la base de données
rm forum.db
go run main.go
```

### Erreur de module
```bash
# Télécharger les dépendances
go mod download
go mod tidy
```

### Port 8080 déjà utilisé
Modifier dans `main.go` :
```go
adresse := "localhost:3000"  // ou un autre port
```

## 🎓 Pour les étudiants

Ce projet est une excellente base pour apprendre :
- La programmation web en Go
- SQLite et les bases de données
- La gestion des sessions
- La sécurité web (hachage, HTTPS, etc.)
- La structuration d'un projet Go

N'hésitez pas à l'étendre et l'améliorer ! 🚀

---

**Auteur** : Projet étudiant en Golang  
**Date** : 2024  
**Licence** : Libre d'utilisation à titre éducatif
