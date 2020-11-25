# norns.online

![111](https://user-images.githubusercontent.com/6550035/99736745-c470c180-2a7b-11eb-80d4-e9b2a02167cf.png)

online [norns](https://monome.org/docs/norns/) on [norns.online](https://norns.online).

**control your norns** and listen to it from the internet. just open a browser to `norns.online/<yourname>` and you'll see and hear your norns!

**share audio** with other norns around the world. currently default the time-lag between browsers/norns is ~4 seconds (so its as if you are socially distanced by 1/4 mile).

what was <a href="https://llllllll.co/t/norns-online-crowdsource-your-norns/38492">just an idea</a> is now a reality.

future directions:

- fix all the 🐛🐛🐛

### Requirements

- norns 
- internet connection

### Documentation 

- K3 toggles internet
- K2 changes name
- K1+K2 updates
- more params in global menu

_note:_ this app requires `ffmpeg` and `mpv`, which are automatically installed if you use this program.

### quick start

![parameters for online](https://raw.githubusercontent.com/schollz/norns.online/main/static/img/online.png)

- press K3. open browser to `norns.online/<yourname>`
- use norns normally, your norns will stay online in the background.

### norns↔norns audio sharing

![parameters for sharing](https://raw.githubusercontent.com/schollz/norns.online/main/static/img/room_sharing.png)

- go to gloal parameters and make sure both "`send audio`" and "`allow rooms`" are set to "`enabled`".
- change the "`room`" to the room you want to share audio. make sure your norns partner uses the same room.
- go to main screen and press K3 to go online. you should now be sharing audio with any other norns in that room.
- adjust "`room vol`" to change the level of incoming audio.

### other uses 

- play with other norns 
- norns as an internet radio
- twitch plays norns (params -> twitch to enable livestream)
- control multiple norns simultaneously
- make demos
- download screenshots
- tech support other people's norns
- !?!?!?

### faq

<details><summary><strong>how does the norns.online webpage work?</strong></summary>
norns runs a service that sends screenshots to <code>norns.online/&lt;yourname&gt;</code>. the website at <code>norns.online/&lt;yourname&gt;</code> sends inputs back to norns. norns listens to to inputs and runs the acceptable ones (adjustable with parameters). if enabled, norns will also stream packets of audio and send those to the website. the website will buffer them and play them so anyone with your address can hear your norns.
</details>


<details><summary><strong>how does audio streaming work?</strong></summary>
a pre-compiled <a href="https://github.com/kmatheussen/jack_capture"><code>jack_capture</code></a> periodically captures the norns output into 2-second flac files into a <code>/dev/shm</code> temp directory. each new flac packet is immediately sent out via websockets and then deleted. because of buffering, expect a lag of at least 4 seconds. when in a room, audio from other norns is piped into your norns via <code>mpv</code>. the incoming audio from other norns is added at the very end of the signal chain so (currently) it cannot be used as input to norns engines.
</details>

<details><summary><strong>is this secure?</strong></summary>
if you are online, you have <a href="https://en.wikipedia.org/wiki/Security_through_obscurity">security through obscurity</a>. that means that <em>anyone</em> with the url <code>norns.online/&lt;yourname&gt;</code> can access your norns so you can make <code>&lt;yourname&gt;</code> complicated to be more secure. code injection is not possible, as i took precautions to make sure the inputs are sanitized on the norns so that only <code>enc()</code> and <code>key()</code> and <code>_menu.setmode()</code> functions are available. but, even with these functions someone could reset your norns / make some havoc. if this concerns you, don&#39;t share <code>&lt;yourname&gt;</code> with anyone or avoid using this script entirely.
</details>


<details><summary><strong>how much bandwidth does this use?</strong></summary>
if audio is enabled, a fair amount. the norns sends out screenshots periodically, but at the highest fps this is only ~18 kB/s.  however, if audio is enabled - the norns sends flac packets periodically (~170 kB/s = ~616 MB/hr). if you are audio-sharing a room you will be receiving about that much for each norns in the room. i tried reducing bandwidth by using lossy audio (ogg) however the gapless audio playback only worked without pops when using flac or wav.
</details>

<details><summary><strong>how much cpu does this use?</strong></summary>
this uses about ~4% of the CPU for capturing and sending audio data. the main usage comes from screenshots, which cost about 2-3% cpu every fps. that means if you run at max of 12 fps you will be using at least %30-40 of cpu.
</details>


## my other norns scripts

- [barcode](https://github.com/schollz/barcode): replays a buffer six times, at different levels & pans & rates & positions, modulated by lfos on every parameter.
- [blndr](https://github.com/schollz/blndr): a quantized delay with time morphing
- [clcks](https://github.com/schollz/clcks): a tempo-locked repeater
- [oooooo](https://github.com/schollz/oooooo): digital tape loops
- [piwip](https://github.com/schollz/piwip): play instruments while instruments play.
- [glitchlets](https://github.com/schollz/glitchlets): 
add glitching to everything.
- [abacus](https://github.com/schollz/abacus): 
sampler sequencer.

## license

mit