-- norns.online v1.1.1
-- share, connect, collaborate.
--
-- llllllll.co/t/norns-online
-- 
-- 
-- 
--    ▼ important notes ▼
-- this script will install about
-- 200 MB of libraries in order 
-- to function (mpv and ffmpeg).
-- do not continue to avoid
-- installing.

local textentry=require 'textentry'
local fileselect=require 'fileselect'
local share=include("norns.online/lib/share")
local json=include("lib/json")


-- default files / directories
CODE_DIR=_path.code.."norns.online/"
DATA_DIR=_path.data.."norns.online/"
CONFIG_FILE=DATA_DIR.."config.json"
KILL_FILE="/dev/shm/norns.online.kill.sh"
START_FILE=CODE_DIR.."start.sh"
SERVER_FILE=CODE_DIR.."norns.online"
LATEST_RELEASE="https://github.com/schollz/norns.online/releases/download/v1.1.0/norns.online"

-- default settings
settings={
  name="",
  room="llllllll",
  allowroom=false,
  allowmenu=false,
  allowencs=false,
  allowkeys=false,
  allowtwitch=false,
  sendaudio=true,
  keepawake=false,
  framerate=2,
  buffertime=4000,
  roomsize=1,
  packetsize=4,
  roomvolume=80,
  is_registered=false,
  is_installed=false,
}
mode=1
uimessage=""
uishift=false
refreshed_dir=false

function init()

  print(res)
  startup=true
  params:add_separator("norns.online")
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

  params:add_control("framerate","max frame rate",controlspec.new(0,15,'lin',1,4,'fps'))
  params:set_action("framerate",function(v)
    settings.framerate=v
    write_settings()
  end)

  params:add_separator("audio sharing")
  params:add_option("sendaudio","send audio",{"disabled","enabled"},2)
  params:set_action("sendaudio",function(v)
    if not startup then os.execute(KILL_FILE) end
    settings.sendaudio=v==2
    write_settings()
    redraw()
  end)

  params:add_control("packetsize","packet size",controlspec.new(1,30,'lin',1,2,'s'))
  params:set_action("packetsize",function(v)
    if not startup then os.execute(KILL_FILE) end
    settings.packetsize=v
    write_settings()
    redraw()
  end)

  params:add_separator("norns<->norns")
  params:add_option("allowroom","allow rooms",{"disabled","enabled"},1)
  params:set_action("allowroom",function(v)
    if not startup then os.execute(KILL_FILE) end
    settings.allowroom=v==2
    write_settings()
    redraw()
  end)

  params:add_text("roomname","room name","llllllll")
  params:set_action("roomname",function(v)
    if not startup then os.execute(KILL_FILE) end
    settings.room=v
    write_settings()
    redraw()
  end)

  params:add_control("roomsize","room size",controlspec.new(1,3,'lin',1,1,'other'))
  params:set_action("roomsize",function(v)
    if not startup then os.execute(KILL_FILE) end
    settings.roomsize=v
    write_settings()
    redraw()
  end)

  params:add_control("buffertime","room buffer time",controlspec.new(100,3000,'lin',100,2000,'ms'))
  params:set_action("buffertime",function(v)
    if not startup then os.execute(KILL_FILE) end
    settings.buffertime=v
    write_settings()
    redraw()
  end)

  params:add_control("roomvolume","room vol",controlspec.new(0,100,'lin',5,80,''))
  params:set_action("roomvolume",function(x)
    settings.roomvolume=x
    write_settings()
    redraw()
  end)

  load_settings()
  write_settings()
  redraw()
  startup=false
end

