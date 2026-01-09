# Architecture du Microservice `game`

Ce service est conçu selon les principes de l'**Architecture Hexagonale** (aussi appelée _Ports and Adapters_). L'objectif est de séparer la logique métier (le cœur de l'application) des détails techniques (base de données, API, etc.), rendant le code plus modulaire, plus facile à tester et à maintenir.

## Structure des Dossiers

L'organisation du code reflète cette séparation des responsabilités :

-   **`/api`**: Contient les définitions du contrat de l'API. Pour ce projet, il s'agit des fichiers `.proto` qui décrivent notre service gRPC.

-   **`/cmd/server`**: Le point d'entrée de notre serveur. C'est un **adaptateur pilotant** (_Driving Adapter_). Son rôle est de recevoir les requêtes externes (ici, des appels gRPC) et de les transmettre au cœur de l'application.

-   **`/internal/core/domain`**: Le **cœur du métier**. Il contient les entités et la logique métier pures (ex: `Quiz`). Cette couche n'a aucune dépendance avec le reste de l'application.

-   **`/internal/core/ports`**: Les **ports** de l'application. Ce sont des interfaces Go qui définissent les contrats dont le cœur a besoin pour communiquer avec l'extérieur (ex: `QuizRepository`). Le cœur dépend de ces abstractions, pas des implémentations.

-   **`/internal/adapters/repository`**: Un **adaptateur piloté** (_Driven Adapter_). Il fournit une implémentation concrète pour un port défini dans `core/ports`. Par exemple, `MongoRepository` implémente les interfaces de repository en utilisant une base de données MongoDB.

-   **`main.go`**: Le point de démarrage de l'application. Son rôle est d'initialiser et de connecter les différentes couches entre elles (c'est ici que se fait l'**injection de dépendances**).

## Comment lancer le serveur grpc ?

1. Générer les fichiers grâce au `game.proto`: `make gen`
2. Build et lancer le serveur: `go run ./main.go`