-- norns.online v0.1.0
-- online norns on norns.online
--
-- llllllll.co/t/norns-online
-- note!!! this script opens your
-- norns to to the net
-- - (ab)use with caution.
--    ▼ instructions below ▼
-- K3 toggles internet
-- K2 changes name
-- K1+K2 updates
-- more params in global menu
-- if you enable audio, make sure
-- to restart norns.online

local json=include("lib/json")
local textentry=require 'textentry'

-- default files / directories
CODE_DIR="/home/we/dust/code/norns.online/"
CONFIG_FILE=CODE_DIR.."config.json"
KILL_FILE="/tmp/norns.online.kill"
START_FILE=CODE_DIR.."start.sh"
SERVER_FILE=CODE_DIR.."norns.online"
LATEST_RELEASE="https://github.com/schollz/norns.online/releases/download/v0.1.0/norns.online"

-- default settings
settings={
  name="",
  allowmenu=true,
  allowencs=true,
  allowkeys=true,
  allowtwitch=false,
  sendaudio=false,
  keepawake=false,
  framerate=5,
}
uimessage=""
ui=1
uishift=false
params:add_separator("norns.online")
function init()
  params:add_option("sendaudio","send audio",{"disabled","enabled"},1)
  params:set_action("sendaudio",function(v)
    settings.sendaudio=v==2
    write_settings()
  end)
  params:add_option("allowmenu","menu",{"disabled","enabled"},2)
  params:set_action("allowmenu",function(v)
    settings.allowmenu=v==2
    write_settings()
  end)
  params:add_option("allowencs","encs",{"disabled","enabled"},2)
  params:set_action("allowencs",function(v)
    settings.allowencs=v==2
    write_settings()
  end)
  params:add_option("allowkeys","keys",{"disabled","enabled"},2)
  params:set_action("allowkeys",function(v)
    settings.allowkeys=v==2
    write_settings()
  end)
  params:add_option("allowtwitch","twitch",{"disabled","enabled"},1)
  params:set_action("allowtwitch",function(v)
    settings.allowtwitch=v==2
    write_settings()
  end)
  
  params:add_option("keepawake","keep awake",{"disabled","enabled"},1)
  params:set_action("keepawake",function(v)
    settings.keepawake=v==2
    write_settings()
  end)
  
  params:add_control("framerate","frame rate",controlspec.new(1,12,'lin',1,5,'fps'))
  params:set_action("framerate",function(v)
    settings.framerate=v
    write_settings()
  end)
  
  settings.name=randomString(5)
  load_settings()
  write_settings()
  redraw()
end

function key(n,z)
  if n==1 then
    uishift=z
  elseif uishift==1 and n==2 then
    update()
  elseif n==2 and z==0 then
    textentry.enter(function(x)
      if x~=nil then
        settings.name=x
      end
    end,settings.name,"norns.online/")
  elseif n==3 and z==1 then
    toggle()
  end
  redraw()
end

function enc(n,z)
  redraw()
end

function redraw()
  screen.clear()
  screen.level(4)
  screen.font_face(3)
  screen.font_size(12)
  screen.move(64,8)
  screen.text_center("you are now")
  screen.move(64,22)
  screen.font_face(3)
  screen.font_size(12)
  screen.level(15)
  print(util.file_exists(KILL_FILE))
  if util.file_exists(KILL_FILE) then
    screen.text_center("online")
    
    screen.level(4)
    screen.move(64,36)
    screen.font_face(3)
    screen.font_size(12)
    screen.text_center("norns.online/")
    
    screen.level(15)
    screen.move(64,58)
    screen.font_face(7)
    screen.font_size(24)
    if string.len(settings.name)>20 then
      screen.move(64,53)
      screen.font_size(12)
    elseif string.len(settings.name)>10 then
      screen.move(64,53)
      screen.font_size(14)
    end
    screen.text_center(settings.name)
  else
    screen.level(15)
    screen.text_center("offline")
  end
  
  screen.font_face(1)
  screen.font_size(8)
  if uimessage~="" then
    screen.level(15)
    x=64
    y=28
    w=string.len(uimessage)*6
    screen.rect(x-w/2,y,w,10)
    screen.fill()
    screen.level(15)
    screen.rect(x-w/2,y,w,10)
    screen.stroke()
    screen.move(x,y+7)
    screen.level(0)
    screen.text_center(uimessage)
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
  tab.print(settings)
  if settings.sendaudio then
    params:set("sendaudio",2)
  else
    params:set("sendaudio",1)
  end
  if settings.allowmenu then
    params:set("allowmenu",2)
  else
    params:set("allowmenu",1)
  end
  if settings.allowencs then
    params:set("allowencs",2)
  else
    params:set("allowencs",1)
  end
  if settings.allowkeys then
    params:set("allowkeys",2)
  else
    params:set("allowkeys",1)
  end
  if settings.allowtwitch then
    params:set("allowtwitch",2)
  else
    params:set("allowtwitch",1)
  end
  params:set("framerate",settings.framerate)
end

function update()
  uimessage="updating"
  redraw()
  os.execute("cd "..CODE_DIR.." && git pull")
  uimessage="building"
  redraw()
  os.execute("cd "..CODE_DIR.."; /usr/local/go/bin/go build")
  uimessage=""
  redraw()
  if not util.file_exists(SERVER_FILE) then
    uimessage="downloading"
    redraw()
    os.execute("curl -L "..LATEST_RELEASE.." -o "..SERVER_FILE)
    os.execute("chmod +x "..SERVER_FILE)
    uimessage=""
    redraw()
  end
  if util.file_exists(SERVER_FILE) then
    show_message("updated.")
  end
end

function toggle()
  if util.file_exists(KILL_FILE) then
    uimessage="stopping"
    redraw()
    clock.run(function()
      for i=1,10000 do
        if not util.file_exists(KILL_FILE) then
          uimessage=""
          redraw()
          break
        end
        clock.sleep(0.1)
      end
    end)
    stop()
  else
    uimessage="starting"
    redraw()
    clock.run(function()
      for i=1,10000 do
        if util.file_exists(KILL_FILE) then
          uimessage=""
          redraw()
          break
        end
        clock.sleep(0.1)
      end
    end)
    start()
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

function randomString(length)
  if not length or length<=0 then return '' end
  math.randomseed(os.clock()^5)
  return randomString(length-1)..charset[math.random(1,#charset)]
end
