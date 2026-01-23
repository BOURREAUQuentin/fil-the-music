# Projet Quiz Musical - Architecture Micro-services

Application de quiz musical permettant de tester ses connaissances sur la musique à travers une architecture micro-services moderne.

## Présentation du Projet

### Principe

Application de quiz musical où les utilisateurs peuvent tester leurs connaissances musicales à travers différents types de quiz.

### Fonctionnement

Un quiz est composé d'une suite de questions avec une seule bonne réponse par question. L'application propose plusieurs quiz prédéfinis et permet la création de nouveaux quiz.

### Types de Questions

Chaque quiz peut contenir différents types de questions pour varier les défis :

- **Deviner la date** : Retrouver l'année de sortie d'un morceau ou d'un album
- **Deviner le titre de l'album** : Identifier l'album d'origine d'un morceau
- **Deviner l'artiste d'un morceau** : Reconnaître l'interprète d'une chanson
- **Deviner le titre d'un morceau** : Identifier le nom d'une chanson à partir d'informations contextuelles

### Types de Quiz

#### Quiz Simple (Admin uniquement)
- Questions et réponses entièrement personnalisables
- Contenu défini manuellement par l'administrateur
- Pas de vérification automatique des données
- Liberté totale sur le contenu
- Accessible à tous

#### Quiz Par Genre (Utilisateurs)
- Filtrage par genre musical (exemple : "Pop")
- Récupération des données depuis le Catalogue
- Basé sur les données Spotify (artistes, morceaux, dates)
- Création automatisée accessible à tous

> **⚠️ Attention** : Tous les genres ne sont pas testables car certains n'ont pas de données disponibles. Spotify ne renvoie pas systématiquement de résultats pour tous les genres musicaux. Il est recommandé de tester avec des genres populaires comme "pop", "french rap", etc.

#### Quiz "For You" (Utilisateurs)
- Personnalisé selon les préférences de l'utilisateur (lien Spotify)
- Basé sur les artistes favoris de l'utilisateur
- Création automatique adaptée au profil
- Création automatisée accessible à tous

#### Quiz Random (Utilisateurs)
- Questions aléatoires issues des données du Catalogue
- Création automatisée accessible à tous

### Actions Utilisateur

#### Admin
- Créer des quiz simples avec contenu libre
- Créer et gérer les utilisateurs
- Gérer le catalogue musical

#### User
- Répondre à tous types de quiz (simple, personnalisé, "for you")
- Créer des quiz personnalisés basés sur le Catalogue
- Consulter les scores obtenus
- Accéder à l'historique complet des quiz effectués
- Rejouer n'importe quel quiz autant de fois que souhaité

## Architecture Globale

```
Client / Insomnia
       ↓
Gateway REST (TypeScript)
       ↓
   ┌───┴───┬───────┐
   ↓       ↓       ↓
Service  Service  Service
 User   Catalog   Game
(TS)   (Python)   (Go)
   ↑       ↑       ↑
   └───────┴───────┘
      Spotify API
```

L'application suit une architecture micro-services avec 5 services principaux communiquant via différents protocoles (REST, GraphQL, gRPC).

## Services

### 1. Service Ingestion (Python)

**Type**: HTTP GET  
**Rôle**: Récupération des données depuis l'API Spotify

**Fonctionnalités**:
- Connexion à l'API Spotify (OAuth Token)
- Import des données musicales (artistes, morceaux, albums, dates)
- Envoi des données vers le Service Catalogue

### 2. Service Catalogue (Python - GraphQL)

**Type**: GraphQL  
**Base de données**: MongoDB Catalogue

**Fonctionnalités**:
- Récupérer les informations musicales (artistes, musiques, albums, dates, genres)
- Ajouter de nouvelles données au catalogue
- Requêtes GraphQL pour recherche et filtrage avancés
- Stockage persistant des données musicales
- Source de données pour la génération de questions

### 3. Service User (TypeScript - REST)

**Type**: REST API  
**Base de données**: MongoDB Users

**Fonctionnalités**:
- **Gestion des utilisateurs**: Création, modification, suppression
- **Authentification et autorisation**
- **Gestion des rôles**:
    - Admin: création d'utilisateurs, quiz, gestion du catalogue
    - User: jouer aux quiz, consulter statistiques
- **Historique et statistiques**:
    - Sauvegarde de l'historique des quiz
    - Gestion des scores
    - Statistiques de performance

### 4. Service Game Engine (Go - gRPC)

**Type**: gRPC  
**Base de données**: MongoDB Games

**Fonctionnalités**:
- **Gestion des quiz**:
    - Collection Quizz: Stockage de tous les quiz (simples et personnalisés)
    - Questions et réponses avec validation
    - Collection Game: Historique des parties jouées
- **Gameplay**:
    - Génération et affichage des questions
    - Vérification des réponses
    - Calcul des scores en temps réel
    - Sauvegarde des parties
- **Génération automatique**:
    - Création de questions depuis le Catalogue
    - Communication haute performance via gRPC

### 5. Gateway REST (TypeScript)

**Type**: REST API Gateway  
**Rôle**: Point d'entrée unique pour les clients

**Fonctionnalités**:
- Routage des requêtes vers les services appropriés
- Agrégation des réponses de plusieurs services
- Gestion de l'authentification et validation des tokens
- Exposition d'une API REST unifiée

## Bases de Données

L'application utilise **3 bases MongoDB indépendantes**:

### MongoDB Users
- Données utilisateurs (credentials, profils)
- Rôles et permissions
- Historique des quiz effectués
- Scores et statistiques

### MongoDB Catalogue
- Catalogue musical complet (artistes, morceaux, albums)
- Métadonnées (dates, genres, popularité)
- Données importées depuis Spotify

