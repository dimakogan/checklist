#!/usr/bin/python

import argparse
import custom_style
import matplotlib 
matplotlib.use('Agg')
import math
import matplotlib.pyplot as plt
import matplotlib.markers as markers
import numpy as np
import os
import sys
from matplotlib.ticker import FuncFormatter
import pylab
from matplotlib.lines import Line2D

def plot(file_to_cols, pretty_col_names, scales, labels, out_name, legend=False):

    fig, ax = plt.subplots()

    plt.ticklabel_format(style='plain')
    ax.set_xscale(scales[0])
    ax.set_yscale(scales[1])

    ax.tick_params('x', pad=0.5)
    ax.set_xlim([0, 10000])
    #ax.set_ylim([100, 2*(10**6)])

    f = FuncFormatter(lambda x, pos: int(x))
    ax.xaxis.set_major_formatter(f)

    f = FuncFormatter(lambda x, pos: "${10}^{%d}$" % round(math.log(x, 10)))
    ax.yaxis.set_major_formatter(f)

    linestyles = ["solid", "dashed", "dotted"]
    colors=["red", "blue", "green", "purple"]
    dots=["-", "-", "-", "-"]

    for file_num, filename in enumerate(file_to_cols):
        pretty_name = os.path.splitext(os.path.basename(filename))[0]
        results = np.genfromtxt(filename, names=True, comments='#', skip_header=1, usecols=file_to_cols[filename])

        online_cost = np.array([results[results.dtype.names[1]]])
        offline_cost = np.array([results[results.dtype.names[2]]])

        xs = range(1, 10000)
        ys = []
        for i in xs:
            ys.append(offline_cost/float(i) + online_cost)
        print(online_cost)
        
        plt.plot(
            #results[results.dtype.names[0]],
            #results[col_name], 
            xs,
            [online_cost]*len(xs),
            dots[file_num],
                linewidth=1,
            color=colors[file_num],
            linestyle=linestyles[file_num], 
            label=filename)

        if offline_cost > online_cost*10:
            plt.plot(xs[0], ys[0], color=colors[file_num], marker=markers.CARETRIGHTBASE,  linestyle = 'None', markersize='6', label='Initial setup')
            plt.plot(
                #results[results.dtype.names[0]],
                #results[col_name], 
                xs,
                ys,
                #dots[file_num],
                color=colors[file_num],
                linewidth=1,
                linestyle="solid", 
                marker="o",
                markevery=500,
                label=filename)

        plt.xlabel(labels[0])
        plt.ylabel(labels[1])

    ax.set_yticks([10**i for i in range(1,7)])
    
    # if legend:
    #     all_labels = ax.get_legend_handles_labels()
    #     labels = [all_labels[0], prettyLabels]
    #     plt.legend(*labels, fontsize=6)

    custom_style.remove_chart_junk(plt, ax, grid=True)
    custom_style.save_fig(fig, out_name+".pdf", width=2.2)
    custom_style.save_fig(fig, out_name+".pgf", width=2.2)
    if legend:
        handles, labels = ax.get_legend_handles_labels()

        figlegend1 = pylab.figure(figsize=(4,0.22))
        dummy = Line2D([0], [0], linewidth=0, linestyle=None)
        figlegend1.legend(handles=[dummy]+handles[0:3], labels=["Offline-Online", "online", "offline", "amortized"], loc="center", ncol=4)
        figlegend1.savefig("legend1.pdf")
        figlegend1.savefig("legend1.pgf")

        figlegend2 = pylab.figure(figsize=(1.54,0.22))
        figlegend2.legend(handles=handles[3:5], labels=[ "DPF",  "Matrix"], markerfirst=False, loc="center", ncol=2)
        figlegend2.savefig("legend2.pdf")
        figlegend2.savefig("legend2.pgf")


parser = argparse.ArgumentParser(description='Plot benchmark results.')
parser.add_argument('input_files', metavar='input_files', type=str, nargs='*',
                   help='filenames of TSV benchmark results')
parser.add_argument('--no_offline', action='append')                   
parser.add_argument('-o', 
                    dest='out_basename',
                    default='initial',
                    help='output file basename (default: \'initial\')')

args = parser.parse_args()


names = args.input_files
no_offline_names = args.no_offline
if no_offline_names == None:
    no_offline_names = []

if len(names)+len(no_offline_names) == 0:
    parser.print_help()
    exit(1) 

plot({**{name : [0, 5, 1] for name in names}, 
    **{name : [0, 5] for name in no_offline_names}}, 
    ["", " (Offline)"], 
    ["linear", "log"],
    ["Num Queries", 'Server time (µs)'], 
    args.out_basename+"_server")

plot({**{name : [0, 6, 2] for name in names}, 
    **{name : [0, 6] for name in no_offline_names}}, 
    ["", " (Offline)"], 
    ["linear", "log"],
    ["Num Queries", 'Client time (µs)'], 
    args.out_basename+"_client", legend=True)

plot({**{name : [0, 7, 3] for name in names}, 
    **{name : [0, 7] for name in no_offline_names}}, 
    ["", " (Offline)"], 
    ["linear", "log"],
    ["Num Queries", 'Communication (bytes)'], 
    args.out_basename+"_comm")

"""
plot({name : [0, 4] for name in (names+no_offline_names)[0:1]},
    [""], 
    ["linear", "linear"],
    ["Num Rows", 'Client storage (bytes)'], 
    args.out_basename+"_client_storage.pdf")
"""

