import requests
import json

BASE_URL = "http://localhost:7257/api/v1"

"""
List all assertions up to the chain head

Query parameters (optional):
  - limit: the max number of items in the response
  - offset: the offset index in the DB
  - inbox_max_count: assertions that have a specified value for InboxMaxCount
  - from_block_number: items that were created since a specific block number. Defaults to latest confirmed assertion
  - to_block_number: caps the response to assertions up to and including a block number
  - challenged: fetch only assertions that have been challenged
  - force_update: refetch the updatable fields of each item in the response
"""
def list_assertions(params=None):
    return process_response(requests.get(f"{BASE_URL}/assertions", params=params))

"""
Get an assertion by its identifier

Parameters:
  - hash: an assertion hash or "latest-confirmed"

Query parameters (optional):
  - force_update: refetch the updatable fields of the assertion
"""
def assertion_by_hash(hash, params=None):
    return process_response(requests.get(f"{BASE_URL}/assertions/{hash}", params=params))

"""
Get the latest confirmed assertion at chain head
"""
def latest_confirmed_assertion():
    return process_response(requests.get(f"{BASE_URL}/assertions/latest-confirmed"))

"""
Gets all the edge ministakes in a challenge
"""
def ministakes(assertion_hash, params=None):
    return process_response(requests.get(f"{BASE_URL}/challenge/{assertion_hash}/ministakes", params=params))

"""
Get all the edges corresponding to a challenged assertion with a specific hash

Parameters:
  - assertion_hash: the assertion hash (0x-prefixed)

Query parameters (optional):
  - limit: the max number of items in the response
  - offset: the offset index in the DB
  - status: filter edges that have status "confirmed", "confirmable", or "pending"
  - royal: boolean true or false to get royal edges. If not set, fetches all edges in the challenge.
  - root_edges: boolean true or false to filter out only root edges (those that have a claim_id)
  - rivaled: boolean true or false to get only rivaled edges
  - has_length_one_rival: boolean true or false to get only edges that have a length one rival
  - only_subchallenged_edges: boolean true or false to get only edges that have a subchallenge claiming them
  - from_block_number: items that were created since a specific block number.
  - to_block_number: caps the response to edges up to a block number
  - path_timer_geq: edges with a path timer greater than some N number of blocks
  - origin_id: edges that have a 0x-prefixed origin id
  - mutual_id: edges that have a 0x-prefixed mutual id
  - claim_id: edges that have a 0x-prefixed claim id
  - start_height: edges with a start height
  - end_height: edges with an end height
  - start_commitment: edges with a start history commitment of format "height:hash", such as 32:0xdeadbeef
  - end_commitment: edges with an end history commitment of format "height:hash", such as 32:0xdeadbeef
  - challenge_level: edges in a specific challenge level. level 0 is the block challenge level
  - force_update: refetch the updatable fields of each item in the response
"""
def list_edges(assertion_hash, params=None):
    # print(f"{BASE_URL}/challenge/{assertion_hash}/edges")
    return process_response(requests.get(f"{BASE_URL}/challenge/{assertion_hash}/edges", params=params))

"""
Fetches an edge by its specific id in a challenge

Parameters:
  - assertion_hash: the assertion hash (0x-prefixed)
  - edge_id: the edge id (0x-prefixed)

Query parameters (optional):
  - force_update: refetch the updatable fields of the edge
"""
def edge_by_id(assertion_hash, edge_id, params=None):
    return process_response(requests.get(f"{BASE_URL}/challenge/{assertion_hash}/edges/id/{edge_id}", params=params))

"""
Dumps the locally-tracked, royal edges kept in-memory by the BOLD validator. If a
challenge has completed, it will no longer be tracked in-memory. Instead, prefer list_edges for 
most calls
"""
def list_tracked_royal_edges(params=None):
    return process_response(requests.get(f"{BASE_URL}/tracked/royal-edges", params=params))

def process_response(response):
    if response.status_code == 200:
        try:
            return json.loads(response.text)
        except json.JSONDecodeError:
            print("Error: Received response is not valid JSON.")
    print(f"Error: Received status code {response.status_code} and {response}")
