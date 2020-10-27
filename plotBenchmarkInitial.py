#!/usr/bin/python

import argparse
import matplotlib 
matplotlib.use('Agg')
import math
import matplotlib.pyplot as plt
import numpy as np
import os
import sys

def plot(in_names, cols, pretty_col_names, ylabel, out_name):
    fig, ax = plt.subplots()

    ax.set_xscale('log')
    ax.set_yscale('log')

    ax.tick_params('x', pad=0.5)

    linestyles = ["solid", "dashed", "dotted"]
    colors=["red", "blue"]

    for in_num, in_name in enumerate(in_names):
        pretty_in_name = os.path.splitext(os.path.basename(in_name))[0]
        results = np.genfromtxt(in_name, names=True, skip_header=1, skip_footer=1, usecols=cols)

        for idx, col_name in enumerate(results.dtype.names[1:]):
            plt.plot(results[results.dtype.names[0]],results[col_name]/1000, 
                "-o",
                color=colors[idx],
                linestyle=linestyles[in_num], 
                label=f'{pretty_col_names[idx+1]} ({pretty_in_name})')

        plt.xlabel(pretty_col_names[0])
        plt.ylabel(ylabel)
    fig.legend()
    plt.savefig(out_name)

parser = argparse.ArgumentParser(description='Plot benchmark results.')
parser.add_argument('input_files', metavar='input_files', type=str, nargs='+',
                   help='filenames of TSV benchmark results')
parser.add_argument('-o', 
                    dest='out_basename',
                    default='initial',
                    help='output file basename (default: \'initial\')')

args = parser.parse_args()



names = args.input_files
server_cols = [0, 1, 4]
client_cols = [0, 2 ,5]
comm_cols = [0, 3, 6]

pretty_col_names = ["Num Rows", "Offline", "Online"]

plot(names, server_cols, pretty_col_names, 'Server Running time (ms)', args.out_basename+"_server.pdf")
plot(names, client_cols, pretty_col_names, 'Client Running time (ms)', args.out_basename+"_client.pdf")
plot(names, comm_cols, pretty_col_names, 'Bytes sent', args.out_basename+"_comm.pdf")



