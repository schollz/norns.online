-- norns.online v0.1.0
-- remote control for your norns
--
-- llllllll.co/t/norns.online
-- note!
-- this script opens your norns to
-- to the net - use with caution.
--    ▼ instructions below ▼
--
--

local json=include("lib/json")
local textentry=require 'textentry'

-- default files / directories
CODE_DIR="/home/we/dust/code/norns.online/"
CONFIG_FILE=CODE_DIR.."config.json"
KILL_FILE="/tmp/norns.online.kill"
START_FILE=CODE_DIR.."start.sh"
SERVER_FILE=CODE_DIR.."norns.online"
LATEST_RELEASE="https://github.com/schollz/norns.online/releases/download/v0.0.1/norns.online"

-- default settings
settings={
  name=randomString(5),
  allowmenu=false,
  allowencs=true,
  allowkeys=true,
  allowtwitch=false,
  keepawake=false,
  framerate=5,
}
uimessage=""
ui=1
uishift=false

function init()
  load_settings()
  write_settings()
  redraw()
end

function key(n,z)
  if n==2 then
    uishift=z==1
  elseif n==3 and z==1 and uishift then
    if ui==1 then
      textentry.enter(ui.name)
    elseif ui==2 then
      if util.file_exists(KILL_FILE) then
        stop()
      else
        start()
      end
    elseif ui==3 then
      settings.allowmenu=not settings.allowmenu
    elseif ui==4 then
      settings.allowkeys=not settings.allowkeys
    elseif ui==5 then
      settings.allowencs=not settings.allowencs
    elseif ui==6 then
      settings.allowtwitch=not settings.allowtwitch
    elseif ui==7 then
      settings.keepawake=not settings.keepawake
    elseif ui==8 then
      settings.framerate=settings.framerate+1
      if settings.framerate>12 then
        settings.framerate=1
      end
    elseif ui==9 then
      update()
    end
    write_settings()
  end
  redraw()
end

function enc(n,z)
  ui=util.clamp(ui+sign(z),1,8)
end

function redraw()
  screen.move(1,1)
  if ui==1 then
    screen.level(15)
  else
    screen.level(4)
  end
  screen.text("norns.online/"..settings.name)

  uistuff={}
  local i=1
  uistuff[i]={
    position={1,1},
    name="norns.online/"..settings.name,
  }
  i=i+1
  uistuff[i]={
    position={9,1},
    name="start",
  }
  if util.file_exists(KILL_FILE) then
    uistuff[i].name="stop"
  end
  i=i+1
  uistuff[i]={
    position={17,1},
    name="menu: disabled",
  }
  if settings.allowmenu then
    uistuff[i].name="menu: enabled"
  end
  i=i+1
  uistuff[i]={
    position={25,1},
    name="keys: disabled",
  }
  if settings.allowmenu then
    uistuff[i].name="keys: enabled"
  end
  i=i+1
  uistuff[i]={
    position={33,1},
    name="encs: disabled",
  }
  if settings.allowencs then
    uistuff[i].name="encs: enabled"
  end
  i=i+1
  uistuff[i]={
    position={33,1},
    name="twitch: disabled",
  }
  if settings.allowtwitch then
    uistuff[i].name="twitch: enabled"
  end
  i=i+1
  uistuff[i]={
    position={40,1},
    name="awake: disabled",
  }
  if settings.keepawake then
    uistuff[i].name="awake: enabled"
  end
  i=i+1
  uistuff[i]={
    position={48,1},
    name="framerate: "..settings.framerate,
  }
  i=i+1
  uistuff[i]={
    position={48,1},
    name="update?",
  }
  for i=1,9 do
    if ui==i then
      screen.level(15)
    else
      screen.level(4)
    end
    screen.move(uistuff[i].position[1],uistuff[i].position[2])
    screen.text(uistuff[i].name)
  end

  if uimessage~="" then
    -- get the pixel length of the string
    local width=screen.text_extents(uimessage)

    -- draw our box
    local x=10
    local y=10
    local padding=10
    screen.level(15)
    screen.rect(x,y,width+padding,10)
    screen.fill()

    -- draw our text
    screen.level(0)
    screen.move(x+(padding/2),y+8)
    screen.text(uimessage)

  end
  screen.update()
end

--
-- norns.online stuff
--

function write_settings()
  jsondata=json.encode(settings)
  f=io.open(CONFIG_FILE,"w")
  f:write(jsondata)
  f:close(f)
end

function load_settings()
  if not util.file_exists(CONFIG_FILE) then
    do return end
  end
  data=readAll(CONFIG_FILE)
  settings=json.decode(data)
end

function update()
  uimessage="building"
  redraw()
  os.execute("cd "..CODE_DIR.."; go build")
  uimessage=""
  redraw()
  if not util.file_exists(SERVER_FILE) then
    uimessage="downloading"
    redraw()
    os.execute("curl "..LATEST_RELEASE.." -o "..SERVER_FILE)
    uimessage=""
    redraw()
  end
end

function start()
  write_settings()
  if not util.file_exists(SERVER_FILE) then
    update()
  end
  make_start_sh()
  os.execute(START_FILE)
  redraw()
end

function stop()
  os.execute(KILL_FILE)
  redraw()
end

function make_start_sh()
  startsh="#!/bin/bash\n"
  startsh=startsh..CODE_DIR.."norns.online --config "..CODE_DIR.."config.json > /dev/null &\n"
  f=io.open(START_FILE,"w")
  f:write(startsh)
  f:close(f)
  os.execute("chmod +x "..START_FILE)
end

--
-- utils
--

function sign(x)
  if x>0 then
    return 1
  elseif x<0 then
    return-1
  else
    return 0
  end
end


function show_message(message)
  uimessage=message
  redraw()
  clock.run(function()
    clock.sleep(0.5)
    uimessage=""
    redraw()
  end)
end

function readAll(file)
  local f=assert(io.open(file,"rb"))
  local content=f:read("*all")
  f:close()
  return content
end

local charset={} do -- [a-z]
  for c=97,122 do table.insert(charset,string.char(c)) end
end

local function randomString(length)
  if not length or length<=0 then return '' end
  math.randomseed(os.clock()^5)
  return randomString(length-1)..charset[math.random(1,#charset)]
end

 
