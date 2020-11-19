# norns.online

![111](https://user-images.githubusercontent.com/6550035/99736745-c470c180-2a7b-11eb-80d4-e9b2a02167cf.png)

crowdsource the control of a [norns](https://monome.org/docs/norns/) from [norns.online](https://norns.online).

## how does it work?

- norns runs a server that generates a screenshot at 10fps. screenshots are sent to special address on relay server ([duct.schollz.com](https://duct.schollz.com)) which servers as a multi-process, multi-consumer queue.
- server on norns listens to another address on relay server for commands and sanitizes those commands as either "encoders" or "keys". those commands are then sent to matron via websockets.
- website uses special address (designated by "`#`") to connect to relay and wait for / send info to your norns

## instructions

first SSH into the norns.

### improve the DNS resolution

edit the DHCP

```
> sudo vim /etc/dhcpcd.conf
```

add this line somewhere:

```
static domain_name_servers=192.168.0.1 1.1.1.1 1.0.0.1
```

restart it:

```
> sudo service dhcpcd restart
```


### allow arbitrary lua execution

[add this change](https://github.com/schollz/norns/commit/3202c3f1cfd40ac132d59e66276bfe0653ca2264) to allow arbitrary lua execution

then rebuild `matron` inside the norns:

```
> cd ~/norns
> ./waf configure
> ./waf
> sudo reboot now
``` 

### clone and build the program

```
> cd ~/dust
> git clone https://github.com/schollz/norns.online
> cd norns.online
> go build -v
> ./norns.online --name yourname
```

make sure you choose `yourname` to whatever you want. anyone with knowledge of `yourname` can access your norns.

### Go!

open https://norns.online/#yourname. share it with others and let them control your norns.

### optional: make stream

i haven't figured out how to embed the music stream. until I do, make a twitch stream and it can link from click on the screen.

## License

MIT