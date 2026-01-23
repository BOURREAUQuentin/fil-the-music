import asyncio
import os
import base64
import logging
import time
import re
from typing import Optional, Dict, List, Any
from collections import Counter

import httpx
from dotenv import load_dotenv
from fastapi import FastAPI, BackgroundTasks, HTTPException
from pydantic import BaseModel

# Configuration du logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("IngestionService")

# Chargement des variables d'environnement
load_dotenv()

SPOTIFY_CLIENT_ID = os.getenv("SPOTIFY_CLIENT_ID")
SPOTIFY_CLIENT_SECRET = os.getenv("SPOTIFY_CLIENT_SECRET")
CATALOG_URL = os.getenv("CATALOG_URL", "http://catalog:3200/graphql")

# IDs des Playlists (Loop daemon)
PLAYLIST_FRANCE = "2IgPkhcHbgQ4s4PdCxljAx"
PLAYLIST_GLOBAL = "5ABHKGoOzxkaa28ttQV9sE"

app = FastAPI()

# --- Classes Clientes (Adaptées) ---

class SpotifyClient:
    def __init__(self, client_id: str, client_secret: str):
        self.client_id = client_id
        self.client_secret = client_secret
        self.token: Optional[str] = None
        self.token_expiry: float = 0

    async def get_valid_token(self) -> str:
        """Retourne un token valide, le rafraîchit si nécessaire."""
        if not self.token or time.time() > self.token_expiry:
            await self.authenticate()
        return self.token

    async def authenticate(self):
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
                self.token_expiry = time.time() + json_data.get("expires_in", 3600) - 60
                logger.info("Nouveau token Spotify généré.")
            except httpx.HTTPError as e:
                logger.error(f"Erreur Auth Spotify: {e}")
                raise

    async def get_playlist_tracks(self, playlist_id: str, max_tracks: int = 300) -> List[Dict[str, Any]]:
        """
        Récupère jusqu'à 'max_tracks' morceaux d'une playlist en paginant.
        """
        token = await self.get_valid_token()
        url = f"https://api.spotify.com/v1/playlists/{playlist_id}/tracks"
        headers = {"Authorization": f"Bearer {token}"}

        all_tracks = []
        offset = 0
        limit_per_call = 100

        async with httpx.AsyncClient() as client:
            while len(all_tracks) < max_tracks:
                params = {
                    "limit": limit_per_call,
                    "offset": offset,
                    "fields": "items(track(id,name,album(name,release_date),artists(id,name),popularity)),total"
                }

                try:
                    response = await client.get(url, headers=headers, params=params)
                    if response.status_code == 404:
                        return [] # Playlist introuvable

                    response.raise_for_status()
                    data = response.json()

                    items = data.get('items', [])
                    if not items:
                        break # Plus de morceaux disponibles

                    # On nettoie et on ajoute
                    valid_tracks = [item['track'] for item in items if item.get('track')]
                    all_tracks.extend(valid_tracks)

                    # Si on a reçu moins que demandé (ex: 42 reçus pour limite 100), c'est la fin
                    if len(items) < limit_per_call:
                        break

                    offset += limit_per_call

                except httpx.HTTPError as e:
                    print(f"Erreur lors de la pagination : {e}")
                    break

        return all_tracks[:max_tracks]

    async def get_user_public_playlists_artists(self, user_id: str) -> List[str]:
        """
        Récupère les playlists publiques d'un utilisateur, compte les artistes,
        et retourne le TOP 5.
        """
        token = await self.get_valid_token()
        url = f"https://api.spotify.com/v1/users/{user_id}/playlists"
        headers = {"Authorization": f"Bearer {token}"}
        params = {"limit": 10}

        async with httpx.AsyncClient() as client:
            try:
                response = await client.get(url, headers=headers, params=params)
                if response.status_code == 404:
                    logger.warning(f"Utilisateur Spotify {user_id} introuvable.")
                    # On retourne une liste vide plutôt que de planter, 
                    # le client affichera que rien n'a été trouvé.
                    return []
                response.raise_for_status()
                playlists = response.json().get('items', [])
            except Exception as e:
                logger.error(f"Erreur fetch user playlists: {e}")
                return []
        
        artist_counter = Counter()
        tasks = []
        for pl in playlists:
            if pl and pl.get('id'):
                tasks.append(self.get_playlist_tracks(pl['id']))
        
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        for tracks in results:
            if isinstance(tracks, list):
                for t in tracks:
                    if t.get('artists'):
                        for artist in t['artists']:
                            artist_counter[artist['name']] += 1
        
        top_5 = [name for name, count in artist_counter.most_common(5)]
        logger.info(f"Top 5 artistes pour {user_id}: {top_5}")
        return top_5

    async def search_artist_top_tracks(self, artist_name: str) -> List[Dict[str, Any]]:
        token = await self.get_valid_token()
        async with httpx.AsyncClient() as client:
            search_url = "https://api.spotify.com/v1/search"
            headers = {"Authorization": f"Bearer {token}"}
            params = {"q": artist_name, "type": "artist", "limit": 1}
            try:
                resp = await client.get(search_url, headers=headers, params=params)
                data = resp.json()
                items = data.get('artists', {}).get('items', [])
                if not items: return []
                artist_id = items[0]['id']
            except Exception: return []

            top_url = f"https://api.spotify.com/v1/artists/{artist_id}/top-tracks"
            params_top = {"market": "FR"} 
            try:
                resp = await client.get(top_url, headers=headers, params=params_top)
                data = resp.json()
                return data.get('tracks', [])
            except Exception: return []

    async def get_artists(self, artist_ids: List[str]) -> List[Dict[str, Any]]:
        """Récupère les détails complets (avec genres) pour une liste d'IDs d'artistes."""
        if not artist_ids:
            return []
            
        token = await self.get_valid_token()
        url = "https://api.spotify.com/v1/artists"
        headers = {"Authorization": f"Bearer {token}"}
        
        all_artists = []
        # Spotify API limite à 50 IDs par appel
        chunk_size = 20
        for i in range(0, len(artist_ids), chunk_size):
            chunk = artist_ids[i:i + chunk_size]
            params = {"ids": ",".join(chunk)}
            
            async with httpx.AsyncClient() as client:
                try:
                    response = await client.get(url, headers=headers, params=params)
                    response.raise_for_status()
                    data = response.json()
                    all_artists.extend(data.get('artists', []))
                except Exception as e:
                    logger.error(f"Erreur fetch artists details: {e}")
                    
        return all_artists

