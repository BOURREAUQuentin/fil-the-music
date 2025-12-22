import json
from graphql import GraphQLError
import requests, time

with open('{}/bdd/tracks.json'.format("."), "r") as jsf:
    tracks = json.load(jsf)["tracks"]
    
def write(tracks_data):
    with open('{}/databases/movies.json'.format("."), 'w') as f:
        full = {}
        full['tracks'] = tracks_data
        json.dump(full, f)
        
with open('{}/bdd/artists.json'.format("."), "r") as jsf:
    artists = json.load(jsf)["artists"]
    
def write(artists_data):
    with open('{}/databases/movies.json'.format("."), 'w') as f:
        full = {}
        full['artists'] = artists_data
        json.dump(full, f)

def track_json(_,info):
    return tracks