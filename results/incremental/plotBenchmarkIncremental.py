#!/usr/bin/env python
# coding: utf-8

# In[28]:


import argparse
import matplotlib 
matplotlib.use('Agg')
import math
import datetime
import matplotlib.pyplot as plt
from matplotlib.ticker import FuncFormatter
import numpy as np
import os
import sys
import re

sys.path.insert(1, '../initial')

import custom_style

#get_ipython().run_line_magic('matplotlib', 'inline')


# In[29]:


inc_results_dir = "."

boosted = os.path.join(inc_results_dir, "boosted.txt")
dpf = os.path.join(inc_results_dir, "dpf.txt")
matrix = os.path.join(inc_results_dir, "matrix.txt")

all_files = [boosted, dpf, matrix]

col_num_rows = 0
col_update_size = 1
col_off_server = 2
col_off_client = 3
col_off_comm = 4
col_client_storage = 5
col_on_server = 6
col_on_client = 7
col_on_comm = 8


# In[30]:


# Compute cumulative costs
def compute_with_num_queries(data, num_queries):
    online_time = np.outer(data[:, col_on_server], num_queries)
    online_comm = np.outer(data[:, col_on_comm], num_queries)
    total_time = np.transpose([data[:,col_off_server]]*len(num_queries)) + online_time
    total_comm = np.transpose([data[:,col_off_comm]]*len(num_queries)) + online_comm
    per_query_time = total_time / num_queries
    per_change_comm = total_comm / np.transpose(([data[:, col_update_size]]*len(num_queries)))
    return per_query_time/1000, per_change_comm   


# In[31]:


# Server running time

boosted_data = np.genfromtxt(boosted, comments='#', skip_header=2)
dpf_data = np.genfromtxt(dpf, comments='#', skip_header=2)
matrix_data = np.genfromtxt(matrix, comments='#', skip_header=2)

num_queries = np.logspace(-7,7, base=2)

fig, ax = plt.subplots()

plt.ticklabel_format(style='plain')
ax.set_xscale('log', base=2)
ax.set_yscale('log')

def form(x,p):
    if x>=1: 
        return ('%f' % x).rstrip('0').rstrip('.')
    else:
        return ('1/%f' % int(1/x)).rstrip('0').rstrip('.')

f = FuncFormatter(matplotlib.ticker.FuncFormatter(form))
ax.xaxis.set_major_formatter(f)
ax.get_yaxis().set_major_formatter(matplotlib.ticker.FuncFormatter(lambda x,p: ('%f' % x).rstrip('0').rstrip('.')))

custom_style.remove_chart_junk(plt, ax, grid=True)


ax.set_yticks([0.1,1,10,100])
ax.set_xticks([1.0/100, 1.0/10,1,10,100])
plt.ylim(bottom=0.2, top=400)


linestyles = ["solid", "dashed", "dotted"]
colors=["red", "blue", "green", "purple"]



per_query_time, per_change_comm = compute_with_num_queries(boosted_data, num_queries)
# plt.plot(num_queries,per_query_time[0,:],
#          color=colors[0],
#          marker="o",
#          markevery=3,
#          linewidth=1,
#          label=('Checklist ($B=%d$)' % boosted_data[0,col_update_size]))
# plt.plot(num_queries,per_query_time[1,:],
#          color=colors[0],
#          marker="*",
#          markersize=5,
#          markevery=3,
#          linewidth=1,
#          label=('Checklist ($B=%d$)' % boosted_data[1,col_update_size]))
plt.plot(num_queries,per_query_time[2,:],
         color=colors[0],
         markevery=3,
         marker="d",
         linewidth=1,
         label=('Checklist'))


per_query_time, per_change_comm = compute_with_num_queries(dpf_data, num_queries)
plt.plot(num_queries,per_query_time[1,:],
         color=colors[1],
         linestyle="dashed",
         label='DPF')

per_query_time, per_change_comm = compute_with_num_queries(matrix_data, num_queries)
plt.plot(num_queries,per_query_time[1,:],
         color=colors[2],
         linestyle="dotted",
         label='Matrix')


plt.xlabel("Number of queries per period")
plt.ylabel("Amortized server time\nper query [ms]")
#ax.set_yticks([0.2, 1, 2, 4, 5])
plt.minorticks_off()


fig.legend(fontsize=6)
custom_style.save_fig(fig, "server.pdf", [2.3, 1.8])


# %%