class CatalogClient:
    def __init__(self, url: str):
        self.url = url

    async def send_batch(self, tracks: List[Dict[str, Any]]):
        if not tracks: return

        # 1. Collecter les IDs des artistes
        artist_ids = set()
        for t in tracks:
            if t.get('artists'):
                main_artist = t['artists'][0]
                if main_artist.get('id'):
                    artist_ids.add(main_artist['id'])

        # 2. Récupérer les détails complets des artistes (genres inclus)
        artists_details_map = {}
        if artist_ids:
            # Note: Utilisation de l'instance globale spotify_client
            # Dans un code plus structuré, on passerait le client en dépendance.
            try:
                details = await spotify_client.get_artists(list(artist_ids))
                for d in details:
                    if d and d.get('id'):
                        artists_details_map[d['id']] = d
            except Exception as e:
                logger.error(f"Impossible de récupérer les détails artistes: {e}")

        # 3. Construire la map des artistes à envoyer
        artists_map = {}
        for t in tracks:
            if t.get('artists'):
                main_artist = t['artists'][0]
                a_id = main_artist['id']
                if a_id not in artists_map:
                    # Récupération des genres depuis les détails
                    raw_genres = []
                    if a_id in artists_details_map:
                        raw_genres = artists_details_map[a_id].get('genres', [])

                    final_genres = raw_genres
                    if not final_genres:
                        final_genres = ["pop"]

                    artists_map[a_id] = {
                        "artist_id": a_id,
                        "name": main_artist['name'],
                        "genres": final_genres
                    }
        artists_list = list(artists_map.values())
        
        mutation_artists = """
        mutation AddManyArtists($list: [ArtistInput!]!) {
            add_many_artists(artists_list: $list)
        }
        """
        if artists_list:
            await self._send_graphql(mutation_artists, {"list": artists_list})

        # 4. Préparer et envoyer les tracks
        tracks_input = []
        for t in tracks:
            if t.get('artists') and t.get('album'):
                artist_id = t['artists'][0]['id']
                tracks_input.append({
                    "track_id": t['id'],
                    "title": t['name'],
                    "artist_id": artist_id,
                    "album_name": t['album']['name'],
                    "release_date": t['album']['release_date']
                })

        mutation_tracks = """
        mutation AddManyTracks($list: [TrackInput!]!) {
            add_many_tracks(tracks_list: $list)
        }
        """
        if tracks_input:
            await self._send_graphql(mutation_tracks, {"list": tracks_input})

    async def _send_graphql(self, query: str, variables: Dict[str, Any]):
        async with httpx.AsyncClient() as client:
            try:
                payload = {"query": query, "variables": variables}
                response = await client.post(self.url, json=payload, timeout=30.0)
            except Exception as e:
                logger.error(f"Catalog Connection Error: {e}")

