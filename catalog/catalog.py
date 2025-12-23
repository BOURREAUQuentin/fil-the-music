from ariadne import graphql_sync, make_executable_schema, load_schema_from_path, ObjectType, QueryType, MutationType
from flask import Flask, request, jsonify, make_response
import time, json, requests
from werkzeug.exceptions import NotFound
from flask_cors import CORS
import resolvers as r

app = Flask(__name__)

CORS(app)

PORT = 3200
HOST = '0.0.0.0'

# création du schéma GraphQL
type_defs = load_schema_from_path('catalog.graphql')

query = QueryType()
mutation = MutationType()

track = ObjectType('Track')
artist = ObjectType('Artist')

query.set_field('track_json', r.track_json)
query.set_field('artist_json', r.artist_json)
query.set_field('track_by_id', r.track_by_id)
query.set_field('artist_by_id', r.artist_by_id)

mutation.set_field('add_track', r.add_track)
mutation.set_field('add_many_tracks', r.add_many_tracks)
mutation.set_field('add_artist', r.add_artist)
mutation.set_field('add_many_artists', r.add_many_artists)
mutation.set_field('update_track', r.update_track)
mutation.set_field('update_artist', r.update_artist)
mutation.set_field('remove_track', r.remove_track)
mutation.set_field('remove_artist', r.remove_artist)

schema = make_executable_schema(type_defs, track, artist, query, mutation)

# page d’accueil du service
@app.route("/", methods=['GET'])
def home():
    """
    Home endpoint for the Catalog service.

    Returns:
        Response: HTML welcome message.
    """
    return make_response("<h1 style='color:blue'>Welcome to the Catalog service!</h1>",200)

# route GraphQL
@app.route('/graphql', methods=['POST'])
def graphql_server():
    data = request.get_json()
    success, result = graphql_sync(
                        schema,
                        data,
                        context_value=None,
                        debug=app.debug
                    )
    status_code = 200 if success else 400
    return jsonify(result), status_code

if __name__ == "__main__":
    #p = sys.argv[1]
    print("Server running in port %s"%(PORT))
    app.run(host=HOST, port=PORT)
