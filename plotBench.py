#!/usr/bin/python

import matplotlib 
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import sys
import numpy as np
import math

out_name = sys.argv[1] 
in_name = sys.argv[2]

fig, ax = plt.subplots()

ax.set_xscale('log')
ax.set_yscale('log')

ax.tick_params('x', pad=0.5)

results = np.genfromtxt(in_name, names=True)

for col_name in results.dtype.names[1:]:
    plt.plot(results[results.dtype.names[0]],results[col_name]/10**6, "-o", label=col_name)

plt.xlabel(results.dtype.names[0])
plt.ylabel('Running time (ms)')
fig.legend()
plt.savefig(out_name)
