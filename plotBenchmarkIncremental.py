#!/usr/bin/env python
# coding: utf-8

# In[28]:


import argparse
import matplotlib 
matplotlib.use('Agg')
import math
import datetime
import matplotlib.pyplot as plt
import numpy as np
import os
import sys
import re

#get_ipython().run_line_magic('matplotlib', 'inline')


# In[29]:


inc_results_dir = "results/incremental"

boosted = os.path.join(inc_results_dir, "boosted.tsv")
dpf = os.path.join(inc_results_dir, "dpf.tsv")
matrix = os.path.join(inc_results_dir, "matrix.tsv")

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

num_queries = np.linspace(1,20)

fig, ax = plt.subplots()

ax.set_xscale('linear')
ax.set_yscale('linear')

linestyles = ["solid", "dashed", "dotted"]
colors=["red", "blue", "green", "purple"]

per_query_time, per_change_comm = compute_with_num_queries(boosted_data, num_queries)
plt.plot(num_queries,per_query_time[6,:],
         color=colors[0],
         label=('Boosted (B=%d)' % boosted_data[6,col_update_size]))
plt.plot(num_queries,per_query_time[3,:],
         color=colors[0],
         linestyle='dotted',
         label=('Boosted (B=%d)' % boosted_data[3,col_update_size]))
plt.plot(num_queries,per_query_time[0,:],
         color=colors[0],
         linestyle='dashed',
         label=('Boosted (B=%d)' % boosted_data[0,col_update_size]))


per_query_time, per_change_comm = compute_with_num_queries(dpf_data, num_queries)
plt.plot(num_queries,per_query_time[3,:],
         color=colors[1],
         label='DPF')

per_query_time, per_change_comm = compute_with_num_queries(matrix_data, num_queries)
plt.plot(num_queries,per_query_time[3,:],
         color=colors[2],
         label='Matrix')


plt.xlabel("Number of queries per period")
plt.ylabel("Amortized server time per query [ms]")
plt.xlim(xmin=0.0)
plt.ylim(ymin=0.0)

fig.legend()
plt.show()
plt.savefig(os.path.join(inc_results_dir, "server.pdf"))


# %%
