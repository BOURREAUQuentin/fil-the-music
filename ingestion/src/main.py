import asyncio
import os
import base64
import logging
import time
from typing import Optional, Dict, List, Any

import httpx
from dotenv import load_dotenv

# Configuration du logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("IngestionService")

# Chargement des variables d'environnement
load_dotenv()

# Configuration
SPOTIFY_CLIENT_ID = os.getenv("SPOTIFY_CLIENT_ID")
SPOTIFY_CLIENT_SECRET = os.getenv("SPOTIFY_CLIENT_SECRET")
# Note: Le port par défaut du service catalog dans docker-compose.yml semble être 3200, mais le prompt indiquait 8000.
# Je mets une valeur par défaut, surchargeable via ENV.
CATALOG_URL = os.getenv("CATALOG_URL", "http://catalog:3200/graphql")

# IDs des Playlists
PLAYLIST_FRANCE = "2IgPkhcHbgQ4s4PdCxljAx" # Pareil qu'en dessous
PLAYLIST_GLOBAL = "5ABHKGoOzxkaa28ttQV9sE" # Top 100 monde par un utilisateur et non par spotify sinon ça casse car c'est des rats (merci Spotify)

class SpotifyClient:
    def __init__(self, client_id: str, client_secret: str):
        self.client_id = client_id
        self.client_secret = client_secret
        self.token: Optional[str] = None

    async def authenticate(self):
        """
        Authentification Client Credentials flow.
        Cette méthode est appelée au début de chaque cycle pour obtenir un token frais.
        """
        auth_url = "https://accounts.spotify.com/api/token"
        auth_header = base64.b64encode(f"{self.client_id}:{self.client_secret}".encode()).decode()
        
        headers = {
            "Authorization": f"Basic {auth_header}",
            "Content-Type": "application/x-www-form-urlencoded"
        }
        data = {"grant_type": "client_credentials"}

        async with httpx.AsyncClient() as client:
            try:
                response = await client.post(auth_url, headers=headers, data=data)
                response.raise_for_status()
                json_data = response.json()
                self.token = json_data["access_token"]
                logger.info("Authentification Spotify réussie. Nouveau token généré.")
            except httpx.HTTPError as e:
                logger.error(f"Erreur d'authentification Spotify: {e}")
                raise

    async def get_playlist_tracks(self, playlist_id: str) -> List[Dict[str, Any]]:
        """Récupère les 50 premiers titres d'une playlist."""
        if not self.token:
            raise Exception("Token manquant. Authentifiez-vous d'abord.")

        url = f"https://api.spotify.com/v1/playlists/{playlist_id}/tracks"
        headers = {"Authorization": f"Bearer {self.token}"}
        params = {
            "limit": 50,
            "fields": "items(track(id,name,album(name,release_date),artists(id,name),popularity))"
        }

        async with httpx.AsyncClient() as client:
            try:
                response = await client.get(url, headers=headers, params=params)
                response.raise_for_status()
                data = response.json()
                return [item['track'] for item in data.get('items', []) if item.get('track')]
            except httpx.HTTPError as e:
                logger.error(f"Erreur lors de la récupération de la playlist {playlist_id}: {e}")
                return []

