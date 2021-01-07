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


def plot(file_to_cols, pretty_col_names, scales, labels, out_name):
    fig, ax = plt.subplots()

    ax.set_xscale(scales[0])
    ax.set_yscale(scales[1])

    ax.tick_params('x', pad=0.5)

    for file_num, filename in enumerate(file_to_cols):
        pretty_name = os.path.splitext(os.path.basename(filename))[0]
        results = np.genfromtxt(filename, names=True, comments='#', skip_header=1, usecols=file_to_cols[filename])

        for idx, col_name in enumerate(results.dtype.names[1:]):
            plt.plot(results[results.dtype.names[0]],results[col_name], 
                "-o",
                color=colors[file_num],
                linestyle=linestyles[idx], 
                label=f'{pretty_name}{pretty_col_names[idx]}')

        plt.xlabel(labels[0])
        plt.ylabel(labels[1])
    fig.legend()
    plt.savefig(out_name)

parser = argparse.ArgumentParser(description='Plot benchmark results.')
parser.add_argument('-i',
                    dest='input_dir',
                    default='results/incremental',
                    help='directory containing TSV benchmark results')
parser.add_argument('-o', 
                    dest='out_basename',
                    default='incremental',
                    help='output file basename (default: \'initial\')')

args = parser.parse_args()


boosted = os.path.join(args.input_dir, "boosted.tsv")
dpf = os.path.join(args.input_dir, "dpf.tsv")
matrix = os.path.join(args.input_dir, "matrix.tsv")

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

plot({name : [col_num_rows, col_on_server] for name in all_files}, 
    ["", " (Offline)"], 
    ["linear", "log"],
    ["Num Rows", 'Server Running time (µs)'], 
    args.out_basename+"_server.pdf")

plot({name : [col_num_rows, col_on_client] for name in all_files}, 
    ["", " (Offline)"], 
    ["linear", "log"],
    ["Num Rows", 'Client Running time (µs)'], 
    args.out_basename+"_client.pdf")

plot({name : [col_num_rows, col_on_comm] for name in all_files}, 
    ["", " (Offline)"], 
    ["linear", "linear"],
    ["Num Rows", 'Bytes sent'], 
    args.out_basename+"_comm.pdf")


# Plot offline time as ratio to online times

boosted_times = np.genfromtxt(boosted, comments='#', skip_header=2)
dpf_times = np.genfromtxt(dpf, comments='#', skip_header=2)
matrix_times = np.genfromtxt(matrix, comments='#', skip_header=2)

fig, ax = plt.subplots()

ax.set_xscale('linear')
ax.set_yscale('linear')

ax.tick_params('x', pad=0.5)

# plt.plot(boosted_times[:,col_num_rows],boosted_times[:,col_off_server]/boosted_times[:,col_on_server], 
#         "-o",
#         color=colors[0],
#         label='Boosted')

# plt.plot(dpf_times[:,col_num_rows],boosted_times[:,col_off_server]/dpf_times[:,col_on_server], 
#         "-o",
#         color=colors[1],
#         label='DPF')

# plt.plot(matrix_times[:,col_num_rows],boosted_times[:,col_off_server]/matrix_times[:,col_on_server], 
#         "-o",
#         color=colors[2],
#         label='Matrix')


plt.xlabel('Update batch size')
plt.ylabel('Per Query Cost')

fig.legend()
plt.savefig(args.out_basename+"_inc.pdf")



