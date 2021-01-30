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

def plot(files, labels, out_name, legend=False):

    fig, ax = plt.subplots()

    plt.ticklabel_format(style='plain')
    ax.set_xscale('log')
    ax.set_yscale('log')

    # ax.set_yticks([10**i for i in range(2,7)])
    # ax.tick_params('x', pad=0.5)
    # ax.set_xlim([0, 10000])
    # ax.set_ylim([100, 2*(10**6)])

    # f = FuncFormatter(lambda x, pos: int(x))
    # ax.xaxis.set_major_formatter(f)

    # f = FuncFormatter(lambda x, pos: "$\\textsf{10}^\\textsf{%d}$" % round(math.log(x, 10)))
    # ax.yaxis.set_major_formatter(f)

    linestyles = ["solid", "dashed", "dotted"]
    colors=["red", "blue", "green", "purple"]
    dots=["-", "-", "-", "-"]

    for file_num, filename in enumerate(files):
        results = np.genfromtxt(filename, names=True, comments='#')
        data = results[results.dtype.names[0]]
        n_bins = 50

        # plot the cumulative histogram
        n, bins, patches = plt.hist(data, n_bins, density=True, histtype='step',
                                cumulative=True, label='filename')

#        plt.axis([40000, np.max(data), 0, 1])

        # plt.plot(
        #     #results[results.dtype.names[0]],
        #     #results[col_name], 
        #     xs,
        #     ys,
        #     dots[file_num],
        #     color=colors[file_num],
        #     linestyle=linestyles[file_num], 
        #     label=f'{pretty_name}{pretty_col_names[idx]}')

        plt.xlabel(labels[0])
        plt.ylabel(labels[1])

    if legend:
        all_labels = ax.get_legend_handles_labels()
        labels = [[all_labels[0][i] for i in [0,2,4]], ["Checklist PIR\n(this work)", "DPF", "Non-private"]]
        plt.legend(*labels, fontsize=6)

#    custom_style.remove_chart_junk(plt, ax, grid=True)
    custom_style.save_fig(fig, out_name, [2.3, 1.6])
    if legend:
        figlegend = pylab.figure(figsize=(1.3,1.1))
        all_labels = ax.get_legend_handles_labels()
        labels = [[all_labels[0][i] for i in [0,2,4]], ["Boosted PIR\n(this work)", "DPF", "Matrix"]]
        figlegend.legend(*labels, loc="center")
        figlegend.savefig("legend.pdf")


parser = argparse.ArgumentParser(description='Plot latency benchmark results.')
parser.add_argument('input_files', metavar='input_files', type=str, nargs='*',
                   help='filenames of TSV benchmark results')

args = parser.parse_args()


names = args.input_files

if len(names) == 0:
    parser.print_help()
    exit(1) 

plot(names, ["Latency (ms)", 'CDF'], "latency.pdf")

