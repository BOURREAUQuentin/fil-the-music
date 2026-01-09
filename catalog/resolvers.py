import json
from graphql import GraphQLError
import requests, time
from pymongo import MongoClient

client = MongoClient("mongodb://damien:filthemusic@mongo:27017/")
database = client["catalog_db"]
tracks = database["tracks"]
artists = database["artists"]

### READ OPERATIONS

def track_json(_, info):
    """
    Retrieves the complete list of tracks from the database.
    
    Returns:
        list: A list of dictionaries representing each track.
    Raises:
        GraphQLError: If there is a database connection issue or query failure.
    """
    try:
        tracks_to_return = list(tracks.find({}))
    except Exception as e:
        print(f"MONGO ERROR: {e}", flush=True) 
        raise GraphQLError("Database connection error or query failed")
    return tracks_to_return

def artist_json(_, info):
    """
    Retrieves the complete list of artists from the database.
    
    Returns:
        list: A list of dictionaries representing each artist.
    Raises:
        GraphQLError: If there is a database connection issue or query failure.
    """
    try:
        artists_to_return = list(artists.find({}))
    except Exception as e:
        print(f"MONGO ERROR: {e}", flush=True) 
        raise GraphQLError("Database connection error or query failed")
    return artists_to_return

def track_by_id(_, info, track_id):
    """
    Finds a specific track by its track_id.
    
    Args:
        track_id (str): The unique identifier of the track.
    Returns:
        dict: The data of the found track.
    Raises:
        GraphQLError: If the track is not found or a technical error occurs.
    """
    try:
        track = tracks.find_one({"track_id": track_id})
    except Exception as e:
        print(f"MONGO ERROR: {e}", flush=True) 
        raise GraphQLError("Database connection error or query failed")
    
    if track is None:
        raise GraphQLError(f"Track not found with track_id : {track_id}")
    return track

def artist_by_id(_, info, artist_id):
    """
    Finds a specific artist by its artist_id.
    
    Args:
        artist_id (str): The unique identifier of the artist.
    Returns:
        dict: The data of the found artist.
    Raises:
        GraphQLError: If the artist is not found or a technical error occurs.
    """
    try:
        artist = artists.find_one({"artist_id": artist_id})
    except Exception as e:
        print(f"MONGO ERROR: {e}", flush=True) 
        raise GraphQLError("Database connection error or query failed")
    
    if artist is None:
        raise GraphQLError(f"Artist not found with artist_id : {artist_id}")
    return artist

### CREATE OPERATIONS

def add_track(_, info, track_id, title, artist_id, album_name, release_date):
    """
    Adds a new track after checking for track_id uniqueness and artist existence.
    
    Args:
        track_id (str): Unique track identifier.
        title (str): Track title.
        artist_id (str): Associated artist identifier.
        album_name (str): Name of the album.
        release_date (str): Release date of the track.
    Returns:
        dict: The newly created track object.
    """
    try:
        if tracks.find_one({"track_id": track_id}):
            raise GraphQLError(f"track_id already exists : {track_id}")
            
        if not artists.find_one({"artist_id": artist_id}):
            raise GraphQLError(f"artist_id does not exist : {artist_id}")
            
        new_track = {
            "track_id": track_id,
            "title": title,
            "artist_id": artist_id,
            "album_name": album_name,
            "release_date": release_date
        }
        tracks.insert_one(new_track)
        return new_track
    
    except GraphQLError as ge:
        raise ge
    
    except Exception as e:
        print(f"MONGO ERROR: {e}", flush=True) 
        raise GraphQLError("Database connection error or query failed")

def add_many_tracks(_, info, tracks_list):
    """
    Inserts multiple tracks efficiently after validating all IDs in bulk.
    """
    try:
        input_track_ids = [t["track_id"] for t in tracks_list]
        input_artist_ids = list(set(t["artist_id"] for t in tracks_list))

        # Bulk check existing tracks
        existing = tracks.find({"track_id": {"$in": input_track_ids}}, {"track_id": 1})
        existing_ids = [t["track_id"] for t in existing]
        if existing_ids:
            raise GraphQLError(f"track_ids already exist: {existing_ids}")

        # Bulk check artists existence
        found_artists = artists.find({"artist_id": {"$in": input_artist_ids}}, {"artist_id": 1})
        found_artist_ids = [a["artist_id"] for a in found_artists]
        
        for a_id in input_artist_ids:
            if a_id not in found_artist_ids:
                raise GraphQLError(f"artist_id {a_id} does not exist")

        # Insert everything
        tracks.insert_many(tracks_list)
        return "Tracks added"

    except GraphQLError as ge:
        raise ge
    except Exception as e:
        print(f"MONGO ERROR: {e}", flush=True) 
        raise GraphQLError("Database connection error or query failed")

def add_artist(_, info, artist_id, name, genres):
    """
    Adds a new artist to the database.
    
    Args:
        artist_id (str): Unique artist identifier.
        name (str): Artist's name.
        genres (list): List of musical genres.
    Returns:
        dict: The newly created artist object.
    """
    try:
        if artists.find_one({"artist_id": artist_id}):
            raise GraphQLError(f"artist_id already exists : {artist_id}")
            
        new_artist = {
            "artist_id": artist_id,
            "name": name,
            "genres": genres,
        }
        artists.insert_one(new_artist)
        return new_artist
        
    except GraphQLError as ge:
        raise ge
    
    except Exception as e:
        print(f"MONGO ERROR: {e}", flush=True) 
        raise GraphQLError("Database connection error or query failed")

