#!/usr/bin/env python
# coding: utf-8

# In[28]:


import matplotlib 
matplotlib.use('Agg')
import math
import datetime
import matplotlib.pyplot as plt
import numpy as np
import os
import sys
import re

sys.path.insert(1, '../initial')

import custom_style

def get_date(line):
    s = " ".join(line.split()[0:2])
    return datetime.datetime.strptime(s, "%Y/%m/%d %H:%M:%S")

def get_size(line):
    return int(line.split("Size: ")[1].split()[0])

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
    ylow = 60
    yhigh= 8*10**6
    plt.fill_between([0, 0.2916], [ylow, ylow], [yhigh, yhigh], color = 'k', alpha = 0.1, linewidth=0)
    for i in range(11):
        plt.fill_between([i+0.9166, i+1.2916], [ylow, ylow], [yhigh, yhigh], color = 'k', alpha = 0.1, linewidth=0)

find_xs = []
find_ys = []
fetch_xs = [] 
fetch_ys = []
for line in sys.stdin:
    if "FIND Request" in line or "FIND Response" in line:
        find_xs.append(get_date(line))
        find_ys.append(get_size(line))

    if "FETCH Request" in line or "FETCH Response" in line:
        fetch_xs.append(get_date(line))
        fetch_ys.append(get_size(line))

find_xs, find_ys = dedup_dates(find_xs, find_ys)
fetch_xs, fetch_ys = dedup_dates(fetch_xs, fetch_ys)

find_xs = normalize_xs(find_xs)
fetch_xs = normalize_xs(fetch_xs)

fig, ax = plt.subplots()
plt.scatter(find_xs, find_ys)
plt.scatter(fetch_xs, fetch_ys)
ax.set_xlim([0, None])
ax.set_ylim([60, 8*10**6])

all_xs = sorted(find_xs + fetch_xs)
plot_evenings(plt)

plt.xticks(np.arange(0, max(find_xs)+1, 1.0))

plt.xlabel("Time (days)")
plt.ylabel("Message length (bytes)")
plt.yscale("log")
#plt.ylim(ymin=0.0)

#fig.legend(bbox_to_anchor=(0.91,0.77))
custom_style.save_fig(fig, "log.pdf", [3.5, 2.1])


# %%
