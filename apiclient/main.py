from utils import pretty
from methods import *
from visualizations import render_assertion_chain

def check_honest_party():
    latest_confirmed = latest_confirmed_assertion()
    challenged_assertions = list_assertions({
        "challenged": "true",
        "from_block_number": latest_confirmed.get('creationBlock'),
    })
    print(pretty(challenged_assertions))
    for assertion in challenged_assertions:
        edges = list_edges(assertion.get('hash'), {
            "force_update": "true",
            "royal": "true",
            "challenge_level": 0,
        })
        print("Printing all royal edges")
        print(pretty(edges))

def check_ministakes(assertion_hash):
    stakes = ministakes(assertion_hash, {
        "force_update": "true",
    })
    print(pretty(stakes))

def check_confirmable_royal_edges(assertion_hash):
    challenged_assertion = assertion_by_hash(assertion_hash, {
        "force_update": "true",
    })
    print(pretty(challenged_assertion))
    edges = list_edges(hash, {
        "force_update": "true",
        "royal": "true",
    })
    chal_period = challenged_assertion.get('confirmPeriodBlocks')
    for edge in edges:
        if edge.get('status') != 'confirmed' and edge.get('timeUnrivaled') >= chal_period:
            print(f"Found confirmable royal edge that is still pending {edge.get('id')}")

def check_unfinished_royal_subchallenges(assertion_hash):
    edges_with_claims = list_edges(assertion_hash, {
        "force_update": "true",
        "only_subchallenged_edges": "true",
        "royal": "true",
    })
    claim_ids = set()
    for edge in edges_with_claims:
        if edge.get('refersTo') == 'edge':
            claim_ids.add(edge.get('claimId'))

    for claim_id in claim_ids:
        edge = edge_by_id(hash, claim_id)        
        if edge.get('status') != 'confirmed':
            print(f"Unconfirmed, subchallenged royal edge: {claim_id}")

def assertion_hashes_since_latest_confirmed():
    latest_confirmed = latest_confirmed_assertion()
    all_assertions = list_assertions({
        "from_block_number": latest_confirmed.get('creationBlock'),
    })
    print(len(all_assertions))

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
    print(f"Latest confirmed inbox_max_count={latest_confirmed.get('inboxMaxCount')}")
    challenged_assertions = list_assertions({
        "challenged": "true",
        "from_block_number": latest_confirmed.get('creationBlock'),
    })
    print(f"Found {len(challenged_assertions)} challenged assertions since latest confirmed")
    for assertion in challenged_assertions:
        print(f"Hash={assertion.get('hash')}, inbox_max_count={assertion.get('inboxMaxCount')}")

    chal_manager_addrs = set()
    roots = set()
    all_assertions = list_assertions({
        "from_block_number": latest_confirmed.get('creationBlock'),
    })
    for assertion in all_assertions:
        chal_manager_addrs.add(assertion.get('challengeManager'))
        roots.add(assertion.get('wasmModuleRoot'))
    
    print(f"Found {len(all_assertions)} assertions since latest confirmed")
    print(f"Latest confirmed block number = {latest_confirmed.get('creationBlock')}")
    print(f"Latest created assertion={all_assertions[-1]}")

    for addr in chal_manager_addrs:
        print(f"chal_manager={addr}")
    for root in roots:
        print(f"wasm_module_root={root}")

def fetch_tracked():
    edges = list_tracked_royal_edges({
        "force_update": "true",
    })
    edges = edges[0].get('royalEdges')
    for edge in edges:
        print(pretty(edge))

"""
Showcases an investigation into an existing challenge, looking deeper into
the edges that are being created by the honest party, looking at subchallenges,
checking if our computed timers match up what is found onchain, etc. 
"""
def existing_challenge_inspection_workflow(assertion_hash):
    print(f"Investigating an existing challenge with parent assertion hash {assertion_hash}")
    challenged_assertion = assertion_by_hash(assertion_hash, {
        "force_update": "true",
    })
    print("Printing assertion that was the root of the challenge")
    print(pretty(challenged_assertion))

    # We get the root, block challenge, royal edges for the challenge above
    edges = list_edges(hash, {
        "force_update": "true",
        "royal": "true",
        "challenge_level": 3,
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

if __name__ == "__main__":
    simple_assertion_chain_workflow()