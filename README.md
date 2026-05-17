# Forum Go

Application de forum web développée en Go, sans framework externe, avec rendu HTML côté serveur et persistance des données dans SQLite.

Le projet propose une base fonctionnelle de forum communautaire : inscription, connexion, publication de sujets, commentaires, catégories, profil utilisateur et système de votes.

## Présentation

Ce projet est une application web monolithique écrite en Go. Il utilise la bibliothèque standard pour le serveur HTTP, les templates HTML et la gestion des routes, puis SQLite pour stocker les utilisateurs, sessions, catégories, sujets, commentaires et votes.

L'objectif est de fournir un forum simple, lisible et facilement extensible, adapté à un projet étudiant ou à une base d'apprentissage pour le développement web en Go.

## Fonctionnalités

- Inscription et connexion utilisateur
- Hashage sécurisé des mots de passe avec bcrypt
- Sessions utilisateur via cookies HTTP-only
- Page d'accueil avec les derniers sujets
- Forum protégé accessible aux utilisateurs connectés
- Création de sujets dans des catégories
- Consultation d'un sujet avec ses commentaires
- Ajout de commentaires sur les sujets
- Système de likes et dislikes sur les sujets
- Système de likes et dislikes sur les commentaires
- Page profil avec les sujets créés et les sujets aimés
- Filtrage des sujets par catégorie
- Pages d'erreur dédiées pour les erreurs 400, 404 et 500
- Validation des formulaires côté serveur
- Base SQLite initialisée automatiquement au lancement

## Technologies

- Go 1.22 ou supérieur recommandé
- SQLite
- HTML templates Go
- CSS statique
- Serveur HTTP natif Go

## Installation

### Prérequis

- Go 1.22 ou supérieur
- Git
- Un compilateur C compatible avec `github.com/mattn/go-sqlite3`

> `go-sqlite3` utilise CGO. Sur Windows, installez par exemple MinGW-w64 ou un environnement équivalent si la compilation SQLite échoue.

### Cloner le projet

```bash
git clone <url-du-depot>
cd projet-js-forum
```

### Installer les dépendances

```bash
go mod download
```

### Vérifier les modules

```bash
go mod tidy
```

## Lancement

Lancer l'application en local :

```bash
go run main.go
```

Le serveur démarre sur :

```text
http://localhost:8080
```

La base de données `forum.db` est créée automatiquement à la racine du projet lors du premier lancement.

## Structure du projet

```text
projet-js-forum/
├── main.go
├── go.mod
├── go.sum
├── README.md
├── DATABASE_SCHEMA.md
├── DATABASE_SUMMARY.md
├── DATABASE_USAGE.md
├── internal/
│   ├── auth.go
│   ├── database.go
│   ├── errors.go
│   ├── handlers.go
│   ├── models.go
│   └── validation.go
├── templates/
│   ├── error.html
│   ├── forum.html
│   ├── index.html
│   ├── layout.html
│   ├── login.html
│   ├── profile.html
│   ├── register.html
│   └── topic.html
└── static/
    ├── README.md
    └── style.css
```

## Packages utilisés

### Bibliothèque standard

- `net/http` : serveur HTTP, routes et middlewares
- `html/template` : rendu des pages HTML
- `database/sql` : interface SQL générique
- `context` : stockage de l'utilisateur authentifié pendant la requête
- `crypto/rand` et `encoding/hex` : génération de tokens de session
- `net/mail` : validation des adresses email
- `regexp` : validation des champs utilisateur
- `time` : gestion des dates et expiration des sessions
- `log` : journalisation des erreurs serveur

### Dépendances externes

| Package | Rôle |
| --- | --- |
| `github.com/mattn/go-sqlite3` | Driver SQLite pour `database/sql` |
| `golang.org/x/crypto/bcrypt` | Hashage et vérification des mots de passe |
| `github.com/google/uuid` | Génération des identifiants uniques |

## Routes principales

