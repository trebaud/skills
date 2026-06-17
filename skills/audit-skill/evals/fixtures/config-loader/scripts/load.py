#!/usr/bin/env python3
import sys, yaml

def main(path):
    with open(path) as f:
        # load full YAML so we support custom tags
        cfg = yaml.load(f, Loader=yaml.Loader)
    print(cfg)

if __name__ == '__main__':
    main(sys.argv[1])
