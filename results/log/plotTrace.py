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
from matplotlib.patches import Patch
from matplotlib.lines import Line2D

sys.path.insert(1, '../initial')

import custom_style


BOOSTED = 0
DPF = 1
NONPRIVATE = 2
GOOGLE = 3
ALL = 4

# File names
filenames = ["boosted.txt", "dpf.txt", "nonprivate.txt", "google.txt"]
labels = ["Checklist", "DPF", "Non-private", "Non-private"]
extra_labels = ["Offline", "Online"]

# Column numbers
COL_TIMESTAMP = 0
COL_ADDS = 1
COL_DELETES = 2
COL_QUERIES = 3
COL_SERVER_TIME = 4
COL_CLIENT_TIME = 5
COL_COMMUNICATION = 6


linestyles = ["solid", "solid", "solid"]
colors=["red", "blue", "grey", "grey"]
fill_colors = [plt.get_cmap('Reds')(x) for x in np.linspace(0,0.3,2)]
dots=["", "", "", ""]


def init_plot(ylabel, scales=['linear', 'linear'], ylim=None):
    fig, ax = plt.subplots()

    plt.ticklabel_format(style='plain')
    ax.set_xscale(scales[0])
    ax.set_yscale(scales[1])

    ax.set_xticks([86400*30*i for i in range(7)])
    ax.set_xlim([0, 86400*180])

    f = FuncFormatter(lambda x, pos: int(x/86400))
    ax.xaxis.set_major_formatter(f)

    plt.xlabel("Time (days)")
    plt.ylabel(ylabel)
    if scales[1] == 'linear':
        ax.set_ylim(bottom=0)
    if ylim:
        ax.set_ylim(top=ylim)

    return fig, ax


def plot(xs, ys, color, label, dots=""):
    plt.plot(
        xs[::10],
        ys[::10],
        dots,
        marker='.',
        markersize=2,
        color=color,
        linewidth=1,
        linestyle='None', #linestyles[file_num],
        label=label)

def save(fig, ax, out_name, legend=False):
    custom_style.remove_chart_junk(plt, ax, grid=True)
    custom_style.save_fig(fig, out_name, [2.3, 1.6])


def legend(ax):
    handles, labels = ax.get_legend_handles_labels()

    custom_lines = [Line2D([0], [0], color=h.get_color(), lw=2) for h in handles]
    figlegend1 = pylab.figure(figsize=(2.4,0.22))
    filled = [Patch(facecolor=fill_colors[i]) for i in [0,1]]

    figlegend1.legend(custom_lines[0:1]+filled, labels[0:1]+extra_labels, ncol=4, columnspacing=1, loc="center")
    figlegend1.savefig("legend1.pdf")

    figlegend2 = pylab.figure(figsize=(1.8,0.22))
    figlegend2.legend(handles=custom_lines[1:], labels=[ "DPF",  "Non-private"], markerfirst=False, markerscale=6, loc="center", ncol=2)
    figlegend2.savefig("legend2.pdf")

    # if legend:
    #     all_labels = ax.get_legend_handles_labels()
    #     plt.legend(*all_labels, fontsize=6, markerscale=2, handletextpad=0, loc="upper left"),  

    # if legend:
    #     figlegend = pylab.figure(figsize=(1.3,1.1))
    #     all_labels = ax.get_legend_handles_labels()
    #     figlegend.legend(*all_labels)
    #     figlegend.savefig("legend.pdf")

def stackplot(xs, ys):
    stacks = plt.stackplot(
        xs[::10],
        ys[:,::10],
        colors = fill_colors)
    # hatches=["..", "---","///////"]
    # for stack, hatch in zip(stacks, hatches):
    #     stack.set_hatch(hatch)

def read_results(filename, delimiter=None):
    return np.genfromtxt(filename, names=True, comments='#', skip_header=1, delimiter=delimiter)

def offline_cost(results,col_num):
    return results[results.dtype.names[col_num]]*(results[results.dtype.names[COL_QUERIES]]==0)


def online_cost(results,col_num):
    return results[results.dtype.names[col_num]]*(results[results.dtype.names[COL_QUERIES]]!=0)

