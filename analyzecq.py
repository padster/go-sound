import numpy as np
from scipy.fftpack import fft
from scipy.signal import get_window
import cmath
import math

import sys, os
import matplotlib.pyplot as plt


columns = []
bpo = 24
octaves = 7
MASK = 1 << octaves
TAU = 2 * np.pi

def fixInput(m):
    print "%f to %f" % (np.max(m), np.min(m))
    m = np.transpose(m)
    return m

def display(mX, pX):
    mX = fixInput(mX)
    pX = fixInput(pX)
    fig, (ax0, ax1) = plt.subplots(nrows=2, sharex=True)
    ax0.imshow(mX, cmap='gray')
    ax1.imshow(pX, cmap='gist_rainbow')
    plt.show()

def trailingZeros(num):
    if num % 2 == 0:
        return 1 + trailingZeros(num / 2)
    return 1

def identity(x):
    return x

def flatPhase(values):
    i = 1
    while i < len(values):
        delta = values[i] - values[i-1]
        # Move to the closest multiple of TAU
        delta = round(delta / TAU) * TAU
        values[i] -= delta
        i = i + 1
    return values

def valuesForBin(octave, bin, f=identity):
    result = []
    skip = 1 << octave
    bindex = octave * bpo + bin
    for i in range(0, len(columns), skip):
        result.append(f(columns[i][bindex]))
    return np.array(result)

def pairwise(octave1, bin1, octave2, bin2, f=identity):
    r1, r2 = [], []
    skip = 1 << max(octave1, octave2)
    bindex1 = octave1 * bpo + bin1
    bindex2 = octave2 * bpo + bin2
    for i in range(0, len(columns), skip):
        r1.append(f(columns[i][bindex1]))
        r2.append(f(columns[i][bindex2]))
    return r1, r2

def running_mean(x, N):
   cumsum = np.cumsum(np.insert(x, 0, 0)) 
   return (cumsum[N:] - cumsum[:-N]) / N

def readFile(inputFile='out.cq'):
    columns = []
    values = np.memmap(inputFile, dtype=np.complex64, mode="r")

    at = 0
    columnCounter = MASK
    columnsWritten = 0
    while at < len(values):
        samplesInColumn = trailingZeros(columnCounter) * bpo
        columns.append(values[at:at+samplesInColumn])
        columnCounter = (columnCounter % MASK) + 1
        at = at + samplesInColumn

        columnsWritten = columnsWritten + 1
        if columnsWritten % 10000 == 0:
            print "%d columns written" % columnsWritten

    # Normalize: find last full size column
    at = -1
    while len(columns[at]) != octaves * bpo:
        at -= 1
    print "%d columns read" % (len(columns) + at)
    return columns[:at]


def phaseGraphsForBin(bin):
    for octave in range(octaves):
        phase = valuesForBin(octave, bin, cmath.phase)
        phase = flatPhase(phase)
        factor = 1 << octave
        x = range(0, len(phase) * factor, factor)
        plt.plot(x, phase * factor, label='octave ' + str(octave))
    plt.legend(
        loc='upper center', bbox_to_anchor=(0.5, 1.0),
        ncol=3, fancybox=True, shadow=True)
    plt.show()

columns = readFile()
phaseGraphsForBin(3)
