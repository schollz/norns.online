# norns.online

![111](https://user-images.githubusercontent.com/6550035/99736745-c470c180-2a7b-11eb-80d4-e9b2a02167cf.png)

Crowdsource the control of a [norns](https://monome.org/docs/norns/) from [norns.online](https://norns.online).

## Instructions

First SSH into the norns.

### Improve the DNS resolution

Edit the DHCPCD

```
> sudo vim /etc/dhcpcd.conf
```

Add this line somewhere:

```
static domain_name_servers=192.168.0.1 1.1.1.1 1.0.0.1
```

Restart it:

```
> sudo service dhcpcd restart
```


### Allowing arbitrary lua execution

[Add this change](https://github.com/schollz/norns/commit/3202c3f1cfd40ac132d59e66276bfe0653ca2264) to allow arbitrary lua execution

Then rebuild `matron` inside the norns.

```
> cd ~/norns
> ./waf configure
> ./waf
> sudo reboot now
``` 

### Clone and build the program

```
> git clone https://github.com/schollz/norns.online
> go build -v
> ./norns.online --name yourname
```

Make sure you choose `yourname` to whatever you want. Anyone with knowledge of `yourname` can access your norns.


### Go!

Open https://norns.online/#yourname

### Optional: make stream

I haven't figured out how to embed the music stream. Until I do, make a twitch stream and it can link from click on the screen.

## License

MIT