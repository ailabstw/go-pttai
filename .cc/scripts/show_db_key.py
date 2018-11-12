#!/usr/bin/env python

import sys
import plyvel
import os


the_dir = sys.argv[1]

if the_dir[0] != '.':
    the_dir = os.path.sep.join(['.', the_dir])

db = plyvel.DB(the_dir, create_if_missing=False)

for k, v in db:
    print(k)

db.close()
