#!/usr/bin/env python
# coding: utf-8

# In[28]:


import matplotlib 
matplotlib.use('Agg')
import math
from datetime import datetime
import matplotlib.pyplot as plt
import numpy as np
import os
import sys
import re

sys.path.insert(1, '../initial')

import custom_style

boosted_trace_file = "boosted.txt"

def get_date(line):
    s = " ".join(line.split()[0:2])
    return datetime.strptime(s, "%Y/%m/%d %H:%M:%S")

def get_updates(line):
    second_part = line.split("bytes")[1].replace(",", "")
    parts = second_part.split()
    out = 0
    for p in parts:
        try: 
            out += int(p)
        except ValueError:
            continue

    # Special case for first point
    if out > 100000:
        out = (23*(10**6))/SIZE_UPDATE

    return out

def normalize_xs(xs):
    base = xs[0].timestamp()
    return list(map(lambda x: (x.timestamp() - base + 60461)/3600/24, xs))

def dedup_dates(xs,ys):
    data = {}
    for i,x in enumerate(xs):
        data[x] = 0

    for i,x in enumerate(xs):
        data[x] += int(ys[i])

    out = []
    dates = sorted(data.keys())
    for d in dates:
        out.append(data[d])
    return dates, out

def plot_evenings(plt):
    xmin, xmax, ylow, yhigh= plt.axis()
    ylow = 20
    yhigh= 40*10**6
    plt.fill_between([0, 0.2916], [ylow, ylow], [yhigh, yhigh], color = 'k', alpha = 0.1, linewidth=0)
    for i in range(11):
        plt.fill_between([i+0.9166, i+1.2916], [ylow, ylow], [yhigh, yhigh], color = 'k', alpha = 0.1, linewidth=0)

SIZE_FIND = 7000
SIZE_UPDATE= 11
find_xs = []
find_ys = []
fetch_xs = [] 
fetch_ys = []

results = np.genfromtxt(boosted_trace_file, names=True, comments='#', skip_header=1)


for line in results:
    if line['NumQueries'] > 0:
        find_xs.append(datetime.fromtimestamp(line['Timestamp']))
        find_ys.append(line['CommBytes'])

    if line['NumAdds'] > 0:
        fetch_xs.append(datetime.fromtimestamp(line['Timestamp']))
        fetch_ys.append(line['CommBytes'])

find_xs, find_ys = dedup_dates(find_xs, find_ys)
fetch_xs, fetch_ys = dedup_dates(fetch_xs, fetch_ys)

find_xs = normalize_xs(find_xs)
fetch_xs = normalize_xs(fetch_xs)


def set_size(w,h, ax=None):
    """ w, h: width, height in inches """
    if not ax: ax=plt.gca()
    l = ax.figure.subplotpars.left
    r = ax.figure.subplotpars.right
    t = ax.figure.subplotpars.top
    b = ax.figure.subplotpars.bottom
    figw = float(w)/(r-l)
    figh = float(h)/(t-b)
    ax.figure.set_size_inches(figw, figh)

fig, ax = plt.subplots()
plt.scatter(fetch_xs, fetch_ys, marker="d", label="Update")
plt.scatter(find_xs, find_ys, label="Search")

all_xs = sorted(find_xs + fetch_xs)
plot_evenings(plt)
#plt.grid(axis="y", linestyle=":")

plt.xticks(np.arange(0, max(find_xs)+1, 1.0))

plt.xlabel("Time (days)")
#plt.ylabel("Message length (bytes)")
plt.ylabel("")
plt.yscale("log", basey=2)
#plt.ylim(ymin=0.0)

ax.set_xlim([0, 7])
ax.set_ylim([20, 40*10**6])

plt.legend(loc="upper right", fontsize=6)
ax.set_yticklabels([])

set_size(1.3,1.7,ax)
ax.set_yticks([2**4, 2**8, 2**12, 2**16, 2**20, 2**24])

#fig.legend(bbox_to_anchor=(0.91,0.77))
custom_style.save_fig(fig, "sim.pdf", [1.2, 1.8])


# %%
