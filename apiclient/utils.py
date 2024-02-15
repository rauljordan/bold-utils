import json
from pygments import highlight, lexers, formatters

def pretty(json_data):
    json_str = json.dumps(json_data, indent=4, sort_keys=True)
    colorful_json = highlight(json_str, lexers.JsonLexer(), formatters.TerminalFormatter())
    return colorful_json