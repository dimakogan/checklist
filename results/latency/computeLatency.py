#!/usr/bin/python

import numpy as np

for name in ["boosted.txt", "dpf.txt", "nonprivate.txt", "google.txt"]:
    print("%s: %d\n" % (name, np.average([int(l) for l in open(name, "r")])))

