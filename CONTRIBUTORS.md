# 🤝 Guide de Contribution

## 🌍 Langue

**Toute la communication du projet doit être en anglais :**

- ✅ Code (variables, fonctions, classes, commentaires)
- ✅ Messages de commit
- ✅ Titres et descriptions des Pull Requests
- ✅ Revues de code et commentaires
- ✅ Documentation technique
- ✅ Issues et discussions

**Exemple :**
```python
# ✅ Bon
def calculate_quiz_score(answers: list) -> int:
    """Calculate the total score for a quiz."""
    return sum(1 for answer in answers if answer.is_correct)

# ❌ Mauvais
def calculer_score_quiz(reponses: list) -> int:
    """Calculer le score total pour un quiz."""
    return sum(1 for reponse in reponses if reponse.est_correcte)
```

## 🔄 Workflow Git

### Branches Principales

Nous utilisons un workflow inspiré de **Git Flow** :

```
main
  └── develop
       ├── feat/xxxx/intitule-feature
       └── fix/xxxx/intitule-bug
```

> **Note :** `xxxx` = les **4 premières lettres du nom de famille** en minuscules

### Structure des Branches

- **`main`** : Code prêt pour la production. Ne jamais commit directement dessus.
- **`develop`** : Branche principale de développement. Base pour toutes les branches de fonctionnalités.
- **Branches feat/fix/etc.** : Créées depuis `develop`, fusionnées dans `develop`.

### Étapes du Workflow

1. **Toujours partir de `develop` :**
   ```bash
   git checkout develop
   git pull origin develop
   ```

2. **Créer votre branche de feature :**
   ```bash
   # Exemple pour Martin
   git checkout -b feat/mart/user-authentication
   
   # Exemple pour Dupont  
   git checkout -b feat/dupo/spotify-integration
   ```

3. **Travailler sur votre feature et commit régulièrement**

4. **Pusher votre branche :**
   ```bash
   git push origin feat/xxxx/intitule-feature
   ```

5. **Créer une Pull Request** pour fusionner dans `develop`

6. **Après review et approbation**, votre PR sera fusionnée

## 🌿 Convention de Nommage des Branches

### Format

```
<type>/<xxxx>/<description>
```

Où `xxxx` = **4 premières lettres du nom de famille** en minuscules

### Types de Branches

- **`feat/`** : Nouvelle fonctionnalité
- **`fix/`** : Correction de bug
- **`docs/`** : Modifications de documentation

### Exemples

```bash
# ✅ Bons exemples
feat/mart/user-authentication
feat/dupo/spotify-integration
fix/bern/quiz-score-calculation
docs/dura/api-documentation

# ❌ Mauvais exemples
feat/martin/user-auth           # Nom complet au lieu de 4 lettres
feature/mart/new-stuff          # "feature" au lieu de "feat"
mart/fix                        # Manque le préfixe de type
fix-bug                         # Manque l'identifiant
feat/MART/Feature               # Majuscules (doit être en minuscules)
feat/mart/add_new_quiz          # Underscores (doit être des tirets)
```

### Règles

- Utiliser uniquement des **minuscules**
- Utiliser exactement les **4 premières lettres du nom de famille**
- Utiliser des **tirets** (`-`) pour séparer les mots, pas d'underscores ou d'espaces
- Garder un nom **court mais descriptif**
- Utiliser des descriptions en **anglais**

## 💬 Convention des Messages de Commit

Libre à vous d'utiliser ce que vous voulez tant qu'il est en anglais.

## 📝 Guide des Pull Requests

### Avant de Créer une PR

1. **Assurez-vous que votre code fonctionne**
   ```bash
   # Tester localement
   docker-compose up --build
   ```

2. **Vérifiez que votre branche est à jour avec develop**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout feat/xxxx/votreFeature
   git rebase develop
   ```

3. **Vérifiez vos commits**
   - Commits logiques et atomiques
   - Messages de commit clairs et descriptifs

### Créer une Pull Request

1. **Titre en anglais, clair et descriptif**
   ```
   ✅ Add user authentication with JWT tokens
   ✅ Fix quiz score calculation for multiple attempts
   ❌ Fix
   ❌ Modifications
   ```

2. **Labels appropriés**
   - `feature` : Nouvelle fonctionnalité
   - `fix` : Correction de bug
   - `documentation` : Documentation

---

**Merci de contribuer au projet Quiz Musical ! 🎵**