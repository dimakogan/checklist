import custom_style
from custom_style import setup_columns,col
import matplotlib.pyplot as plt
import sys
import numpy as np
from matplotlib.ticker import FuncFormatter
import math

out_name = sys.argv[1] 
in_name = sys.argv[2]

fig, ax = plt.subplots()

def get_data(col):
    x = []
    y = []
    with open(in_name, 'r') as f:
        for i, line in enumerate(f):
            if i == 0: continue
            p = line.split()
            if col >= len(p): continue
            xpoint = float(p[0])
            ypoint = float(p[col])/(10**6)

            if ypoint <0: continue 
            x.append(xpoint)
            y.append(ypoint)

    return (x,y)


#ax.set_yscale('log')
ax.set_xscale('log')

ax.xaxis.set_major_formatter(FuncFormatter(lambda x, pos: "$2^{%d}$" % int(math.log(max(x,1),2))))
ax.yaxis.set_major_formatter(FuncFormatter(lambda y, pos: "%d$" % y))
ax.set_xticks([2**i for i in range(12,24)])
ax.set_xlim([2**16,2**24 / 1.8])
ax.tick_params('x', pad=0.5)
#ax.set_ylim([1, 2.5*10**3])
#ax.set_yticks([0, 40, 80, 120])

(x,y) = get_data(1)
plt.plot(x, y, "-o", label="No robustness")
#, ":", label="Ideal", color="gray")

(x,y) = get_data(2)
plt.plot(x, y, "-o", label="Prio-MPC")
#, ":", label="Ideal", color="gray")


custom_style.remove_chart_junk(plt, ax,grid=True)

#plt.title("Comparison of Mixing Methods")
plt.xlabel('Database table size (96-byte rows)')
plt.ylabel('Running time (ms)')

#plt.legend(loc='upper right', frameon=False, bbox_to_anchor=(0.1, 1.02, 1., .102), ncol=2)
custom_style.save_fig(fig, out_name, [3, 1.75])
#plt.show()
