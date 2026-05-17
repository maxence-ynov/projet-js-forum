# 📊 Structure de la Base de Données

## Vue d'ensemble

Le forum utilise **SQLite** avec 6 tables organisées selon une architecture relationnelle propre.

## 📋 Tables et Schéma

### 1️⃣ **users** - Utilisateurs
Stocke les informations de tous les utilisateurs du forum.

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY, Unique | Identifiant unique (UUID) |
| `username` | TEXT | UNIQUE, NOT NULL | Nom d'utilisateur |
| `email` | TEXT | UNIQUE, NOT NULL | Adresse email |
| `password` | TEXT | NOT NULL | Hash bcrypt du mot de passe |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Date de création |

**Clés primaires** : `id`  
**Index uniques** : `username`, `email`

---

### 2️⃣ **categories** - Catégories
Organise les topics en catégories (ex: "Programmation", "Débats", etc).

```sql
CREATE TABLE categories (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY, Unique | Identifiant unique (UUID) |
| `name` | TEXT | UNIQUE, NOT NULL | Nom de la catégorie |
| `description` | TEXT | - | Description (optionnel) |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Date de création |

**Clés primaires** : `id`  
**Index uniques** : `name`

---

### 3️⃣ **topics** - Sujets/Posts
Les sujets créés par les utilisateurs, organisés par catégorie.

```sql
CREATE TABLE topics (
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
```

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY, Unique | Identifiant unique (UUID) |
| `category_id` | TEXT | NOT NULL, FK → categories | Catégorie du topic |
| `user_id` | TEXT | NOT NULL, FK → users | Créateur du topic |
| `title` | TEXT | NOT NULL | Titre du topic |
| `content` | TEXT | NOT NULL | Contenu du topic |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Date de création |
| `updated_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Dernière modification |

**Clés primaires** : `id`  
**Clés étrangères** :
- `category_id` → `categories(id)` (CASCADE DELETE)
- `user_id` → `users(id)` (CASCADE DELETE)

**Relations** :
- 1 category : N topics
- 1 user : N topics

---

### 4️⃣ **comments** - Commentaires
Les réponses/commentaires des utilisateurs sur les topics.

```sql
CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    topic_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY, Unique | Identifiant unique (UUID) |
| `topic_id` | TEXT | NOT NULL, FK → topics | Topic concerné |
| `user_id` | TEXT | NOT NULL, FK → users | Auteur du commentaire |
| `content` | TEXT | NOT NULL | Contenu du commentaire |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Date de création |
| `updated_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Dernière modification |

**Clés primaires** : `id`  
**Clés étrangères** :
- `topic_id` → `topics(id)` (CASCADE DELETE)
- `user_id` → `users(id)` (CASCADE DELETE)

**Relations** :
- 1 topic : N comments
- 1 user : N comments

---

### 5️⃣ **likes** - Likes/Aimes
Gère les likes pour topics ET commentaires.

```sql
CREATE TABLE likes (
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
```

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY, Unique | Identifiant unique (UUID) |
| `user_id` | TEXT | NOT NULL, FK → users | Utilisateur qui like |
| `topic_id` | TEXT | FK → topics (NULL ok) | Topic liké (NULL si comment) |
| `comment_id` | TEXT | FK → comments (NULL ok) | Commentaire liké (NULL si topic) |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Date du like |

**Clés primaires** : `id`  
**Clés étrangères** :
- `user_id` → `users(id)` (CASCADE DELETE)
- `topic_id` → `topics(id)` (CASCADE DELETE)
- `comment_id` → `comments(id)` (CASCADE DELETE)

**Contrainte UNIQUE** : `(user_id, topic_id, comment_id)`  
⚠️ Assure qu'un utilisateur ne peut liker qu'une seule fois le même objet

**Relations** :
- 1 user : N likes
- 1 topic : N likes (optionnel)
- 1 comment : N likes (optionnel)

---

### 6️⃣ **sessions** - Sessions Utilisateur
Gère l'authentification et les sessions actives.

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

| Colonne | Type | Contraintes | Description |
|---------|------|-------------|-------------|
| `id` | TEXT | PRIMARY KEY, Unique | Identifiant unique (UUID) |
| `user_id` | TEXT | NOT NULL, FK → users | Utilisateur connecté |
| `token` | TEXT | UNIQUE, NOT NULL | Token de session sécurisé |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Date de création |
| `expires_at` | DATETIME | NOT NULL | Date d'expiration |

**Clés primaires** : `id`  
**Index uniques** : `token`  
**Clés étrangères** :
- `user_id` → `users(id)` (CASCADE DELETE)

**Relations** :
- 1 user : N sessions

---

## 🔗 Diagramme des Relations

```
┌─────────────┐
│   users     │
│  (base)     │
└──────┬──────┘
       │ 1:N (creator of topics/comments/likes)
       ├────────────┬──────────────┬──────────┐
       │            │              │          │
   ┌───▼────────┐ ┌─▼────────┐ ┌─▼──────┐ ┌─▼──────┐
   │  sessions  │ │ categories │ │topics  │ │comments│
   │            │ │            │ │        │ │        │
   └────────────┘ └────┬───────┘ └─┬──────┘ └─┬──────┘
                       │           │         │
                       │ 1:N      │ 1:N    │ 1:N
                       └───────────┼────────┘
                                   │
                              ┌────▼──────┐
                              │   likes    │
                              │(top/com)   │
                              └────────────┘

Dépendances CASCADE DELETE :
- Supprimer un user → Supprime ses topics, comments, likes, sessions
- Supprimer une category → Supprime ses topics et les likes/comments associés
- Supprimer un topic → Supprime ses comments et likes
- Supprimer un comment → Supprime ses likes
```

---

## 🔍 Requêtes Courantes

### Récupérer tous les topics avec infos auteur et catégorie
```sql
SELECT t.id, t.title, t.content, t.created_at, 
       u.username, c.name as category_name
FROM topics t
JOIN users u ON t.user_id = u.id
JOIN categories c ON t.category_id = c.id
ORDER BY t.updated_at DESC;
```

### Compter les comments par topic
```sql
SELECT topic_id, COUNT(*) as comment_count
FROM comments
GROUP BY topic_id;
```

### Récupérer les topics les plus likés
```sql
SELECT t.id, t.title, COUNT(l.id) as like_count
FROM topics t
LEFT JOIN likes l ON t.id = l.topic_id AND l.comment_id IS NULL
GROUP BY t.id
ORDER BY like_count DESC;
```

### Vérifier si un user a liké un topic
```sql
SELECT COUNT(*) > 0 as has_liked
FROM likes
WHERE user_id = ? AND topic_id = ? AND comment_id IS NULL;
```

---

## 💾 Fonction d'Initialisation

Le fichier `database.go` gère automatiquement :

✅ Connexion à `forum.db`  
✅ Création des tables si elles n'existent pas  
✅ Vérification des contraintes  
✅ Relations avec CASCADE DELETE  

```go
// Exemple
err := InitDB("forum.db")
// Toutes les tables sont créées automatiquement
```

---

## 🛡️ Sécurité

- ✅ Contraintes UNIQUE pour éviter les doublons
- ✅ Clés étrangères avec CASCADE DELETE
- ✅ Constraints UNIQUE sur les likes (pas de duplicatas)
- ✅ Tokens de session uniques et expirables
- ✅ Mots de passe hashés en bcrypt

---

## 📊 Types SQL Utilisés

| Type SQL | Utilisation | Exemple |
|----------|-------------|---------|
| `TEXT` | IDs (UUID), noms, contenu | `id`, `username`, `content` |
| `DATETIME` | Timestamps | `created_at`, `updated_at` |
| `UNIQUE` | Contrainte unicité | username, email, token |
| `PRIMARY KEY` | Clé primaire | `id TEXT PRIMARY KEY` |
| `FOREIGN KEY` | Référence autre table | `REFERENCES users(id)` |
| `ON DELETE CASCADE` | Suppression en cascade | Supprime orphelins auto |
| `DEFAULT` | Valeur par défaut | `DEFAULT CURRENT_TIMESTAMP` |
| `NOT NULL` | Champ obligatoire | `NOT NULL` |

---

## 🚀 Performance

Pour optimiser les requêtes fréquentes, créer des index sur :

```sql
-- Recherche de topics par catégorie
CREATE INDEX IF NOT EXISTS idx_topics_category ON topics(category_id);

-- Recherche de comments par topic
CREATE INDEX IF NOT EXISTS idx_comments_topic ON comments(topic_id);

-- Recherche de likes
CREATE INDEX IF NOT EXISTS idx_likes_user ON likes(user_id);
CREATE INDEX IF NOT EXISTS idx_likes_topic ON likes(topic_id);
CREATE INDEX IF NOT EXISTS idx_likes_comment ON likes(comment_id);

-- Recherche de sessions
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
```

(À ajouter optionnellement dans `database.go` si nécessaire)
