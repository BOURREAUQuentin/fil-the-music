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

track = ObjectType('Track')

query.set_field('track_json', r.track_json)

schema = make_executable_schema(type_defs, track, query)

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
