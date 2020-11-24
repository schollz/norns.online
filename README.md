# norns.online

![111](https://user-images.githubusercontent.com/6550035/99736745-c470c180-2a7b-11eb-80d4-e9b2a02167cf.png)

online [norns](https://monome.org/docs/norns/) on [norns.online](https://norns.online).

**control your norns** and listen to it from the internet. 

**share audio*** with other norns around the world.

## quick start tutorial

![parameters for online](https://raw.githubusercontent.com/schollz/norns.online/main/static/img/online.png)

- install `norns.online` and then run it. 
- at the main screen press K3 to start/stop being online.
- the first time you run it will install `ffmpeg` and `mpv` and the `norns.online` server.
- when online, you can access your norns by opening a webpage to `norns.online/<yourname>`. you should hear audio and be able to use the inputs.
- you can change `<yourname>` by pressing K2.
- disable audio / change frame rate / adjust controls in global params

## sharing audio tutorial

![parameters for sharing](https://raw.githubusercontent.com/schollz/norns.online/main/static/img/audio_sharing.png)

- go to gloal parameters and make sure both "`send audio`" and "`allow rooms`" are set to "`enabled`".
- change the "`room`" to the room you want to share audio in
- go to main screen and press K3 to go online. you should now be sharing audio with any other norns in that room (max 3 per room).

## inner workings

**how does it work?** norns runs a service that sends screenshots and audio to `norns.online/<yourname>`. the website at `norns.online/<yourname>` sends inputs back to norns. norns listens to to inputs and runs the acceptable ones (adjustable with parameters). what was [just an idea](https://llllllll.co/t/norns-online-crowdsource-your-norns/38492) is now a norns script.

**how does audio streaming work?** a pre-compiled [`jack_capture`](https://github.com/kmatheussen/jack_capture) periodically captures the norns output into 4-second mp3 files into the `/dev/shm` temp directory. these mp3s are read and sent via websockets to the browser. the norns then deletes old files so excess memory is not used. expect a lag of at least 4 seconds. when in a room, audio from other norns is piped into your norns via `mpv`. the combined audio should only be accessible from the output of your norns (not on the browser).

**note of caution:** if you are using this and your norns is "online", then *anyone* with the url `norns.online/<yourname>` can access your norns. even though the inputs are sanitized on the norns so that only `enc()` and `key()` and `_menu.setmode()` functions are available, even with these functions someone could reset your norns / make some havoc. if this concerns you, don't share `<yourname>` with anyone or avoid using this script.


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

possible uses:

- play with other norns 
- norns as an internet radio
- twitch plays norns (params -> twitch to enable livestream)
- control multiple norns simultaneously
- make demos
- download screenshots
- tech support other people's norns
- !?!?!?


## my other norns SCRIPTS

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