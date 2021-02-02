#!/usr/bin/python

import numpy as np
import os


def average(name):
    if os.path.exists(name):
        return "%d" % np.average([int(l) for l in open(name, "r")])
    else:
        return "-"

for name in ["boosted", "dpf", "nonprivate", "google"]:
    print("%20s:\t PERSISTENT HTTP: %7s, HTTP: %7s, TLS: %7s \n" % (name, average(name+"_persistent.txt"), average(name+".txt"), average(name+"_tls.txt")))

