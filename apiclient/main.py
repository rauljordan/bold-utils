import requests
import json
from pygments import highlight, lexers, formatters

BASE_URL = "http://localhost:3000/api/v1"

def list_assertions(params=None):
    response = requests.get(f"{BASE_URL}/assertions", params=params)
    if response.status_code == 200:
        try:
            return json.loads(response.text)
        except json.JSONDecodeError:
            print("Error: Received response is not valid JSON.")
    print(f"Error: Received status code {response.status_code} and {response.text}")

def all_edges(assertion_hash, params=None):
    response = requests.get(f"{BASE_URL}/challenge/{assertion_hash}/edges", params=params)
    if response.status_code == 200:
        try:
            return json.loads(response.text)
        except json.JSONDecodeError:
            print("Error: Received response is not valid JSON.")
    print(f"Error: Received status code {response.status_code} and {response}")

def pretty(json_data):
    json_str = json.dumps(json_data, indent=4, sort_keys=True)
    colorful_json = highlight(json_str, lexers.JsonLexer(), formatters.TerminalFormatter())
    return colorful_json

def chal_hash(json_response):
    first_item = json_response[0]
    # Check if 'hash' field exists in the first item
    hash_value = first_item.get('hash')
    if hash_value is not None:
        return hash_value
    else:
        return "Hash field not found in the first item."

if __name__ == "__main__":
    assertions = list_assertions()
    challenge_hash = chal_hash(assertions)
    edges = all_edges(challenge_hash, {
        "force_update": "true",
        "royal": "true",
    })
    edges_by_id = {edge['id']: edge for edge in edges if 'id' in edge}
    lowermost = edges[-1]
    print(pretty(lowermost))

    # Check if each edge's computed ancestors are royal
    ancestors = lowermost.get('ancestors', [])
    for an in ancestors:
        found = False
        for edge in edges:
            if edge.get('id') == an:
                found = True
        if not found:
            print(f"Ancestor with id {an[:8]} not found in royal edges")
    
    # Next, check if each royal edge's cumulative path timer 
    # is equal to the time unrivaled of it and its ancestors
    for edge in edges:
        path_timer = edge.get('cumulativePathTimer')
        computed_sum = edge.get('timeUnrivaled')
        ancestors = edge.get('ancestors')
        for ancestor in ancestors:
            if ancestor not in edges_by_id:
                continue
            else:
                computed_sum += edges_by_id[ancestor].get('timeUnrivaled')
        if computed_sum != path_timer:
            print(f"Edge with id {edge.get('id')[:8]} had computed timer {computed_sum} vs {path_timer}")
