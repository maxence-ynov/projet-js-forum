#cc
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

Pour Docker, supprimez `./data/forum.db` et relancez `docker compose up` si nécessaire.

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

sal