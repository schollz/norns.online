# norns.online

![111](https://user-images.githubusercontent.com/6550035/99736745-c470c180-2a7b-11eb-80d4-e9b2a02167cf.png)

online [norns](https://monome.org/docs/norns/) on [norns.online](https://norns.online).

**control your norns** and listen to it from the internet. just open a browser to `norns.online/<yourname>` and you'll see your norns!

**share audio** with other norns around the world. play with other people as if you are 1/4 mile apart (where sound takes ~4 seconds to reach the other person).

what was <a href="https://llllllll.co/t/norns-online-crowdsource-your-norns/38492">just an idea</a> is now a reality

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

<details><summary><strong>how does the norns.online webpage work?</strong></summary>
norns runs a service that sends screenshots to <code>norns.online/&lt;yourname&gt;</code>. the website at <code>norns.online/&lt;yourname&gt;</code> sends inputs back to norns. norns listens to to inputs and runs the acceptable ones (adjustable with parameters). if enabled, norns will also stream packets of audio and send those to the website. the website will buffer them and play them so anyone with your address can hear your norns.
</details>


<details><summary><strong>how does audio streaming work?</strong></summary>
a pre-compiled <a href="https://github.com/kmatheussen/jack_capture"><code>jack_capture</code></a> periodically captures the norns output into 4-second files into the <code>/dev/shm</code> temp directory. these are converted to ogg-format are read and sent via websockets to the browser. the norns then deletes old files so excess memory is not used. expect a lag of at least 4 seconds. when in a room, audio from other norns is piped into your norns via <code>mpv</code>. the incoming audio from other norns is added at the very end of the signal chain so (currently) it cannot be used as input to norns engines.
</details>

<details><summary><strong>is this secure?</strong></summary>
if you are online, you have <a href="https://en.wikipedia.org/wiki/Security_through_obscurity">security through obscurity</a>. that means that <em>anyone</em> with the url <code>norns.online/&lt;yourname&gt;</code> can access your norns so you can make <code>&lt;yourname&gt;</code> complicated to be more secure. code injection is not possible, as i took precations to make sure the inputs are sanitized on the norns so that only <code>enc()</code> and <code>key()</code> and <code>_menu.setmode()</code> functions are available. but, even with these functions someone could reset your norns / make some havoc. if this concerns you, don&#39;t share <code>&lt;yourname&gt;</code> with anyone or avoid using this script entirely.
</details>


<details><summary><strong>how much bandwidth does this use?</strong></summary>
not too much. the norns sends out screenshots (~1.2 kB each) and - if audio is enabled - the norns sends flac packets (~170 kB / second) periodically. if you use a fps of 4 + audio enabled, then you are sending out ~171 kB / second, which is ~616 MB/hour. if you are audio-sharing a room you will be receiving about that much for each norns in the room.
</details>


<details><summary><strong>how do i prevent audible pops?</strong></summary>
the audible pops in playback on the browser or norns are from badly switched buffers. i've found that upgrading the norns greatly helps to reduce this:<pre>
> ssh we@norns.local
> sudo apt upgrade
</pre>
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