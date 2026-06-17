#!/usr/bin/env python3
import csv, sys

def main(path, col):
    vals = []
    with open(path, newline='') as f:
        for row in csv.DictReader(f):
            try:
                vals.append(float(row[col]))
            except (KeyError, ValueError):
                continue
    if not vals:
        print('no numeric values found'); return
    print(f'count={len(vals)} mean={sum(vals)/len(vals):.3f} min={min(vals)} max={max(vals)}')

if __name__ == '__main__':
    main(sys.argv[1], sys.argv[2])
