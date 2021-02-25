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
ALL = 3

WINDOW_SIZE_SEC = 5

# File names
throughput_filenames = ["boosted.txt", "dpf.txt", "nonprivate.txt"]
latency_filenames = ["boosted_latency.txt", "dpf_latency.txt", "nonprivate_latency.txt"]
labels = ["Checklist", "DPF", "Non-private"]

# Column names
# Seconds,Workers,Queries,Latency,Errors
# Time,Latency


linestyles = ["solid", "solid", "solid"]
colors=["red", "blue", "grey", "grey"]
fill_colors = [plt.get_cmap('Reds')(x) for x in np.linspace(0,0.3,2)]
dots=["", "", "", ""]


def init_plot(ylabel, scales=['linear', 'linear'], ylim=None):
    fig, ax = plt.subplots()

    plt.ticklabel_format(style='plain')
    ax.set_xscale(scales[0])
    ax.set_yscale(scales[1])

    if scales[0] == 'linear':
        ax.set_xticks([1000*i for i in range(6)])
    if scales[1] == 'linear':
        ax.set_yticks([60*i for i in range(8)])
        ax.get_yaxis().set_major_formatter(matplotlib.ticker.FuncFormatter(lambda x,p: ('%f' % x).rstrip('0').rstrip('.')))
 
    plt.xlabel("Throughput (requests/second)")
    plt.ylabel("Latency (msec)")
    # if scales[1] == 'linear':
    #     ax.set_ylim(bottom=0)
    # if ylim:
    #     ax.set_ylim(top=ylim)

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
    plt.legend(fontsize=6, bbox_to_anchor=(0.1,0.5))


def read_results(filename):
    return np.genfromtxt(filename, names=True, comments='#', delimiter=',')


fig, ax = init_plot('Latency (msec)', scales=["linear", "linear"])
#fig, ax = plt.subplots()

for i in [0,1,2]:
    throughput = read_results(throughput_filenames[i])
    latency = read_results(latency_filenames[i])

    throughputs = {}
    for row in throughput:
        time = row['Seconds']
        reqs = row['Queries']
        if (time not in throughputs) or (throughputs[time] < reqs):
            throughputs[time] = reqs

    latencies = {}
    for row in latency:
        time = row['Seconds']
        lats = []
        if time in latencies:
            lats = latencies[time]
        lats += [row['Latency']]
        latencies[time] = lats

    windows = {}
    time2window = {}
    for row in throughput:
        num_workers = row['Workers']
        time = row['Seconds']
        time2window[time] = num_workers
        if num_workers not in windows:
            windows[num_workers] = (time,time)
            continue
        times = windows[num_workers]
        if time < times[0]:
            windows[num_workers] = (time,times[1])
        elif times[1] < time:
            windows[num_workers] = (times[0],time)

    window2throughput = {}
    for w,(start,end) in windows.items():
        window2throughput[w] = (throughputs[end]-throughputs[start])/(end-start)
    
    window2latencies = {}
    for time, window in time2window.items():
        lats = []
        tp = 0
        if window in window2latencies:
            lats = window2latencies[window]
        if time in latencies:
            lats += latencies[time]
        window2latencies[window] = lats

    window_avg_latency = []
    window_90th_latency = []
    window_throughput = []
    for w in sorted(window2latencies.keys()):
        tp = window2throughput[w]
        if len(window_throughput)>0 and tp < window_throughput[-1]:
            continue
        # if len(window_throughput)>0 and tp < window_throughput[-1]*1.01:
        #     continue
        ls = window2latencies[w]
        l90 = int(np.percentile(ls, 90))
        lavg = int(np.average(ls))

        if l90 > 500:
            continue

        while len(window_throughput)>0 and (l90 <= window_90th_latency[-1]):
            window_throughput = window_throughput[:-1]
            window_avg_latency = window_avg_latency[:-1]
            window_90th_latency = window_90th_latency[:-1]

        window_throughput += [tp]
        window_90th_latency += [l90]
        window_avg_latency += [lavg]
#        print(f"{i}, {w}, {tp}, {l90}")

    plt.plot(
        window_throughput,
        window_90th_latency,
        # dots,
        marker='o',
        # markersize=2,
        linewidth=1,
        color=colors[i],
        label=labels[i])

legend(ax)
save(fig, ax, "stress.pdf")

