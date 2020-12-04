#!/usr/bin/python

import argparse
import matplotlib 
matplotlib.use('Agg')
import math
import matplotlib.pyplot as plt
import numpy as np
import os
import sys


linestyles = ["solid", "dashed", "dotted"]
colors=["red", "blue", "green", "purple"]


parser = argparse.ArgumentParser(description='Plot benchmark-updates results.')
parser.add_argument('-i',
                    dest='input_dir',
                    default='results/incremental',
                    help='directory containing TSV benchmark results')
parser.add_argument('-o', 
                    dest='out_basename',
                    default='incremental',
                    help='output file basename (default: \'initial\')')
parser.add_argument('-q', 
                    dest='num_queries',
                    type=int,
                    default=100,
                    help='scale num queries at each batch (default: 100)')

args = parser.parse_args()


boosted = os.path.join(args.input_dir, "boosted.tsv")
dpf = os.path.join(args.input_dir, "dpf.tsv")
matrix = os.path.join(args.input_dir, "matrix.tsv")

num_queries = args.num_queries

all_files = [boosted, dpf, matrix]

col_num_updates = 0
col_off_server = 1
col_off_client = 2
col_off_comm = 3
col_client_storage = 4
col_on_server = 5
col_on_client = 6
col_on_comm = 7

boosted_data = np.genfromtxt(boosted, comments='#', skip_header=2)
dpf_data = np.genfromtxt(dpf, comments='#', skip_header=2)
matrix_data = np.genfromtxt(matrix, comments='#', skip_header=2)

# Compute cumulative costs
for data in [boosted_data, dpf_data, matrix_data]:
    data[:, col_on_server:col_on_comm+1] *= num_queries
    data[:, col_off_server:col_off_comm+1] += data[:, col_on_server:col_on_comm+1] 
    data[:, col_off_server:col_off_comm+1] = np.cumsum(data[:, col_off_server:col_off_comm+1], axis=0)

# Server running time

fig, ax = plt.subplots()

ax.set_xscale('linear')
ax.set_yscale('linear')

ax.tick_params('x', pad=0.5)

labels = ['Boosted', 'DPF', 'Matrix']
for idx, data in enumerate([boosted_data, dpf_data, matrix_data]):
    plt.plot(data[:,col_num_updates],data[:,col_off_server],
            color=colors[idx],
            label=labels[idx])

plt.xlabel('Num DB Updates')
plt.ylabel('Total server running time (µs)')
plt.xlim(xmin=0.0)
plt.ylim(ymin=0.0)

fig.legend()
plt.savefig(args.out_basename+"_server.pdf")


# Client running time

fig, ax = plt.subplots()

ax.set_xscale('linear')
ax.set_yscale('linear')

ax.tick_params('x', pad=0.5)

labels = ['Boosted', 'DPF', 'Matrx']
for idx, data in enumerate([boosted_data, dpf_data, matrix_data]):
    plt.plot(data[:,col_num_updates],data[:,col_off_client],
            color=colors[idx],
            label=labels[idx])

plt.xlabel('Num DB Updates')
plt.ylabel('Total client running time (µs)')
plt.xlim(xmin=0.0)
plt.ylim(ymin=0.0)

fig.legend()
plt.savefig(args.out_basename+"_client.pdf")

# Communication cost

fig, ax = plt.subplots()

ax.set_xscale('linear')
ax.set_yscale('linear')

ax.tick_params('x', pad=0.5)

labels = ['Boosted', 'DPF', 'Matrx']
for idx, data in enumerate([boosted_data, dpf_data, matrix_data]):
    plt.plot(data[:,col_num_updates],data[:,col_off_comm],
            color=colors[idx],
            label=labels[idx])

plt.xlabel('Num DB Updates')
plt.ylabel('Total communication (bytes)')
plt.xlim(xmin=0.0)
plt.ylim(ymin=0.0)

fig.legend()
plt.savefig(args.out_basename+"_comm.pdf")

