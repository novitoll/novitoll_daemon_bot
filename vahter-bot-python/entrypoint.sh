#!/usr/bin/env bash

set -e

python setup.py install

flask run --host=0.0.0.0 --port=5000