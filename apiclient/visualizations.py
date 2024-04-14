from methods import latest_confirmed_assertion, list_assertions
import graphviz

def render_assertion_chain(render_path="graph", force_update=False):
    latest_confirmed = latest_confirmed_assertion()
    print(f"Latest confirmed assertion: inbox_max_count={latest_confirmed.get('inboxMaxCount')}, " +
          f"creation_block={latest_confirmed.get('creationBlock')}")
    print("Fetching all assertions created since...")
    all_assertions = list_assertions({
        "force_update": "true" if force_update else "false",
        "from_block_number": latest_confirmed.get('creationBlock'),
    })
    if not(all_assertions):
        print("No assertions created since latest confirmed")
        return
    else: 
        print(f"{len(all_assertions)} assertions created since latest confirmed")
        last_assertion = all_assertions[-1]
        print(f"Latest created: inbox_max_count={last_assertion.get('inboxMaxCount')}, " +
          f"creation_block={last_assertion.get('creationBlock')}")
        block_diff = last_assertion.get('creationBlock')-latest_confirmed.get('creationBlock')
        print(f"{block_diff} blocks since the latest confirmed assertion")
        dot = graphviz.Digraph(comment='Assertion Chain', format='png')
        for item in all_assertions:
            label = f"Hash: {item['hash'][:8]}\nInboxMaxCount: {item['inboxMaxCount']}"
            dot.node(item['hash'], label)
            if item['parentAssertionHash']:
                dot.edge(item['parentAssertionHash'], item['hash'])
        print(f"Rendering graphviz png to file: {render_path}")
        dot.render(render_path)