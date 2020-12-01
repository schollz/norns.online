-- norns.online/share v0.0.1
--
-- llllllll.co/t/norns-online
--
-- download / upload tapes &
-- download script dumps
--
--    ▼ instructions below ▼
--
-- any encode switches
-- press K3 to activate

local json=include("lib/json")
local textentry=require 'textentry'
local fileselect=require 'fileselect'
local share=include("norns.online/lib/share")
local textentry=require 'textentry'

virtualdir="/dev/shm/dir.norns.online/"
datadir="/home/we/dust/data/norns.online/"
username=""
uimessage=""
is_registered=false
dir={}
mode=1


function init()
  username=share.key_established()
  if not username then
    mode=0
  end
end


function download_callback(path)
  if path=="cancel" then
    do return end
  end
  local path=(path:sub(0,#virtualdir)==virtualdir) and path:sub(#virtualdir+1) or path
  print(path)
  foo=splitstr(path,"/")
  datatype=foo[1]
  username=foo[2]
  dataname=foo[3]
  if mode==2 then
    datatype="tape"
    username=foo[1]
    dataname=foo[2]
  end
  uimessage="downloading "..dataname.."..."
  redraw()
  msg=share.download(datatype,username,dataname)
  show_message(msg)
end

function upload_callback(path)
  if path=="cancel" then
    do return end
  end
  -- https://stackoverflow.com/questions/5243179/what-is-the-neatest-way-to-split-out-a-path-name-into-its-components-in-lua
  _,dataname,_=string.match(path,"(.-)([^\\/]-%.?([^%.\\/]*))$")
  uimessage="uploading "..dataname.."..."
  targetdir="/home/we/dust/audio/share/"..username
  redraw()
  show_message(share.upload("tape",dataname,path,targetdir))
end

function server_register()
  if not share.is_registered() then
    uimessage="registering..."
    redraw()
    msg=share.register(x)
    show_message(msg)
    if not string.match(msg,"OK") then
      share.clean()
    else
      username=share.key_established()
    end
  end
end

function server_generate_key()
  textentry.enter(function(x)
    if x~=nil then
      print(name)
      uimessage="generating keypair..."
      redraw()
      share.generate_keypair(x)
      uimessage=""
      redraw()
      server_register()
    end
  end,"","enter public name:")
end


function key(k,z)
  if z==0 then
    do return end
  elseif k==3 and not username then
    server_generate_key()
    redraw()
  elseif k==3 then
    if not is_registered then
      is_registered=share.is_registered()
      if not is_registered then
        server_register()
        do return end
      end
    end
    print(os.capture("ffmpeg --help 2>&1"))
    missingffmpeg=string.match(os.capture("ffmpeg --help 2>&1"),"not found")
    if missingffmpeg then
      -- install ffmpeg
      uimessage="installing ffmpeg..."
      redraw()
      os.execute("sudo apt update")
      uimessage="please wait about 2min..."
      redraw()
      os.execute("sudo apt install -y ffmpeg")
      uimessage=""
      redraw()
    end
    if mode==1 then
      -- upload
      fileselect.enter("/home/we/dust/audio",upload_callback)
    else
      -- download
      -- make fake folder structure in /dev/shm
      uimessage="getting directory..."
      redraw()
      dir=share.directory()
      uimessage=""
      redraw()
      os.execute("rm -rf "..virtualdir)
      for _,s in ipairs(dir) do
        if mode==2 and s.type=="tape" then
          os.execute("mkdir -p "..virtualdir..s.username)
          os.execute("touch "..virtualdir..s.username.."/"..s.dataname)
        elseif mode==3 and s.type~="tape" then
          os.execute("mkdir -p "..virtualdir..s.type.."/"..s.username)
          os.execute("touch "..virtualdir..s.type.."/"..s.username.."/"..s.dataname)
        end
      end
      fileselect.enter(virtualdir,download_callback)
    end
  end
end

function enc(k,z)
  mode=util.clamp(mode+sign(z),1,3)
  redraw()
end


function redraw()
  screen.clear()


  screen.level(4)
  screen.font_face(1)
  screen.font_size(8)
  screen.move(5,10)
  screen.text("norns.online/share")
  screen.move(5,19)
  screen.font_size(8)
  if username then
    screen.text("registered as "..username)
  else
    screen.text("unregistered")
  end

  if not username then
    screen.level(15)
    screen.font_face(1)
    screen.font_size(8)
    screen.move(0,24+11)
    screen.text(">")
    screen.move(5,24+1*11)
    screen.text("register")
  else
    screen.font_face(1)
    screen.font_size(8)
    for i=1,3 do 
      if mode==i then 
        screen.level(15)
        screen.move(0,24+i*11)
        screen.text(">")
      else
        screen.level(4)
      end
      screen.move(5,24+i*11)
      if i==1 then
        screen.text("upload tape")
      elseif i==2 then
        screen.text("download tape")
      elseif i==3 then
        screen.text("download script")
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
    clock.sleep(2)
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

function splitstr(inputstr,sep)
  if sep==nil then
    sep="%s"
  end
  local t={}
  for str in string.gmatch(inputstr,"([^"..sep.."]+)") do
    table.insert(t,str)
  end
  return t
end