class CatalogClient:
    def __init__(self, url: str):
        self.url = url

    async def send_batch(self, tracks: List[Dict[str, Any]]):
        """
        Envoie les artistes puis les titres en lot au service Catalog.
        Respecte la séparation Artist / Track du schéma GraphQL.
        """
        if not tracks:
            return

        # 1. Préparer les artistes uniques
        # On utilise l'ID de l'artiste principal comme ID unique
        artists_map = {}
        for t in tracks:
            if t['artists']:
                main_artist = t['artists'][0]
                a_id = main_artist['id']
                if a_id not in artists_map:
                    artists_map[a_id] = {
                        "artist_id": a_id,
                        "name": main_artist['name'],
                        "genres": [] # L'API track ne donne pas les genres, on envoie vide pour l'instant
                    }

        artists_list = list(artists_map.values())
        
        # 2. Envoyer les artistes (add_many_artists)
        # Mutation: add_many_artists(artists_list: [ArtistInput!]!): String
        mutation_artists = """
        mutation AddManyArtists($list: [ArtistInput!]!) {
            add_many_artists(artists_list: $list)
        }
        """
        
        if artists_list:
            logger.info(f"Envoi de {len(artists_list)} artistes...")
            await self._send_graphql(mutation_artists, {"list": artists_list})

        # 3. Préparer les tracks
        # Input TrackInput: { track_id, title, artist_id, album_name, release_date }
        tracks_input = []
        for t in tracks:
            artist_id = t['artists'][0]['id'] if t['artists'] else "unknown"
            tracks_input.append({
                "track_id": t['id'],
                "title": t['name'],
                "artist_id": artist_id,
                "album_name": t['album']['name'],
                "release_date": t['album']['release_date']
            })

        # 4. Envoyer les tracks (add_many_tracks)
        # Mutation: add_many_tracks(tracks_list: [TrackInput!]!): String
        mutation_tracks = """
        mutation AddManyTracks($list: [TrackInput!]!) {
            add_many_tracks(tracks_list: $list)
        }
        """
        
        if tracks_input:
            logger.info(f"Envoi de {len(tracks_input)} titres...")
            await self._send_graphql(mutation_tracks, {"list": tracks_input})

    async def _send_graphql(self, query: str, variables: Dict[str, Any]):
        async with httpx.AsyncClient() as client:
            try:
                payload = {"query": query, "variables": variables}
                response = await client.post(self.url, json=payload, timeout=30.0)
                
                if response.status_code != 200:
                    logger.error(f"Erreur HTTP Catalog ({response.status_code}): {response.text}")
                    return

                result = response.json()
                if "errors" in result:
                    logger.error(f"Erreur GraphQL: {result['errors']}")
                else:
                    logger.info("Lot envoyé avec succès.")

            except httpx.HTTPError as e:
                logger.error(f"Erreur de connexion au Catalog: {e}")
            except Exception as e:
                logger.error(f"Erreur inattendue: {e}")

async def run_cycle():
    """Exécute un cycle complet d'ingestion."""
    logger.info("Début du cycle d'ingestion.")
    
    spotify = SpotifyClient(SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET)
    catalog = CatalogClient(CATALOG_URL)

    try:
        # 1. Authentification
        await spotify.authenticate()

        # 2. Récupération des playlists
        tasks = [
            spotify.get_playlist_tracks(PLAYLIST_FRANCE),
            spotify.get_playlist_tracks(PLAYLIST_GLOBAL)
        ]
        results = await asyncio.gather(*tasks)
        
        tracks_france = results[0]
        tracks_global = results[1]
        
        logger.info(f"Récupéré {len(tracks_france)} titres France et {len(tracks_global)} titres Global.")

        # 3. Fusion (dédoublonnage)
        all_tracks_map = {t['id']: t for t in (tracks_france + tracks_global)}
        all_tracks = list(all_tracks_map.values())

        # 4. Envoi groupé
        await catalog.send_batch(all_tracks)

    except Exception as e:
        logger.critical(f"Erreur critique durant le cycle: {e}")

    logger.info("Cycle terminé.")

async def main():
    """Boucle principale."""
    if not SPOTIFY_CLIENT_ID or not SPOTIFY_CLIENT_SECRET:
        logger.error("Variables d'environnement SPOTIFY_CLIENT_ID ou SPOTIFY_CLIENT_SECRET manquantes.")
        return

    logger.info("Service Ingestion démarré. Fréquence: 1h.")

    while True:
        start_time = time.time()
        
        await run_cycle()
        
        # Calcul du temps de sommeil restant
        elapsed = time.time() - start_time
        sleep_time = max(0, 3600 - elapsed)
        
        logger.info(f"Mise en veille pour {int(sleep_time)} secondes...")
        await asyncio.sleep(sleep_time)

if __name__ == "__main__":
    asyncio.run(main())
