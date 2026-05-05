import json
import sys

path = sys.argv[1] if len(sys.argv) > 1 else "results.json"

with open(path) as f:
    d = json.load(f)

m = d["metrics"]


def ms(metric, key):
    v = m.get(metric, {}).get("values", {}).get(key)
    return f"{v:.2f} ms" if v is not None else "n/a"


def rps(metric, key):
    v = m.get(metric, {}).get("values", {}).get(key)
    return f"{v:.2f} req/s" if v is not None else "n/a"


def pct(metric, key):
    v = m.get(metric, {}).get("values", {}).get(key)
    return f"{v * 100:.2f}%" if v is not None else "n/a"


def num(metric, key):
    v = m.get(metric, {}).get("values", {}).get(key)
    return str(int(v)) if v is not None else "n/a"


dur = "http_req_duration"
print(f"  latencia med   : {ms(dur, 'med')}")
print(f"  latencia p50   : {ms(dur, 'p(50)')}")
print(f"  latencia p90   : {ms(dur, 'p(90)')}")
print(f"  latencia p95   : {ms(dur, 'p(95)')}")
print(f"  latencia p99   : {ms(dur, 'p(99)')}")
print(f"  latencia max   : {ms(dur, 'max')}")
print()
print(f"  reqs/s         : {rps('http_reqs', 'rate')}")
print(f"  falhas         : {pct('http_req_failed', 'rate')}")
print(f"  erros schema   : {num('response_schema_errors', 'count')}")
print(f"  checks ok      : {pct('checks', 'rate')}")
