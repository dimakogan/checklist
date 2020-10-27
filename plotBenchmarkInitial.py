#!/usr/bin/python

import matplotlib 
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import os
import sys
import numpy as np
import math

def plot(in_names, cols, pretty_col_names, out_name):
    fig, ax = plt.subplots()

    ax.set_xscale('log')
    ax.set_yscale('log')

    ax.tick_params('x', pad=0.5)

    linestyles = ["solid", "dashed", "dotted"]
    colors=["red", "blue"]

    for in_num, in_name in enumerate(in_names):
        pretty_in_name = os.path.splitext(os.path.basename(in_name))[0]
        results = np.genfromtxt(in_name, names=True, skip_footer=1, usecols=cols)

        for idx, col_name in enumerate(results.dtype.names[1:]):
            plt.plot(results[results.dtype.names[0]],results[col_name]/1000, 
                "-o",
                color=colors[idx],
                linestyle=linestyles[in_num], 
                label=f'{pretty_col_names[idx+1]} ({pretty_in_name})')

        plt.xlabel(pretty_col_names[0])
        plt.ylabel('Running time (ms)')
    fig.legend()
    plt.savefig(out_name)

names = sys.argv[1:] 
server_cols = [0, 1, 3]
client_cols = [0, 2 ,4]

pretty_col_names = ["Num Rows", "Offline", "Online"]

plot(names, server_cols, pretty_col_names, "initial_server.pdf")
plot(names, client_cols, pretty_col_names, "initial_client.pdf")