def add_many_artists(_, info, artists_list):
    """Validates artist_id uniqueness and inserts multiple artists in bulk."""
    try:
        input_artist_ids = [a["artist_id"] for a in artists_list]

        # Bulk check for existing artists
        existing = artists.find({"artist_id": {"$in": input_artist_ids}}, {"artist_id": 1})
        existing_ids = [a["artist_id"] for a in existing]
        
        if existing_ids:
            raise GraphQLError(f"artist_ids already exist: {existing_ids}")

        # Perform the bulk insertion
        artists.insert_many(artists_list)
        
        return "Artists added"

    except GraphQLError as ge:
        raise ge
    except Exception as e:
        print(f"MONGO ERROR: {e}", flush=True) 
        raise GraphQLError("Database connection error or query failed")
    
### UPDATE OPERATIONS

def update_track(_, info, track_id, title=None, artist_id=None, album_name=None, release_date=None):
    """
    Updates an existing track's information based on the provided track_id.
    
    Args:
        track_id (str): The identifier of the track to update.
        title, artist_id, album_name, release_date (str, optional): Fields to update.
    Returns:
        dict: The updated track object.
    Raises:
        GraphQLError: If the track is not found or a database error occurs.
    """
    try:
        update_data = {}
        if title: update_data["title"] = title
        if artist_id: 
            # Check if new artist exists before updating
            if not artists.find_one({"artist_id": artist_id}):
                raise GraphQLError(f"Functional Error: artist_id {artist_id} does not exist")
            update_data["artist_id"] = artist_id
        if album_name: update_data["album_name"] = album_name
        if release_date: update_data["release_date"] = release_date

        if not update_data:
            raise GraphQLError("No fields provided for update")

        query_filter = {'track_id': track_id}
        update_operation = {'$set': update_data}
        
        result = tracks.update_one(query_filter, update_operation)

        if result.matched_count == 0:
            raise GraphQLError(f"Track not found with track_id: {track_id}")

        return tracks.find_one({"track_id": track_id})

    except GraphQLError as ge:
        raise ge
    except Exception as e:
        print(f"TECHNICAL ERROR: {e}", flush=True)
        raise GraphQLError("Database connection error or query failed")
    
def update_artist(_, info, artist_id, name=None, genres=None):
    """
    Updates an existing artist's information based on the provided artist_id.
    
    Args:
        artist_id (str): The identifier of the artist to update.
        name (str, optional): The new name of the artist.
        genres (list, optional): The new list of genres.
    Returns:
        dict: The updated artist object.
    Raises:
        GraphQLError: If the artist is not found or a database error occurs.
    """
    try:
        update_data = {}
        if name: 
            update_data["name"] = name
        if genres is not None: 
            update_data["genres"] = genres

        if not update_data:
            raise GraphQLError("No fields provided for update")

        query_filter = {'artist_id': artist_id}
        update_operation = {'$set': update_data}
        
        result = artists.update_one(query_filter, update_operation)

        if result.matched_count == 0:
            raise GraphQLError(f"Functional Error: Artist not found with artist_id: {artist_id}")

        return artists.find_one({"artist_id": artist_id})

    except GraphQLError as ge:
        raise ge
    except Exception as e:
        print(f"TECHNICAL ERROR: {e}", flush=True)
        raise GraphQLError("Database connection error or query failed")
    
def remove_track(_, info, track_id):
    """
    Deletes a track by ID and returns the updated list of all tracks.
    
    Args:
        track_id (str): The unique identifier of the track to remove.
    Returns:
        list: The remaining tracks in the collection.
    """
    try:
        track = tracks.find_one({"track_id": track_id})
    
        if track is None:
            raise GraphQLError(f"Track not found with track_id : {track_id}")
        
        query_filter = { "track_id": track_id }
        result = tracks.delete_one(query_filter)
        return (f"The track with track_id {track_id}, has been removed")
    
    except GraphQLError as ge:
        raise ge
    except Exception as e:
        print(f"MONGO ERROR: {e}", flush=True) 
        raise GraphQLError("Database connection error or query failed")
    
def remove_artist(_, info, artist_id):
    """
    Deletes an artist by ID and returns the deletion result.
    
    Args:
        artist_id (str): The unique identifier of the artist to remove.
    Returns:
        DeleteResult: The result object from MongoDB delete operation.
    """
    try:
        artist = artists.find_one({"artist_id": artist_id})
    
        if artist is None:
            raise GraphQLError(f"Artist not found with artist_id : {artist_id}")
        
        query_filter = { "artist_id": artist_id }
        result = artists.delete_one(query_filter)
        return (f"The artist with artist_id {artist_id}, has been removed")
    
    except GraphQLError as ge:
        raise ge
    except Exception as e:
        print(f"MONGO ERROR: {e}", flush=True) 
        raise GraphQLError("Database connection error or query failed")