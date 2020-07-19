#!/usr/bin/python

import re
import string
import sys

param = sys.argv[1]

lines = sys.stdin.readlines()

headers = [param]
data = {}

constants = {}

for line in lines:
    if not re.search("Benchmark.*ns/op", line):
        continue
    fields = line.split()
    name = fields[0]
    time = fields[2]

    parts_of_name = re.split(param+"=([^,-]*)",name)
    if len(parts_of_name) < 3:
        sys.stderr.write("Suspicious line: " + line + "\n")
        constants[name] = time
    else:
        name = parts_of_name[0] + parts_of_name[2]
        param_val = int(parts_of_name[1])
        if not param_val in data:
            data[param_val] = {}    
        data[param_val][name] = time
    
    if not name in headers:
        headers += [name]
    

print("\t".join(headers))
for val in sorted(data):    
    data[val].update(constants)
    print("\t".join([str(val)]+[data[val][h] for h in headers[1:] ]))