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


def plot(file_to_cols, scales, labels, out_name, add_y = 0, legend=False, ylim=None):

    fig, ax = plt.subplots()

    plt.ticklabel_format(style='plain')
    ax.set_xscale(scales[0])
    ax.set_yscale(scales[1])

    ax.set_xticks([86400*30*i for i in range(7)])
    #ax.set_ylim(bottom=10000)
    # ax.set_yticks([10**i for i in range(2,7)])
    # ax.tick_params('x', pad=0.5)
    ax.set_xlim([0, 86400*180])
    #ax.set_ylim([100, 2*(10**6)])

    f = FuncFormatter(lambda x, pos: int(x/86400))
    ax.xaxis.set_major_formatter(f)

#    f = FuncFormatter(lambda x, pos: "$\\textsf{10}^\\textsf{%d}$" % round(math.log(x, 10)))
    #ax.yaxis.set_major_formatter(f)

    linestyles = ["solid", "solid", "solid"]
    colors=["red", "blue", "grey", "purple"]
    dots=["", "", "", ""]

    for file_num, filename in enumerate(file_to_cols):
        try:
            results = np.genfromtxt(filename, names=True, comments='#', skip_header=1, usecols=file_to_cols[filename], invalid_raise=True)
        except:
            results = np.genfromtxt(filename, names=True, comments='#', skip_header=1, delimiter=',', usecols=file_to_cols[filename])

        timestamp = results[results.dtype.names[0]]
        cost = results[results.dtype.names[1]]
        
        xs = timestamp-timestamp[0]
        ys = np.cumsum(cost+add_y)/10**6

        print("Total %s for %s: %d"  % (results.dtype.names[1], filename, ys[-1]))

        plt.plot(
            #results[results.dtype.names[0]],
            #results[col_name], 
            xs[::10],
            ys[::10],
            dots[file_num],
            marker='.',
            markersize=2,
            color=colors[file_num],
            linewidth=1,
            linestyle='None', #linestyles[file_num],
            label=filename)

        plt.xlabel(labels[0])
        plt.ylabel(labels[1])

    ax.set_ylim(bottom=0)
    if ylim:
        ax.set_ylim(top=ylim)
    if legend:
        all_labels = ax.get_legend_handles_labels()
        labels = [[all_labels[0][i] for i in range(len(file_to_cols))], ["Checklist", "DPF", "Non-private"]]
        plt.legend(*labels, fontsize=6, markerscale=2, handletextpad=0, loc="upper left"),  

    custom_style.remove_chart_junk(plt, ax, grid=True)
    custom_style.save_fig(fig, out_name, [2.3, 1.6])
    if legend:
        figlegend = pylab.figure(figsize=(1.3,1.1))
        all_labels = ax.get_legend_handles_labels()
        labels = [[all_labels[0][i] for i in range(len(file_to_cols))], ["Checklist\n(this work)", "DPF", "Non-private"]]
        figlegend.legend(*labels)
        figlegend.savefig("legend.pdf")


parser = argparse.ArgumentParser(description='Plot benchmark results.')
parser.add_argument('input_files', metavar='input_files', type=str, nargs='*',
                   help='filenames of TSV benchmark results')
parser.add_argument('-o', 
                    dest='out_basename',
                    default='trace',
                    help='output file basename (default: \'trace\')')

args = parser.parse_args()


names = args.input_files

if len(names) == 0:
    parser.print_help()
    exit(1) 

plot({name : [0, 4] for name in names}, 
    ["linear", "linear"],
    ["Time (days)", 'Server CPU time\n(sec, cumulative)'], 
    args.out_basename+"_server.pdf", ylim=200,
    legend=True)

plot({name : [0, 6] for name in names}, 
    ["linear", "linear"],
    ["Time (days)", 'Communication\n(MB, cumulative)'], 
    args.out_basename+"_comm.pdf", ylim=120)

plot({name : [0, 5] for name in names}, 
    ["linear", "linear"],
    ["Time (days)", 'Client CPU time\n(sec, cumulative)'], 
    args.out_basename+"_client.pdf", ylim=60)

"""
plot({name : [0, 4] for name in (names+no_offline_names)[0:1]},
    [""], 
    ["linear", "linear"],
    ["Num Rows", 'Client storage (MB)'], 
    args.out_basename+"_client_storage.pdf")
"""