# --- Instances Globales ---
spotify_client = SpotifyClient(SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET)
catalog_client = CatalogClient(CATALOG_URL)

# --- Background Task ---
async def ingest_artist_data(artist_name: str):
    logger.info(f"[Background] Ingestion pour l'artiste: {artist_name}")
    try:
        tracks = await spotify_client.search_artist_top_tracks(artist_name)
        if tracks:
            await catalog_client.send_batch(tracks)
    except Exception as e:
        logger.error(f"[Background] Erreur ingestion {artist_name}: {e}")

async def daemon_loop():
    logger.info("Démarrage du Daemon Loop.")
    while True:
        try:
            await spotify_client.get_valid_token()
            tasks = [
                spotify_client.get_playlist_tracks(PLAYLIST_FRANCE, 50),
                spotify_client.get_playlist_tracks(PLAYLIST_GLOBAL, 50)
            ]
            results = await asyncio.gather(*tasks, return_exceptions=True)
            all_tracks = []
            for res in results:
                if isinstance(res, list): all_tracks.extend(res)
            
            unique_tracks = {t['id']: t for t in all_tracks}.values()
            if unique_tracks:
                await catalog_client.send_batch(list(unique_tracks))
        except Exception as e:
            logger.error(f"[Daemon] Erreur: {e}")
        await asyncio.sleep(3600)

@app.on_event("startup")
async def startup_event():
    asyncio.create_task(daemon_loop())

# --- API Endpoints ---

class IngestUserRequest(BaseModel):
    identifier: str

@app.post("/ingest/user")
async def ingest_user_profile(req: IngestUserRequest, background_tasks: BackgroundTasks):
    """
    Accepte un ID Spotify OU une URL de profil (https://open.spotify.com/user/...)
    Extrait l'ID et lance l'ingestion.
    """
    raw_input = req.identifier
    logger.info(f"Reçu demande ingestion: {raw_input}")

    # Extraction ID via Regex
    # Supporte: 
    # - https://open.spotify.com/user/1123456789
    # - spotify:user:1123456789
    # - 1123456789 (brut)
    spotify_id = raw_input
    
    match_url = re.search(r'user/([a-zA-Z0-9]+)', raw_input)
    if match_url:
        spotify_id = match_url.group(1)
    
    match_uri = re.search(r'spotify:user:([a-zA-Z0-9]+)', raw_input)
    if match_uri:
        spotify_id = match_uri.group(1)

    logger.info(f"ID extrait: {spotify_id}")

    # Récupération
    top_artists = await spotify_client.get_user_public_playlists_artists(spotify_id)
    
    if not top_artists:
        return {"status": "not_found_or_empty", "resolved_id": spotify_id, "artists": []}

    # Background ingestion
    for artist in top_artists:
        background_tasks.add_task(ingest_artist_data, artist)
    
    return {"status": "processing", "resolved_id": spotify_id, "artists": top_artists}