# norns.online

![111](https://raw.githubusercontent.com/schollz/norns.online/main/static/img/online2.PNG)

https://vimeo.com/484176216

**connect.** *norns.online* is two scripts and a website.

**listen and visualize.** run the *norns.online* script to beam your audio+visual data to `norns.online/<yourname>`.

**guide.** the website `norns.online/<yourname>` has the same inputs as your norns to provide remote guidance and access.

**collaborate.** use the *norns.online* script to connect to other norns via "*rooms*." anyone in a "room" shares live audio. the latency is ~4 seconds, so you can imagine that you all are playing together while being socially distanced by a distance of a quarter-mile.

**share.** use the *norns.online/share* script to upload or download tapes and also download script dumps. scripts that support will have `SHARE` available to load script tapes. 

_note:_ the script requires `ffmpeg` and `mpv`, which are automatically installed if you use this program (~300 MB).

**future directions:**

- fix all the 🐛🐛🐛
- audio sharing sync (very hard)
- audio sharing as input to softcut/engine

### Requirements

- norns 
- internet connection

### Documentation 


### share tapes and download script saves

![parameters for online](https://raw.githubusercontent.com/schollz/norns.online/main/static/img/upload.png)

- open the *norns.online/share* script
- register if you haven't already.
- download or upload tapes, and download script saves
- _note:_ *uploading* a script save must be done from a script's `SHARE` parameter

### beam your norns

![parameters for online](https://raw.githubusercontent.com/schollz/norns.online/main/static/img/online.png)

- open the *norns.online* script
- press K3. open browser to `norns.online/<yourname>`. if this is the first time running, wait for the `mpv` and `ffmpeg` programs to be installed (~300 MB).
- use norns normally, your norns will stay online in the background.
- press K2 to change name, or K1+K2 to update

### norns↔norns audio sharing

![parameters for sharing](https://raw.githubusercontent.com/schollz/norns.online/main/static/img/room_sharing.png)

- open the *norns.online* script
- go to gloal parameters and make sure both "`send audio`" and "`allow rooms`" are set to "`enabled`".
- change the "`room`" to the room you want to share audio. make sure your norns partner uses the same room.
- go to main screen and press K3 to go online. you should now be sharing audio with any other norns in that room.
- adjust "`room vol`" to change the level of incoming audio.

### uses

- make an internet radio from your norns
- collaborate between two norns
- twitch plays norns
- make demos (screen capture `norns.online`)
- download screenshots (right-click image at `norns.online` to download)
- tech support other people's norns
- !?!?!?

### faq

<details><summary><strong>how does the norns.online streaming work?</strong></summary>
norns runs a service that sends screenshot updates to <code>norns.online/&lt;yourname&gt;</code>. the website at <code>norns.online/&lt;yourname&gt;</code> sends inputs back to norns. norns listens to to inputs and runs the acceptable ones (adjustable with parameters). if enabled, norns will also stream packets of audio and send those to the website. the website will buffer them and play them so anyone with your address can hear your norns.
</details>


<details><summary><strong>how does audio streaming work?</strong></summary>
a pre-compiled <a href="https://github.com/kmatheussen/jack_capture"><code>jack_capture</code></a> periodically captures the norns output into 2-second flac files into a <code>/dev/shm</code> temp directory. each new flac packet is immediately sent out via websockets and then deleted. because of buffering, expect a lag of at least 4 seconds. when in a room, audio from other norns is piped into your norns via <code>mpv</code>. the incoming audio from other norns is added at the very end of the signal chain so (currently) it cannot be used as input to norns engines.
</details>

<details><summary><strong>is norns.online secure?</strong></summary>
<p>
for <em>norns.online</em>,if you are online, you have <a href="https://en.wikipedia.org/wiki/Security_through_obscurity">security through obscurity</a> (weak security). that means that <em>anyone</em> with the url <code>norns.online/&lt;yourname&gt;</code> can access your norns so you can make <code>&lt;yourname&gt;</code> complicated to be more secure. code injection is not possible, as i took precautions to make sure the inputs are sanitized on the norns so that only <code>enc()</code> and <code>key()</code> and <code>_menu.setmode()</code> functions are available. but, even with these functions someone could reset your norns / make some havoc. if this concerns you, don&#39;t share <code>&lt;yourname&gt;</code> with anyone or avoid using this script entirely.
</p>
<p>
for sharing on <em>norns.online/share</em>, everything is public but everything is also <strong>authenticated</strong>. authentication means that the data you download from someone named "bob" is truly data from the user who registered as "bob" and not someone posing as "bob". the server does not ensure that "bob" is a good or bad person, but only that the "bob" the server knows is the "bob" that registered with the server. authentication is provided through using rsa key-pairs. the server verifies your data comes from who you say you are by checking the signature on the hash of anything you upload. in theory, other people can obtain your key-pair directly from you to independently verify your data is actually coming from you (so the server need not be trusted), but this is not implemented yet.
</p>
</details>


<details><summary><strong>how much bandwidth does this use?</strong></summary>
if audio is enabled, a fair amount. the norns sends out screenshots periodically, but at the highest fps this is only ~18 kB/s.  however, if audio is enabled - the norns sends flac packets periodically (~170 kB/s = ~616 MB/hr). if you are audio-sharing a room you will be receiving about that much for each norns in the room. i tried reducing bandwidth by using lossy audio (ogg) however the gapless audio playback only worked without pops when using flac or wav.
</details>

<details><summary><strong>how much cpu does this use?</strong></summary>
not too much. on a raspberry pi 3b+ this uses about ~4% total CPU for capturing and sending audio data. screenshots also take cpu and higher fps takes more. the exact fps depends on the max fps (set in params) and how fast the screen changes (only updated screens are sent). at max it might take up to 30% of the cpu (15 fps!), but usually its 1-15%.
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