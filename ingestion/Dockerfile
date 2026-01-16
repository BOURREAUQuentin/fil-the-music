# Utilisation d'une image légère Python 3.11
FROM python:3.11-slim

# Configuration des variables d'environnement pour optimiser Python dans Docker
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# Installation des dépendances système si nécessaire (ici curl pour healthcheck optionnel)
RUN apt-get update && apt-get install -y --no-install-recommends curl && rm -rf /var/lib/apt/lists/*

# Copie des requirements et installation
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copie du code source
COPY src/ ./src/

# Création d'un utilisateur non-root pour la sécurité
RUN useradd -m appuser
USER appuser

# Commande de lancement
CMD ["python", "src/main.py"]
