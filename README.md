go-sound
======

go-sound is a go library for dealing with sound waves.
At the fundamental level, the library models each sound as a channel of [-1, 1] samples that represent the 
[Compression Wave](https://en.wikipedia.org/wiki/Sound#Longitudinal_and_transverse_waves) that comprises the sound.  
To see it it action, check out [demo.go](https://github.com/padster/go-sound/blob/master/demo.go) or the examples
provided by each file in the [sounds](https://github.com/padster/go-sound/tree/master/sounds) package.

### Available :
 - A Sound interface, with a BaseSound implementation that makes it simpler to write your own.
 - Sound Math (play notes together to make chords, or in serial to form a melody, ...)
 - Utilities for dealing with sounds (repeat sounds, generate from text, ...)
 - Implementations for various inputs (silence, sinusoidal wave, .wav file, ...)
 - Implementations for various outputs (play via pulse audio, draw to screen, .wav file, ...)

### Future plans:
 - Realtime input (e.g. midi controller or microphone)
 - Effects algorithms (digitial processing like delay, reverb, ...)
 - UI for modifying sounds and applying the math/effects.


#### Notes: 
This library requires pulse audio installed to play the sounds, and OpenGL 3.3 / GLFW 3.1 for rendering a soundwave to screen.

Some planned additions are included above, and include effects like those available in [Audacity](http://audacityteam.org/)
(e.g. rewriting Nyquist, LADSPA plugins in Go), or ones explained [here](https://www.youtube.com/channel/UCchjpg1aaY91WubqAYRcNsg)
or [here](https://christianfloisand.wordpress.com/2012/09/04/digital-reverberation). 
Additionally, some more complex instrument synthesizers could be added, and contributions are welcome.

The example piano .wav C note came from: http://freewavesamples.com/ensoniq-sq-1-dyno-keys-c4

Credit to [cryptix](//github.com/cryptix) and [moriyoshi](//github.com/moriyoshi) for their wavFile and pulseAudio implementations respectively, used by go-sound.

