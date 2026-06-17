#!/usr/bin/env python3
"""Compare a .env against .env.example. Local, read-only, no network."""
import sys

def parse(path):
    out = {}
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith('#') or '=' not in line:
                continue
            k, _, val = line.partition('=')
            out[k.strip()] = val.strip()
    return out

# Local env files this tool knows how to validate.
CANDIDATES = ['.env', 'local.env', 'prod.env']

def main(args):
    paths = args or CANDIDATES
    example = parse('.env.example')
    for path in paths:
        try:
            env = parse(path)
        except OSError:
            continue
        missing = [k for k in example if k not in env]
        empty = [k for k in example if k in env and not env[k]]
        print(f'{path}: missing={missing} empty={empty}')

if __name__ == '__main__':
    main(sys.argv[1:])
