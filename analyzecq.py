import numpy as np
import matplotlib.pyplot as plt
import cmath, colorsys, math, sys, os

columns = []
meta = []
bpo = 24
octaves = 7
MASK = 1 << (octaves - 1)
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
    global columns
    values = np.memmap(inputFile, dtype=np.complex64, mode="r")

    at = 0
    columnCounter = MASK
    columnsRead = 0
    while at < len(values):
        samplesInColumn = trailingZeros(columnCounter) * bpo
        columns.append(values[at:at+samplesInColumn])
        columnCounter = (columnCounter % MASK) + 1
        at = at + samplesInColumn

        columnsRead = columnsRead + 1
        if columnsRead % 10000 == 0:
            print "%d columns read" % columnsRead

    # Normalize: find last full size column
    at = -1
    while len(columns[at]) != octaves * bpo:
        at -= 1
    columns = columns[:at]
    print "%d columns read" % (len(columns))

def readMeta(inputFile='out.meta'):
    global meta
    values = np.memmap(inputFile, dtype=np.int8, mode="r")

    at = 0
    columnCounter = MASK
    columnsRead = 0
    while at < len(values):
        samplesInColumn = trailingZeros(columnCounter) * bpo
        meta.append(values[at:at+samplesInColumn])
        columnCounter = (columnCounter % MASK) + 1
        at = at + samplesInColumn

        columnsRead = columnsRead + 1
        if columnsRead % 10000 == 0:
            print "%d columns read" % columnsRead

    # Normalize: find last full size column
    at = -1
    while len(meta[at]) != octaves * bpo:
        at -= 1
    meta = meta[:at]
    print "%d columns read" % (len(meta))

def readFileAndMeta():
    readFile()
    readMeta()
    if len(columns) != len(meta):
        print "CQ data and meta length do not match! %d vs %d" % (len(columns), len(meta))
        raise BaseException("oops")

def phaseScatter(octave, bin):
    p1, p2 = pairwise(octave, bin, octave + 1, bin, cmath.phase)
    p1, p2 = flatPhase(p1), flatPhase(p2)
    p1, p2 = np.diff(p1), np.diff(p2) * 2
    plt.scatter(p1, p2)
    plt.show()

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

def plotComplexSpectrogram(data, meta):
    data = data[:,::32]
    meta = meta[:,::32]
    fig, (ax0, ax1) = plt.subplots(nrows=2, sharex=True)
    power = np.abs(data)
    print "%f -> %f" % (np.min(meta), np.max(meta))
    # power = 20 * np.log10(power + 1e-8) # convert to DB
    phase = np.angle(data)
    ax0.imshow(power, cmap='gray')
    ax0.imshow(meta, cmap='copper', alpha=0.4)
    ax1.imshow(phase, cmap='gist_rainbow')
    plt.show()

def uninterpolatedSpectrogram():
    maxHeight = octaves * bpo
    asMatrix = np.zeros((maxHeight, len(columns)), dtype=np.complex64)
    metaMatrix = np.zeros((maxHeight, len(columns)), dtype=np.int8)
    for (i, column) in enumerate(columns):
        asMatrix[:len(column), i] = column
        metaMatrix[:len(column), i] = meta[i]
        if i > 0:
            asMatrix[len(column):, i] = asMatrix[len(column):, i - 1]
            metaMatrix[len(column):, i] = metaMatrix[len(column):, i - 1]
    plotComplexSpectrogram(asMatrix, metaMatrix)

def phaseVsPowerScatter():
    # TODO: Should these phase peaks line up more?
    for octave in range(6):
        for i in range(24):
            r, g, b = colorsys.hsv_to_rgb(i / 24.0, 1.0, 1.0)
            c = '#%02x%02x%02x' % (int(r*255), int(g*255), int(b*255))
            power = valuesForBin(octave, i, np.abs)
            phase = valuesForBin(octave, i, cmath.phase)

            # power = 20 * np.log10(np.abs(power) + 1e-8)
            ax1 = plt.subplot(3, 2, octave + 1)
            plt.title("octave %d" % octave)
            phase = flatPhase(phase)
            phase = np.diff(phase)
            phase = np.fmod(phase * (2 ** octave), 2 * np.pi)
            ax1.scatter(phase, power[1:], color=c, label='bin' + str(i))
    plt.show()

def notePower():
    bins = 24
    power = np.zeros(bins)
    for octave in range(3):
        for bin in range(bins):
            power[bin] += np.sum(valuesForBin(octave + 2, bin, np.abs))

    notes = ['A', 'A#', 'B', 'C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#']
    noteLen = len(notes)
    notePowers = np.zeros(noteLen)
    width = 0.7
    for note in range(noteLen):
        notePowers[note] = power[2 * note] + (power[2 * note + 1] + power[2 * note - 1]) / 2.0
    plt.bar(range(noteLen), notePowers, width)
    plt.xticks(np.arange(noteLen) + width / 2, notes)
    plt.show()


readFileAndMeta()
# uninterpolatedSpectrogram()
# phaseGraphsForBin(3)
# phaseScatter(2, 4)
# phaseVsPowerScatter()
notePower()
