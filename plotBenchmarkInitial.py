#!/usr/bin/python

import argparse
import matplotlib 
matplotlib.use('Agg')
import math
import matplotlib.pyplot as plt
import numpy as np
import os
import sys

def plot(file_to_cols, pretty_col_names, labels, out_name):
    fig, ax = plt.subplots()

    ax.set_xscale('log')
    ax.set_yscale('log')

    ax.tick_params('x', pad=0.5)

    linestyles = ["solid", "dashed", "dotted"]
    colors=["red", "blue", "green"]

    for file_num, filename in enumerate(file_to_cols):
        pretty_name = os.path.splitext(os.path.basename(filename))[0]
        results = np.genfromtxt(filename, names=True, skip_header=1, skip_footer=1, 
            usecols=file_to_cols[filename])

        for idx, col_name in enumerate(results.dtype.names[1:]):
            plt.plot(results[results.dtype.names[0]],results[col_name]/1000, 
                "-o",
                color=colors[file_num],
                linestyle=linestyles[idx], 
                label=f'{pretty_name}{pretty_col_names[idx]}')

        plt.xlabel(labels[0])
        plt.ylabel(labels[1])
    fig.legend()
    plt.savefig(out_name)

parser = argparse.ArgumentParser(description='Plot benchmark results.')
parser.add_argument('input_files', metavar='input_files', type=str, nargs='+',
                   help='filenames of TSV benchmark results')
parser.add_argument('--no_offline', action='append')                   
parser.add_argument('-o', 
                    dest='out_basename',
                    default='initial',
                    help='output file basename (default: \'initial\')')

args = parser.parse_args()



names = args.input_files
no_offline_names = args.no_offline

plot({**{name : [0, 4, 1] for name in names}, 
    **{name : [0, 4] for name in no_offline_names}}, 
    ["", " (Offline)"], 
    ["Num Rows", 'Server Running time (ms)'], 
    args.out_basename+"_server.pdf")

plot({**{name : [0, 5, 2] for name in names}, 
    **{name : [0, 5] for name in no_offline_names}}, 
    ["", " (Offline)"], 
    ["Num Rows", 'Client Running time (ms)'], 
    args.out_basename+"_client.pdf")

plot({**{name : [0, 6, 3] for name in names}, 
    **{name : [0, 6] for name in no_offline_names}}, 
    ["", " (Offline)"], 
    ["Num Rows", 'Bytes sent'], 
    args.out_basename+"_comm.pdf")