### MongoDB Games
- **Collection Quizz**: Quiz disponibles, questions et réponses
- **Collection Game**: Historique des parties et scores

## Technologies Utilisées

### Langages
- **Python**: Services Ingestion et Catalogue
- **TypeScript**: Services Gateway et User
- **Go**: Service Game Engine

### Frameworks
- **Python**: FastAPI (Ingestion), GraphQL natif (Catalogue)
- **TypeScript**: Express.js (Gateway et User)
- **Go**: gRPC natif (Game Engine)

### Protocoles
- **REST**: Gateway, User, Ingestion
- **GraphQL**: Catalogue
- **gRPC**: Game Engine

### Base de Données
- **MongoDB**: 3 instances dédiées

### API Externe
- **Spotify API**: Source des données musicales (OAuth)

## Installation et Lancement

### Prérequis
- Docker et Docker Compose installés
- Compte développeur Spotify pour obtenir les credentials API

### Configuration des Variables d'Environnement

Avant de lancer l'application, vous devez configurer les variables d'environnement nécessaires.

#### 1. Service Ingestion

Créez un fichier `.env` dans le dossier `ingestion/` :

```bash
# Spotify API Credentials
SPOTIFY_CLIENT_ID=votre_client_id_spotify
SPOTIFY_CLIENT_SECRET=votre_client_secret_spotify

# Service URLs
CATALOG_URL=http://catalog:3200/graphql
```

> **Comment obtenir vos credentials Spotify ?**
> 1. Rendez-vous sur [Spotify for Developers](https://developer.spotify.com/dashboard)
> 2. Connectez-vous avec votre compte Spotify
> 3. Créez une nouvelle application
> 4. Récupérez votre `Client ID` et `Client Secret`

#### 2. Service User

Créez un fichier `.env` dans le dossier `user/` :

```bash
# JWT Configuration
JWT_SECRET=votre_secret_jwt_tres_securise_ici
```

> **⚠️ Sécurité** : Modifiez impérativement `JWT_SECRET` avec une chaîne aléatoire sécurisée en production.

#### 3. Gateway

Créez un fichier `.env` dans le dossier `gateway/` :

```bash
# Gateway Port
PORT=8080
```

### Installation

```bash
# Cloner le repository
git clone https://github.com/BOURREAUQuentin/fil-the-music.git
cd fil-the-music

# Configurer les variables d'environnement
# Suivez les instructions de configuration ci-dessus

# Lancer l'environnement Docker
docker-compose up --build -d
```

### Accès

L'API est accessible sur `http://localhost:8080`

## Tests

Des exports Insomnia sont mis à disposition pour faciliter le test des endpoints :
- **Global :** Un export complet est disponible à la racine du projet `Insomnia_global.yml`.
- **Spécifique :** Plusieurs dossiers de services (Catalog, Game, User) contiennent leur propre export dédié.

### Configuration spécifique du Service User

Pour tester ce service, une préparation spécifique est nécessaire, notamment pour l'administrateur (gestion du hachage de mot de passe) et la création de l'utilisateur.

#### 1. Pré-requis : Initialisation Admin
Le mot de passe de l'administrateur présent en base de données doit être correctement haché (bcrypt). Un script utilitaire est fourni pour définir le mot de passe de l'admin à `admin123`.

Exécutez les commandes suivantes dans votre terminal :

```bash
# 1. Démarrer les conteneurs
docker-compose up --build -d

# 2. Entrer dans le conteneur du service User
docker compose exec user sh

# 3. Lancer le script de mise à jour du mot de passe
npx ts-node src/help/setAdminPassword.ts
```

✅ **Note** : Une fois le script terminé ("Password updated : admin123"), le mot de passe de l'admin sera `admin123`.

#### 2. Workflow de connexion

Une fois le pré-requis validé, suivez cet ordre pour obtenir vos tokens :

**Utilisateur Standard :**
- Créer un utilisateur via la route `register`.
- Se connecter via la route `login` pour récupérer le `user_token` et le `user_id`.

**Administrateur :**
- Se connecter via la route `login` (avec le mot de passe `admin123`) pour récupérer l'`admin_token` et l'`admin_id`.

#### 3. Variables d'environnement Insomnia
Dans Insomnia, configurez les variables d'environnement suivantes avec les données récoltées :

| Variable      | Valeur / Description                                    |
| :------------ |:--------------------------------------------------------|
| `base_url`    | `http://localhost:8080`                                 |
| `admin_token` | Token obtenu à l'étape connexion admin                  |
| `user_token`  | Token obtenu à l'étape connexion user                   |
| `admin_id`    | ID de l'administrateur obtenu à l'étape connexion admin |
| `user_id`     | ID de l'utilisateur obtenu à l'étape connexion user     |

Vous pouvez maintenant tester les endpoints protégés en utilisant les tokens appropriés.

## Dockerisation

L'application est entièrement containerisée:
- 5 conteneurs pour les micro-services
- 3 conteneurs MongoDB
- Docker Compose pour l'orchestration

## Équipe

Projet réalisé par un groupe de 4 élèves dans le cadre du cours d'Architectures Distribuées.

## Notes Techniques

- **Communication inter-services**: REST, GraphQL, gRPC
- **Scalabilité**: Architecture pensée pour être facilement extensible
- **Performance**: Utilisation de gRPC pour les opérations critiques (Game Engine)
- **Sécurité**: Authentification JWT, hachage bcrypt des mots de passe

## Licence

Ce projet est réalisé dans un cadre académique.

BOURREAU Quentin / SORIN Kevin / CARFANTAN Thomas / KOWALSKI Damien

---

**Note**: Ce projet nécessite un token Spotify API valide pour fonctionner. Consultez la [documentation Spotify](https://developer.spotify.com/documentation/web-api) pour obtenir vos credentials.