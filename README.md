# norns.online

![111](https://user-images.githubusercontent.com/6550035/99736745-c470c180-2a7b-11eb-80d4-e9b2a02167cf.png)

online [norns](https://monome.org/docs/norns/) on [norns.online](https://norns.online).

**control your norns** and listen to it from the internet. 

**share audio** with other norns around the world.


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

### quick start

![parameters for online](https://raw.githubusercontent.com/schollz/norns.online/main/static/img/online.png)

- press K3. open browser to `norns.online/<yourname>`
- use norns normally!

### norns↔norns audio sharing

![parameters for sharing](https://raw.githubusercontent.com/schollz/norns.online/main/static/img/audio_sharing.png)

- go to gloal parameters and make sure both "`send audio`" and "`allow rooms`" are set to "`enabled`".
- change the "`room`" to the room you want to share audio. make sure your norns partner uses the same room.
- go to main screen and press K3 to go online. you should now be sharing audio with any other norns in that room (max 3 per room).

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

<details><summary><strong>how does webpage work?</strong></summary>
norns runs a service that sends screenshots and audio to <code>norns.online/&lt;yourname&gt;</code>. the website at <code>norns.online/&lt;yourname&gt;</code> sends inputs back to norns. norns listens to to inputs and runs the acceptable ones (adjustable with parameters). what was <a href="https://llllllll.co/t/norns-online-crowdsource-your-norns/38492">just an idea</a> is now a norns script.
</details>


<details><summary><strong>how does audio streaming work?</strong></summary>
a pre-compiled <a href="https://github.com/kmatheussen/jack_capture"><code>jack_capture</code></a> periodically captures the norns output into 4-second files into the <code>/dev/shm</code> temp directory. these are converted to ogg-format are read and sent via websockets to the browser. the norns then deletes old files so excess memory is not used. expect a lag of at least 4 seconds. when in a room, audio from other norns is piped into your norns via <code>mpv</code>. the combined audio should only be accessible from the output of your norns (not on the browser).
</details>

<details><summary><strong>is this secure?</strong></summary>
if you are online, you have <a href="https://en.wikipedia.org/wiki/Security_through_obscurity">security through obscurity</a>. that means that <em>anyone</em> with the url <code>norns.online/&lt;yourname&gt;</code> can access your norns so you can make <code>&lt;yourname&gt;</code> complicated to be more secure. code injection is not possible, as i took precations to make sure the inputs are sanitized on the norns so that only <code>enc()</code> and <code>key()</code> and <code>_menu.setmode()</code> functions are available. but, even with these functions someone could reset your norns / make some havoc. if this concerns you, don&#39;t share <code>&lt;yourname&gt;</code> with anyone or avoid using this script entirely.
</details>


<details><summary><strong>how much bandwidth does this use?</strong></summary>
not too much. the norns sends out screenshots (~1.2 kB each) and - if audio is enabled - the norns sends ogg packets (~17.5 kB / second) periodically. if you use a fps of 4 + audio enabled, then you are sending out ~22.3 kB / second, which is ~80 MB/hour. if you are audio-sharing a room you will be receiving about that much for each norns in the room.
</details>

<details><summary><strong>ogg vs mp3 vs flac?</strong></summary>
audio sharing uses ogg. through flac is lossless (and therefore the best theoretical quality), ogg sounds really good (to me) for 10x less bandwidth. i tried mp3, but for some reason the mp3s will consistenly cause popping when the buffer switches over to the next packet, even at 320 kbps - this did not occur for ogg.
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