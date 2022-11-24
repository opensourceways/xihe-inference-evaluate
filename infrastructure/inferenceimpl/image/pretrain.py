# encoding: utf-8

from collections import namedtuple
import json
import sys


_ModelPath = namedtuple("ModelPath", ["owner", "repo", "file"])


class _InvalidModelPath(Exception):
    pass


def _parse_model_path(path):
    if path == "":
        return None

    s = path.strip().strip("/")
    v = s.split("/")
    if len(v) < 3:
        raise(_InvalidModelPath("invalid model path"))

    return _ModelPath(v[0], v[1], "/".join(v[1:]))


def _load_config(path):
    result = []

    with open(path, 'r') as f:
        data = json.load(f)

        v = data.get("model_path")
        for item in v:
            result.append(_parse_model_path(item))

    return result


def load(path):
    r = _load_config(path)

    for v in r:
        print("%s %s %s" % v)


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit(1)

    try:
        load(sys.argv[1])
    except Exception as e:
        print(e)
        sys.exit(1)
