#!/usr/bin/python

import argparse
import matplotlib 
matplotlib.use('Agg')
import math
import matplotlib.pyplot as plt
import numpy as np
import os
import sys
from matplotlib.ticker import FuncFormatter
import pylab

sys.path.insert(1, '../initial')

import custom_style


def plot(file_to_cols, scales, labels, out_name, legend=False):

    fig, ax = plt.subplots()

    plt.ticklabel_format(style='plain')
    ax.set_xscale(scales[0])
    ax.set_yscale(scales[1])

    # ax.set_xticks([86400*7*i for i in range(10)])
    # ax.set_yticks([10**i for i in range(2,7)])
    #ax.tick_params('x', pad=0.5)
    # ax.set_xlim([0, 10000])

    f = FuncFormatter(lambda x, pos: "%d\n{\\fontsize{6}{7}\\selectfont(%d%%)}"%(int(x*100), int(x*100)))
    ax.xaxis.set_major_formatter(f)

#    f = FuncFormatter(lambda x, pos: "$\\textsf{10}^\\textsf{%d}$" % round(math.log(x, 10)))
    #ax.yaxis.set_major_formatter(f)

    colors=["red", "blue", "green", "purple"]
    dots=[".", ".", ".", "."]

    for file_num, filename in enumerate(file_to_cols):
        results = np.genfromtxt(filename, names=True, comments='#', skip_header=1, usecols=file_to_cols[filename])

        changes = results[results.dtype.names[0]]
        changes = changes/2000000
        # changes[0] = 0
        cost = results[results.dtype.names[1]]
        
        xs = changes
        ys = cost/10**6

        avg = np.cumsum(ys)/ range(1,len(ys)+1)

        plt.plot(xs[0], ys[0], color=colors[file_num], marker='s',  linestyle = 'None', markersize='6', label='Initial setup')

        plt.plot(
            #results[results.dtype.names[0]],
            #results[col_name], 
            xs,
            ys,
            color=colors[file_num],
            linestyle="None",
            markersize = 2,
            marker = "X",
            label='Waterfall update')

        plt.plot(xs, avg, color="purple", linestyle='--', linewidth=1, label='Running average')


        ax.set_yticks([0.0001,0.001,0.01,0.1,1,10])
        #ax.set_ylim([None, 3])
        ax.set_xlim([0, None])
        plt.xlabel(labels[0])

        ax.get_yaxis().set_major_formatter(matplotlib.ticker.FuncFormatter(lambda x,p: ('%f' % x).rstrip('0').rstrip('.')))

        plt.ylabel(labels[1])
        plt.legend(loc="right", fontsize=6, bbox_to_anchor=(1.04,0.37))

        # all_labels = ax.get_legend_handles_labels()
        # #labels = [all_labels[0], ["Boosted PIR\n(this work)", "DPF", "Matrix"][0:len(all_labels[0])]]
        # plt.legend(all_labels, fontsize=6)

    custom_style.remove_chart_junk(plt, ax, grid=True)
    custom_style.save_fig(fig, out_name,  [2.3,1.8])


parser = argparse.ArgumentParser(description='Plot benchmark results.')
parser.add_argument('input_files', metavar='input_files', type=str, nargs='*',
                   help='filenames of TSV benchmark results')
parser.add_argument('-o', 
                    dest='out_basename',
                    default='updates',
                    help='output file basename (default: \'updates\')')

args = parser.parse_args()


names = args.input_files

if len(names) == 0:
    parser.print_help()
    exit(1) 

plot({name : [0, 1] for name in names}, 
    ["linear", "log"],
    ["Updates \\fontsize{6}{7}\\selectfont{(% DB changed)}", 'Server time (sec)'], 
    args.out_basename+"_server.pdf")

plot({name : [0, 2] for name in names}, 
    ["linear", "linear"],
    ["DB Changes", 'Client time (sec)'], 
    args.out_basename+"_client.pdf", legend=True)

plot({name : [0, 3] for name in names}, 
    ["linear", "log"],
    ["DB Changes", 'Communication (MB)'], 
    args.out_basename+"_comm.pdf")

"""
plot({name : [0, 4] for name in (names+no_offline_names)[0:1]},
    [""], 
    ["linear", "linear"],
    ["Num Rows", 'Client storage (MB)'], 
    args.out_basename+"_client_storage.pdf")
"""

