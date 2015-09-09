# python showspec.py <input file>
# Runs that through ConstantQ to produce a CQSpectrogram,
# writes that to out.raw, and then draws the result using matplotlib.

import matplotlib.pylab
import numpy
import sys
import subprocess

bins = 24
# TODO: Pass bins to go run
subprocess.call(["go", "run", "cqspectrogram.go"] + sys.argv[1:2])
ys1 = numpy.memmap("out.raw", dtype=numpy.complex64, mode="r").reshape((-1, bins*8)).T
ys1 = numpy.nan_to_num(ys1.copy())
# ys1[numpy.abs(ys1) < 1e-6] = 0

def plot_complex_spectrogram(ys, ax0, ax1):
    ax0.imshow(numpy.log(numpy.abs(ys)+1e-8), vmin=-12, vmax=5, cmap='gray')
    ax1.imshow(numpy.angle(ys), cmap='gist_rainbow')

# ys1 = ys1[:,::32]
if len(sys.argv) < 3:
    fig, (ax0, ax1) = matplotlib.pylab.subplots(nrows=2, sharex=True)
    plot_complex_spectrogram(ys1, ax0, ax1)
# else:
    # TODO: Support drawing two at a time
    # subprocess.call(["./makeSpectrogram", "-b %d" % bins] + sys.argv[2:3])
    # ys2 = numpy.memmap("out.raw", dtype=numpy.complex64, mode="r").reshape((-1, bins*8)).T
    # ys2 = ys2[:, ::32]

    # fig, (ax0, ax1, ax2, ax3) = matplotlib.pylab.subplots(nrows=4, sharex=True)
    # plot_complex_spectrogram(ys1, ax0, ax2)
    # plot_complex_spectrogram(ys2, ax1, ax3)
matplotlib.pylab.show()
