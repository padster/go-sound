# python showspec.py <input file>
# Runs that through ConstantQ to produce a CQSpectrogram,
# writes that to out.raw, and then draws the result using matplotlib.

import matplotlib.pylab
import numpy as np
import sys
import subprocess

bins = 72
octaves = 7
# TODO: Pass bins to go run
# subprocess.call(["go", "run", "cqspectrogram.go"] + sys.argv[1:2])
ys1 = np.memmap("out.raw", dtype=np.complex64, mode="r").reshape((-1, bins*octaves)).T
ys1 = np.nan_to_num(ys1.copy())

# ys1[numpy.abs(ys1) < 1e-6] = 0

def running_mean(x, N):
    cumsum = np.cumsum(np.insert(x, 0, np.zeros(N))) 
    return (cumsum[N:] - cumsum[:-N]) / N 

def plot_complex_spectrogram(ys, ax0, ax1):
    rows, cols = ys.shape
    values = np.abs(ys)

    # for i in range(0, rows):
        # values[i, :] = running_mean(values[i, :], 32)

    # values = np.log(np.abs(values) + 1e-8)
    # values = np.log(np.abs(ys) + 1e-8)
    # values = np.abs(ys)
    # values = np.abs(ys)
    # ax0.imshow(values, vmin=-12, vmax=5, cmap='gray')

    # colMax = np.mean(values, axis=1)[::-1]
    # colMax = np.insert(colMax, 0, 0)
    # ax1.plot(colMax)
    # for i in range(0, rows, 12):
        # ax1.axvline(i, color='r')
    # print "Max = %d" % np.argmax(colMax)
    ax0.imshow(values, cmap='gray')
    # ax1.plot(values[:, 256])
    # colSum = np.std(values, axis=0)
    # colSumD = np.diff(colSum)
    # ax1.plot(running_mean(colSum, 5))
    # notes = np.where((colSumD > 250) & (colSumD < 600))[0]
    # notes = np.insert(notes, 0, 0)
    # dupes = np.diff(notes)
    # notes = notes[np.where(dupes > 50)[0] + 1]
    # notes = notes[np.where(notes < 10500)[0]] + 18
    # notes = np.concatenate((
        # np.arange(50, 300, 52),
        # np.arange(328, 600, 60),
        # np.arange(670, 1380, 54)
    # ))
    # print notes
    # i = 0
    # for note in notes[0:5]:
        # ax0.plot(values[:, note])
        # i += 9
        # ax0.axvline(note, color='r')
        # ax1.axvline(note, color='r')
    # print "# notes = %d" % len(notes)
    # 50 changes
    #ax1.imshow(np.angle(ys), cmap='gist_rainbow')

ys1 = ys1[:,::16]
# ys1 = ys1[:,256:284]

if len(sys.argv) < 3:
    fig, (ax0, ax1) = matplotlib.pylab.subplots(nrows=2, sharex=True)
    plot_complex_spectrogram(ys1, ax0, ax1)
# else:
    # TODO: Support drawing two at a time
    # subprocess.call(["./makeSpectrogram", "-b %d" % bins] + sys.argv[2:3])
    # ys2 = np.memmap("out.raw", dtype=np.complex64, mode="r").reshape((-1, bins*8)).T
    # ys2 = ys2[:, ::32]

    # fig, (ax0, ax1, ax2, ax3) = matplotlib.pylab.subplots(nrows=4, sharex=True)
    # plot_complex_spectrogram(ys1, ax0, ax2)
    # plot_complex_spectrogram(ys2, ax1, ax3).
matplotlib.pylab.show()