function key(k,z)
  if k==1 then 
    uishift = z==1
  end
  if z==0 then
    do return end
  end
  if not settings.is_installed then
    show_message("checking installation...")
    install_prereqs()
  end
  if uishift and k==2 then 
    norns_online_update()
  elseif k==3 and mode==5 and util.file_exists(KILL_FILE) then
    print("killing server")
      -- kill
      stop()
      redraw()
  elseif k==3 and settings.name=="" then
    print(settings.name)
    print("register mode")
    server_generate_key()
    redraw()
  elseif k==3 and mode==5 then
    start()
  elseif k==3 then
    print("go mode")
    print(settings.is_registered)
    if not settings.is_registered then
      settings.is_registered=share.is_registered(settings.name)
      if not settings.is_registered then
        server_register(settings.name)
      else
        write_settings()
      end
    end
    if mode==1 then
      -- upload
      fileselect.enter("/home/we/dust/audio",upload_callback)

    elseif mode==2 or mode ==3 then
      -- download
      local dirtogo="tape"
      if mode ==3 then 
        dirtogo=""
      end
      refresh_directory()
      fileselect.enter(share.get_virtual_directory(dirtogo),function(x)
        if x == "cancel" then 
          do return end 
        end 
        uimessage="downloading..."
        redraw()
        msg = share.download_from_virtual_directory(x)
        show_message(msg)
        redraw()
      end)
    elseif mode==4 then
      -- delete something
      refresh_directory()
      fileselect.enter(share.get_virtual_directory(),function(x)
        if x == "cancel" then 
          do return end 
        end 
        _,filename,_=share.split_path(x)
        uimessage="deleting "..filename.."..."
        redraw()
        x = share.trim_virtual_directory(x)
        foo=share.splitstr(x,"/")
        datatype=foo[1]
        username=foo[2]
        dataname=foo[3]
        msg = share._delete(username,datatype,dataname)
        show_message(msg)
        print(x)
        os.execute("rm -rf "..share.get_virtual_directory())
        refreshed_dir = false
      end)
    end
  end
end

function enc(n,z)
  mode=util.clamp(mode+sign(z),1,5)
  redraw()
end

function redraw()
  screen.clear()

  screen.level(4)
  screen.font_face(1)
  screen.font_size(8)
  screen.move(1,8)
  if settings.name then
    screen.text("norns.online/"..settings.name)
  else
    screen.text("norns.online")
  end

  start_point=11
  if not settings.name then
    screen.level(15)
    screen.font_face(1)
    screen.font_size(8)
    screen.move(0,start_point+11)
    screen.text(">")
    screen.move(7,start_point+1*11)
    screen.text("register")
  else
    screen.font_face(1)
    screen.font_size(8)
    for i=1,5 do
      if mode==i then
        local j = i 
        if j >2 then
          j = j-1
        end
        screen.level(15)
        screen.move(0,start_point+j*11)
        screen.text(">")
      else
        screen.level(4)
      end
      if i==1 then
        screen.move(7,start_point+1*11)
        screen.text("upload tape")
      elseif i==2 or i==3 then
        if mode ==2 or mode ==3 then 
          screen.level(15)
        else
          screen.level(4)
        end
        screen.move(7,start_point+2*11)
        screen.text("download ")
        screen.move(7+40,start_point+2*11)
        if mode==3 then 
          screen.level(4)
        elseif mode==2 then
          screen.level(15)
        end
        screen.text("tape")
        screen.move(7+61,start_point+2*11)
        screen.level(4)
        screen.text("or")
        if mode==3 then 
          screen.level(15)
        elseif mode==2 then 
          screen.level(4)
        end
        screen.move(7+71,start_point+2*11)
        screen.text("script")
      elseif i==4 then
        screen.move(7,start_point+3*11)
        screen.text("delete something")
      elseif i== 5 then
        screen.move(7,start_point+4*11)
        if util.file_exists(KILL_FILE) then
          screen.text("go offline")
          x=110
          y=start_point+4*11-6
          w=30
          screen.level(15)
          screen.rect(x-w/2,y,w,10)
          screen.fill()
          screen.level(15)
          screen.rect(x-w/2,y,w,10)
          screen.stroke()
          screen.move(x,y+7)
          screen.level(0)
          screen.text_center("LIVE")
        else
          screen.text("go online")
        end
      end
    end
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

function refresh_directory()
  if not refreshed_dir then
    refreshed_dir=true 
    uimessage="refreshing directory..."
    redraw()
    share.make_virtual_directory()
    uimessage=""
    redraw()
  end
end

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
  if settings.allowroom then
    params:set("allowroom",2)
  else
    params:set("allowroom",1)
  end
  params:set("roomname",settings.room)
  params:set("roomsize",settings.roomsize)
  params:set("framerate",settings.framerate)
  params:set("buffertime",settings.buffertime)
  params:set("packetsize",settings.packetsize)
  params:set("roomvolume",settings.roomvolume)
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
      show_message("you are online, you can play!")
    end)
    start()
  end
end

