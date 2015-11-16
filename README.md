go-sound
======

go-sound is a go library for dealing with sound waves.
At the fundamental level, the library models each sound as a channel of [-1, 1] samples that represent the 
[Compression Wave](https://en.wikipedia.org/wiki/Sound#Longitudinal_and_transverse_waves) that comprises the sound.  
To see it it action, check out [demo.go](https://github.com/padster/go-sound/blob/master/demo.go) or the examples
provided by each file in the [sounds](https://github.com/padster/go-sound/tree/master/sounds) package.

A tutorial explaining the basics behind sound wave modelling in code, and how it is implemented in go-sound, is available on my blog: http://padsterprogramming.blogspot.ch/2015/11/sounds-good-part-1.html  

### Features :
 - A Sound interface, with a BaseSound implementation that makes it simpler to write your own.
 - Sound Math (play notes together to make chords, or in serial to form a melody, ...)
 - Utilities for dealing with sounds (repeat sounds, generate from text, ...)
 - Implementations for various inputs (silence, sinusoidal wave, .wav file, ...)
 - Implementations for various outputs (play via pulse audio, draw to screen, .wav file, ...)
 - Realtime input (via MIDI) - with delay though.
 - Sound -> Spectrogram -> Sound conversion using a [Constant Q transform](https://en.wikipedia.org/wiki/Constant_Q_transform)

### In progress:
 - MashApp, a golang server and polymer web app for manipulating sounds using the library.

### Future plans:
 - Inputs and Outputs integrating with [Jack Audio](http://jackaudio.org)
 - Realtime input from microphone, more efficient from MIDI
 - Effects algorithms (digitial processing like reverb, bandpass ...)

#### Notes: 
This library requires pulse audio installed to play the sounds, libflac for reading/writing flac files, and OpenGL 3.3 / GLFW 3.1 for rendering a soundwave to screen.

Some planned additions are included above, and include effects like those available in [Audacity](http://audacityteam.org/)
(e.g. rewriting Nyquist, LADSPA plugins in Go), or ones explained [here](https://www.youtube.com/channel/UCchjpg1aaY91WubqAYRcNsg)
or [here](https://christianfloisand.wordpress.com/2012/09/04/digital-reverberation). 
Additionally, some more complex instrument synthesizers could be added, and contributions are welcome.

The example piano .wav C note came from: http://freewavesamples.com/ensoniq-sq-1-dyno-keys-c4

Frequencies of notes are all obtained from: http://www.phy.mtu.edu/~suits/notefreqs.html

For MIDI input, a number of things are required for portmidi:
 - Instructions to test the midi input device: https://wiki.archlinux.org/index.php/USB_MIDI_keyboards
 - Instructions to set linux up for realtime processing: http://tedfelix.com/linux/linux-midi.html
 - ALSA dev library required (libasound2-dev)
 - I needed to manually install portmidi: http://sourceforge.net/p/portmedia/wiki/Installing_portmidi_on_Linux/
   - This also required removing the "WORKING_DIRECTORY pm_java)" in the ccmake configs
   - And to link against /usr/local/lib/libportmidi.so instead of /usr/lib/libporttime.so

Overall quite a pain and there's still a noticeable delay in the MIDI input, patches to reduce that are welcome!

Credit to [cryptix](//github.com/cryptix), [cocoonlife](//github.com/cocoonlife),  [moriyoshi](//github.com/moriyoshi) and [rakyll](//github.com/rakyll) for their wavFile, pulseAudio and portmidi implementations respectively, used by go-sound.