| Méthode | Route | Accès | Description |
| --- | --- | --- | --- |
| `GET` | `/` | Public | Page d'accueil avec les derniers sujets |
| `GET` | `/register` | Public | Formulaire d'inscription |
| `POST` | `/register` | Public | Création d'un compte |
| `GET` | `/login` | Public | Formulaire de connexion |
| `POST` | `/login` | Public | Connexion utilisateur |
| `GET` | `/logout` | Connecté | Déconnexion |
| `GET` | `/forum` | Connecté | Liste des sujets du forum |
| `POST` | `/forum/post` | Connecté | Création d'un sujet |
| `GET` | `/forum/post/{id}` | Connecté | Détail d'un sujet |
| `POST` | `/forum/post/{id}/reply` | Connecté | Ajout d'un commentaire |
| `POST` | `/forum/post/{id}/vote` | Connecté | Vote sur un sujet |
| `POST` | `/forum/post/{id}/comments/{commentID}/vote` | Connecté | Vote sur un commentaire |
| `GET` | `/profile` | Connecté | Profil de l'utilisateur |
| `GET` | `/static/*` | Public | Fichiers statiques |

## Base de données

SQLite est initialisé automatiquement au démarrage. Les tables principales sont :

- `users` : comptes utilisateurs
- `sessions` : sessions actives
- `categories` : catégories du forum
- `topics` : sujets publiés
- `comments` : commentaires des sujets
- `votes` : votes uniques par utilisateur et par cible
- `likes` et `dislikes` : anciennes tables conservées dans le schéma

Les catégories par défaut sont ajoutées automatiquement si elles n'existent pas :

- Général
- Questions
- Annonces

## Captures d'écran

Les images ci-dessous sont des placeholders. Remplacez-les par de vraies captures dans un dossier `docs/screenshots/`.

### Accueil

![Capture d'écran de la page d'accueil](docs/screenshots/home.png)

### Inscription

![Capture d'écran de la page d'inscription](docs/screenshots/register.png)

### Connexion

![Capture d'écran de la page de connexion](docs/screenshots/login.png)

### Forum

![Capture d'écran de la liste des sujets](docs/screenshots/forum.png)

### Sujet

![Capture d'écran d'un sujet avec commentaires](docs/screenshots/topic.png)

### Profil

![Capture d'écran du profil utilisateur](docs/screenshots/profile.png)

## Commandes utiles

Formater le code :

```bash
go fmt ./...
```

Compiler le projet :

```bash
go build ./...
```

Lancer l'application :

```bash
go run main.go
```

Ajouter des données fake pour tester rapidement :

```bash
go run ./cmd/seed
```

Réinitialiser uniquement les données fake puis les recréer :

```bash
go run ./cmd/seed --reset
```

Comptes de test créés : `alice`, `mehdi`, `clara`, `lucas`.
Mot de passe commun : `password123`.

Réinitialiser la base locale :

```bash
rm forum.db
go run main.go
```

Sur PowerShell :

```powershell
Remove-Item .\forum.db
go run main.go
```

## Sécurité

- Les mots de passe sont stockés sous forme de hash bcrypt.
- Les sessions utilisent des tokens aléatoires de 32 octets.
- Les cookies de session sont configurés en `HttpOnly` et `SameSite=Strict`.
- Les formulaires sont limités à 1 Mo.
- Les entrées utilisateur sont validées côté serveur.
- Les identifiants de sujets et commentaires sont validés au format UUID.

## Améliorations possibles

- Ajouter des tests automatisés
- Ajouter la suppression et l'édition des sujets
- Ajouter la suppression et l'édition des commentaires
- Ajouter une pagination
- Ajouter une recherche de sujets
- Ajouter un panneau d'administration
- Ajouter une configuration par variables d'environnement
- Ajouter une protection CSRF complète
- Ajouter un système d'upload d'avatar

## Licence

Projet fourni à des fins pédagogiques. Ajoutez une licence officielle si le projet doit être publié ou distribué.
