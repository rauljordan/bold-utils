from utils import pretty
from methods import *
from visualizations import render_assertion_chain

"""
Showcases a simple workflow of rendering the assertion chain and showing
any challenged assertions since the latest confirmed
"""
def simple_assertion_chain_workflow():
    # Render the assertion chain since the latest confirmed assertion's 
    # creation block so we can visualize it as a png
    render_assertion_chain(force_update=True)

    # Fetch all challenged assertions since the latest confirmed
    latest_confirmed = latest_confirmed_assertion()
    challenged_assertions = list_assertions({
        "challenged": "true",
        "from_block_number": latest_confirmed.get('creationBlock'),
    })
    print(f"Found {len(challenged_assertions)} challenged assertions since latest confirmed")
    for assertion in challenged_assertions:
        print(f"Hash={assertion.get('hash')}, inbox_max_count={assertion.get('inboxMaxCount')}")

"""
Showcases an investigation into an existing challenge, looking deeper into
the edges that are being created by the honest party, looking at subchallenges,
checking if our computed timers match up what is found onchain, etc. 
"""
def existing_challenge_inspection_workflow():
    hash = "0xd8b6fa197fcad2d1e14d203ebaf3b2e8bfa612685f8957b6a02d7eb760a98188"
    print(f"Investigating an existing challenge with parent assertion hash {hash}")
    challenged_assertion = assertion_by_hash(hash, {
        "force_update": "true",
    })
    print("Printing assertion that was the root of the challenge")
    print(pretty(challenged_assertion))

    # We get the root, block challenge, royal edges for the challenge above
    edges = list_edges(hash, {
        "force_update": "true",
        "royal": "true",
        "challenge_level": 0,
        "root_edges": "true",
    })
    print("Printing royal, block challenge root edges")

    # We filter out only a few relevant fields to make the output cleaner
    # To print out all fields, just do print(pretty(edges)) instead of filtering below
    fields = {
        'id', 
        'challengeLevel', 
        'status', 
        'isRoyal', 
        'startHeight', 
        'endHeight', 
        'claimId', 
        'cumulativePathTimer', 
        'timeUnrivaled', 
        'originId',
        'createdAtBlock',
        'hasLengthOneRival',
        'hasRival',
    }
    edges = [{key: edge[key] for key in edge if key in fields} for edge in edges]
    print(pretty(edges))

    # At the block challenge level, look at the computed, cumulative path timer
    # for our royal edges. We can then verify this computation by first: (1) checking
    # each edge's ancestor is royal, and (2) add up each ancestor's time unrivaled
    # and check it matches the response from our API
    edges = list_edges(hash, {
        "force_update": "true",
        "royal": "true",
        "challenge_level": 0,
    })

    # We group the edges by origin id, as an origin id is the unique 
    # identifier of a specific challenge
    edges_by_origin_id = {}
    for edge in edges:
        origin_id = edge.get('originId')
        if origin_id not in edges_by_origin_id:
            edges_by_origin_id[origin_id] = []
        edges_by_origin_id[origin_id].append(edge)

    edge_ids = {edge.get('id') for edge in edges}
    edges_by_id = {edge.get('id'): edge for edge in edges}

    # Check if each edge's computed ancestors are royal
    for origin_id, grouped_edges in edges_by_origin_id.items():
        for edge in grouped_edges:
            ancestors = edge.get('ancestors') or []
            for ancestor_id in ancestors:
                if ancestor_id not in edge_ids:
                    print(f"Ancestor with id {ancestor_id[:8]} not found in royal edges")
                    return

        print("All ancestors of fetched edges with origin id {origin_id} are confirmed royal")
        claimed_royal_assertion_hash = None

        for edge in grouped_edges:
            claim_id = edge.get('claimId')
            if claim_id != "":
                claimed_royal_assertion_hash = claim_id

        print(f"Claimed assertion hash of challenge {claimed_royal_assertion_hash}")
        assertion_unrivaled_time = challenged_assertion.get('secondChildBlock') - challenged_assertion.get('firstChildBlock')

        for edge in grouped_edges:
            path_timer = edge.get('cumulativePathTimer')
            computed_sum = edge.get('timeUnrivaled') + assertion_unrivaled_time
            ancestors = edge.get('ancestors') or []
            for ancestor in ancestors:
                computed_sum += edges_by_id[ancestor].get('timeUnrivaled')
            if computed_sum != path_timer:
                print(f"Edge with id {edge.get('id')[:8]} had wrong, computed timer {path_timer} vs expected {computed_sum}")

        print("All edges with origin id {origin_id} had the correct path timers")


if __name__ == "__main__":
    simple_assertion_chain_workflow()
    existing_challenge_inspection_workflow()
