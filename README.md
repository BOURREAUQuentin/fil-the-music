# Projet Quiz Musical - Architecture Micro-services

Application de quiz musical permettant de tester ses connaissances sur la musique à travers une architecture micro-services moderne.

## Présentation du Projet

### Principe

Application de quiz musical où les utilisateurs peuvent tester leurs connaissances musicales à travers différents types de quiz.

### Fonctionnement

Un quiz est composé d'une suite de questions avec une seule bonne réponse par question. L'application propose plusieurs quiz prédéfinis et permet la création de nouveaux quiz.

### Types de Quiz

#### Quiz Simple (Admin uniquement)
- Questions et réponses entièrement personnalisables
- Contenu défini manuellement par l'administrateur
- Pas de vérification automatique des données
- Liberté totale sur le contenu

#### Quiz Personnalisé (Utilisateurs)
- Filtrage par genre musical
- Questions générées aléatoirement depuis le Catalogue
- Basé sur les données Spotify (artistes, morceaux, dates)
- Création automatisée accessible à tous

#### Quiz "For You" (Utilisateurs)
- Personnalisé selon les préférences de l'utilisateur
- Basé sur les artistes favoris de l'utilisateur
- Questions générées depuis le Catalogue
- Création automatique adaptée au profil

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
- Token Spotify API

### Installation

```bash
# Cloner le repository
git clone https://github.com/BOURREAUQuentin/fil-the-music.git
cd fil-the-music

# Configurer les variables d'environnement dans ingestion et user
cp .env.example .env
# Éditer .env et ajouter les informations manquantes

# Lancer l'environnement Docker
docker-compose up --build
```

### Accès

L'API est accessible sur `http://localhost:8000`

## Tests

Des exports Insomnia sont mis à disposition pour faciliter le test des endpoints :
- **Global :** Un export complet est disponible à la racine du projet.
- **Spécifique :** Chaque dossier de service contient son propre export dédié.

### Configuration du Service User

Pour tester ce service, une configuration spécifique de l'environnement est requise afin de gérer les droits d'accès.

**1. Obtention des tokens**

Veuillez suivre cet ordre d'exécution pour récupérer les tokens nécessaires :
1. Se connecter en tant qu'**Utilisateur** (récupérer le `user_token`).
2. Se connecter en tant qu'**Administrateur** (récupérer l'`admin_token`).

**2. Variables d'environnement**

Dans Insomnia, configurez les variables d'environnement suivantes :

| Variable      | Valeur / Description                                   |
| :------------ | :----------------------------------------------------- |
| `base_url`    | `http://localhost:3201`                                |
| `admin_token` | Token obtenu à l'étape connexion admin                 |
| `user_token`  | Token obtenu à l'étape connexion user                  |
| `admin_id`    | ID de l'administrateur                                 |
| `user_id`     | ID de l'utilisateur                                    |

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

## Licence

Ce projet est réalisé dans un cadre académique.

---

**Note**: Ce projet nécessite un token Spotify API valide pour fonctionner. Consultez la [documentation Spotify](https://developer.spotify.com/documentation/web-api) pour obtenir vos credentials.