function start()
  if not util.file_exists(SERVER_FILE) then 
    norns_online_update()
  end
  print("starting")
  write_settings()
  make_start_sh()
  os.execute(START_FILE)
  clock.run(function()
    for i=1,10 do
      clock.sleep(0.1)
      redraw()
    end
  end)
end

function stop()
  print("stopping server...")
  os.execute(KILL_FILE)
  clock.run(function()
    for i=1,10 do
      clock.sleep(0.1)
      redraw()
    end
    os.remove(KILL_FILE)
  end)
end

function make_start_sh()
  startsh="#!/bin/bash\n"
  startsh=startsh..CODE_DIR.."norns.online --config "..CONFIG_FILE.." > /dev/null &\n"
  f=io.open(START_FILE,"w")
  f:write(startsh)
  f:close(f)
  os.execute("chmod +x "..START_FILE)
end

function install_prereqs()
  -- install the main program
  if not util.file_exists(SERVER_FILE) then
    norns_online_update()
  end
  print(os.capture("ffmpeg --help 2>&1"))
  print(os.capture("mpv --version 2>&1"))
  missingffmpeg=string.match(os.capture("ffmpeg --help 2>&1"),"not found")
  missingmpv=string.match(os.capture("mpv --version 2>&1"),"not found")
  if missingffmpeg or missingmpv then
    -- install ffmpeg
    uimessage="installing ffmpeg and mpv..."
    redraw()
    os.execute("sudo apt update")
    uimessage="please wait about 2min..."
    redraw()
    os.execute("sudo apt install -y mpv ffmpeg")
    uimessage=""
    redraw()
  end
  missingffmpeg=string.match(os.capture("ffmpeg --help 2>&1"),"not found")
  missingmpv=string.match(os.capture("mpv --version 2>&1"),"not found")
  if missingffmpeg and missingmpv then
    show_message("still missing mpv and ffmpeg")
  elseif missingmpv then
    show_message("still missing mpv")
  elseif missingffmpeg then
    show_message("still missing ffmpeg")
  else
    settings.is_installed=true
    write_settings()
  end
end

function norns_online_update()
  os.execute("rm -f "..SERVER_FILE)
  uimessage="updating server..."
  redraw()
  os.execute("cd "..CODE_DIR.." && git pull")
  -- uimessage="building..."
  -- redraw()
  -- os.execute("cd "..CODE_DIR.."; /usr/local/go/bin/go build")
  uimessage=""
  redraw()
  if not util.file_exists(SERVER_FILE) then
    s = os.capture("cat "..CODE_DIR.."norns.online.lua | grep LATEST_RELEASE")
    latest_release = string.match(s, 'LATEST_RELEASE="([^"]+)')
    uimessage="downloading norns.online..."
    redraw()
    os.execute("curl -L "..latest_release.." -o "..SERVER_FILE)
    os.execute("chmod +x "..SERVER_FILE)
    uimessage=""
    redraw()
  end
  if util.file_exists(SERVER_FILE) then
    show_message("updated.")
  end
end

--
-- sharing server stuff
--



function upload_callback(pathtofile)
  if pathtofile=="cancel" then
    do return end
  end
  _,filename,_=share.split_path(pathtofile)
  uimessage="uploading "..filename.."..."
  target="/home/we/dust/audio/share/"..settings.name.."/"..filename
  redraw()
  msg = share._upload(settings.name,"tape",filename,pathtofile,target)
  if string.match(msg,"need to register") then
    settings.is_registered=false
    write_settings()
  end
  show_message(msg)
  refreshed_dir=false
end

function server_register()
  if not share.is_registered(settings.name) then
    uimessage="registering "..settings.name.."..."
    redraw()
    msg=share.register(settings.name)
    show_message(msg)
    if string.match(msg,"OK") then
      settings.is_registered=true
    else
      settings.is_registered=false
    end
      write_settings()
  end
end

function server_generate_key()
  textentry.enter(function(x)
    if x~=nil then
      uimessage="generating keypair..."
      redraw()
      settings.name=x
      share.generate_keypair(settings.name)
      uimessage=""
      redraw()
      server_register()
    end
  end,"","enter public name:")
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
    clock.sleep(1)
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

function os.capture(cmd)
  local f=assert(io.popen(cmd,'r'))
  local s=assert(f:read('*a'))
  f:close()
  return s
end