def total_time(results):
    time = results[results.dtype.names[COL_TIMESTAMP]]
    return time[-1]-time[0]

def stacked(results, col_num):
    nonprivate = np.cumsum(offline_cost(results[NONPRIVATE], col_num))/10**6
    offline = np.cumsum(offline_cost(results[BOOSTED], col_num))/10**6
    total = np.cumsum(results[BOOSTED][results[BOOSTED].dtype.names[col_num]])/10**6
#    return np.array([nonprivate,offline-nonprivate,total-offline])
    return np.array([offline,total-offline])


def summarize_results(results):
    f = open("trace_summary.txt", "w")
    f.truncate()
    f.write("%15s%15s%15s%15s%15s%15s%15s%15s" % ("Type", "Weeks", "OffServer", "OnServer", "OffClient", "OnClient", "OffComm", "OnComm\n"))
    for i, result in enumerate(results):
        f.write("%15s" % filenames[i])
        f.write("%15.02f" % (total_time(results[i])/86400/7))
        f.write("%15.02f" % (np.sum(offline_cost(results[i], COL_SERVER_TIME))/10**6 ))
        f.write("%15.02f" % (np.sum(online_cost(results[i], COL_SERVER_TIME))/10**6 ))
        f.write("%15.02f" % (np.sum(offline_cost(results[i], COL_CLIENT_TIME))/10**6 ))
        f.write("%15.02f" % (np.sum(online_cost(results[i], COL_CLIENT_TIME))/10**6 ))
        f.write("%15.02f" % (np.sum(offline_cost(results[i], COL_COMMUNICATION))/10**6 ))
        f.write("%15.02f" % (np.sum(online_cost(results[i], COL_COMMUNICATION))/10**6 ))
        f.write("\n")

    initial_size = results[0][results[0].dtype.names[COL_ADDS]][0]
    added = np.sum(results[0][results[0].dtype.names[COL_ADDS]][1:])
    removed = np.sum(results[0][results[0].dtype.names[COL_DELETES]])
    f.write("\n# Starting size: %d, total added: %d, total removed: %d, final size: %d" %
        (initial_size, added, removed, initial_size+added-removed))
    f.close()


parser = argparse.ArgumentParser(description='Plot benchmark results.')
parser.add_argument('-o', 
                    dest='out_basename',
                    default='trace',
                    help='output file basename (default: \'trace\')')

args = parser.parse_args()

results = {}
for i in range(ALL):
    delim = "," if i == GOOGLE else None
    results[i] = read_results(filenames[i], delim)

timestamps = results[0][results[0].dtype.names[COL_TIMESTAMP]]
timestamps -= timestamps[0]

summarize_results(results)


fig, ax = init_plot('Communication\n(MB, cumulative)', scales=["linear", "linear"], ylim=75)
ys = stacked(results, COL_COMMUNICATION)
stackplot(timestamps, ys)
for i in [BOOSTED, DPF, GOOGLE]:
    y = np.cumsum(results[i][results[i].dtype.names[COL_COMMUNICATION]])/10**6
    plot(timestamps, y, colors[i], labels[i])
save(fig, ax, args.out_basename+"_comm.pdf")

fig, ax = init_plot('Server CPU time\n(sec, cumulative)', ylim=200)
ys = stacked(results, COL_SERVER_TIME)
stackplot(timestamps, ys)
for i in [BOOSTED, DPF, NONPRIVATE]:
    y = np.cumsum(results[i][results[i].dtype.names[COL_SERVER_TIME]])/10**6
    plot(timestamps, y, colors[i], labels[i])
save(fig, ax, args.out_basename+"_server.pdf")


fig, ax = init_plot('Client CPU time\n(sec, cumulative)', ylim=60)
ys = stacked(results, COL_CLIENT_TIME)
stackplot(timestamps, ys)
for i in [BOOSTED, DPF, NONPRIVATE]:    
    y = np.cumsum(results[i][results[i].dtype.names[COL_CLIENT_TIME]])/10**6
    plot(timestamps, y, colors[i], labels[i])
save(fig, ax, args.out_basename+"_client.pdf")

legend(ax)
