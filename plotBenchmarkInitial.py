#!/usr/bin/python

import matplotlib 
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import os
import sys
import numpy as np
import math

def plot(in_name, cols, out_name):
    fig, ax = plt.subplots()

    ax.set_xscale('log')
    ax.set_yscale('log')

    ax.tick_params('x', pad=0.5)

    results = np.genfromtxt(in_name, names=True, skip_footer=1, usecols=cols)

    for col_name in results.dtype.names[1:]:
        plt.plot(results[results.dtype.names[0]],results[col_name]/1000, "-o", label=col_name)

    plt.xlabel(results.dtype.names[0])
    plt.ylabel('Running time (ms)')
    fig.legend()
    plt.savefig(out_name)

name = sys.argv[1] 
server_cols = [0, 1, 3]
client_cols = [0, 2 ,4]

plot(name, server_cols, os.path.splitext(name)[0]+"_server.pdf")
plot(name, client_cols, os.path.splitext(name)[0]+"_client.pdf")


