#!/usr/bin/python

import argparse
import custom_style
import matplotlib 
matplotlib.use('Agg')
import math
import matplotlib.pyplot as plt
import numpy as np
import os
import sys
import pylab

def plot(file_to_cols, pretty_col_names, scales, labels, out_name, legend=False):
    fig, ax = plt.subplots()

    plt.ticklabel_format(style='plain')
    ax.set_xscale(scales[0])
    ax.set_yscale(scales[1])

    ax.tick_params('x', pad=0.5)
    ax.set_xlim([0, 10000])
    ax.set_ylim([100, 2*(10**6)])

    linestyles = ["solid", "dashed", "dotted"]
    colors=["red", "blue", "green", "purple"]
    dots=["-", "-", "-", "-"]

    for file_num, filename in enumerate(file_to_cols):
        pretty_name = os.path.splitext(os.path.basename(filename))[0]
        results = np.genfromtxt(filename, names=True, comments='#', skip_header=1, usecols=file_to_cols[filename])

        online_cost = results[results.dtype.names[1]][5]
        offline_cost = results[results.dtype.names[2]][5]
        
        print(offline_cost, online_cost)
        for idx, col_name in enumerate(results.dtype.names[1:]):

            xs = range(1, 10000)
            ys = []
            for i in xs:
                ys.append(offline_cost/float(i) + online_cost)

            plt.plot(
                #results[results.dtype.names[0]],
                #results[col_name], 
                xs,
                ys,
                dots[file_num],
                color=colors[file_num],
                linestyle=linestyles[file_num], 
                label=f'{pretty_name}{pretty_col_names[idx]}')

        plt.xlabel(labels[0])
        plt.ylabel(labels[1])

    custom_style.remove_chart_junk(plt, ax, grid=True)
    custom_style.save_fig(fig, out_name, [2, 1.6])
    if legend:
        figlegend = pylab.figure(figsize=(1.3,1.1))
        figlegend.legend(*ax.get_legend_handles_labels(), loc="center")
        figlegend.savefig("legend.pdf")


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
    ["Num Queries", 'Server time\namortized (µs)'], 
    args.out_basename+"_server.pdf")

plot({**{name : [0, 6, 2] for name in names}, 
    **{name : [0, 6] for name in no_offline_names}}, 
    ["", " (Offline)"], 
    ["linear", "log"],
    ["Num Queries", 'Client time\namortized (µs)'], 
    args.out_basename+"_client.pdf")

plot({**{name : [0, 7, 3] for name in names}, 
    **{name : [0, 7] for name in no_offline_names}}, 
    ["", " (Offline)"], 
    ["linear", "log"],
    ["Num Queries", 'Communication\namortized (bytes)'], 
    args.out_basename+"_comm.pdf", legend=True)

"""
plot({name : [0, 4] for name in (names+no_offline_names)[0:1]},
    [""], 
    ["linear", "linear"],
    ["Num Rows", 'Client storage (bytes)'], 
    args.out_basename+"_client_storage.pdf")
"""

