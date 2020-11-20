-- norns.online v0.1.0
-- remote control for your norns
--
-- llllllll.co/t/norns.online
--
--
--
--    ▼ instructions below ▼
--
--

local json=include("lib/json")
local textentry=require 'textentry'

CONFIG_FILE="/home/we/dust/code/norns.online/config.json"
KILL_FILE="/tmp/norns.online.kill"
START_FILE="/home/we/dust/code/norns.online/start.sh"
LATEST_RELEASE=""
px=48
py=16
name=""

function init()
  params:add{type='binary',name='allow menu',id='allowmenu',behavior='toggle',allow_pmap=true,action=function(v) update_settings() end}
  params:add{type='binary',name='keep awake',id='keepawake',behavior='toggle',allow_pmap=true,action=function(v) update_settings() end}
  
  load_settings()
  redraw()
end

function key(n,z)
  if n==1 and z==1 then
    textentry.enter(name,"","enter name:")
  elseif n==2 and z==1 then
    params:delta('menu',z)
  elseif n==3 then
    params:delta('awake',z)
  end
  redraw()
end

function redraw()
  screen.move(py+px,py)
  screen.level(15)
  screen.text('name: ')
  screen.move(py,py*2)
  screen.level(params:get('allowmenu')==1 and 15 or 2)
  screen.text('allow menu')
  screen.move(py+px,py*2)
  screen.level(params:get('keepawake')==1 and 15 or 2)
  screen.text('keep awake')
  screen.move(py-px,py*0.5)
  if util.file_exists(KILL_FILE) then
    screen.text('stop')
  else
    screen.text('start')
  end
  screen.update()
end

--
-- utils
--

function readAll(file)
  local f=assert(io.open(file,"rb"))
  local content=f:read("*all")
  f:close()
  return content
end

--
--
--

function update_settings()
  redraw()
  write_settings()
end

function write_settings()
  dat={}
  dat.name=name
  dat.menu=false
  dat.keepawake=false
  if params:get("allowmenu")==1 then
    dat.menu=true
  end
  if params:get("keepawake")==1 then
    dat.keepawake=true
  end
  jsondata=json.encode(dat)
  f=io.open(CONFIG_FILE,"w")
  f:write(jsondata)
  f:close(f)
end

function load_settings()
  if not util.file_exists(CONFIG_FILE) then
    do return end
  end
  data=readAll(CONFIG_FILE)
  dat=json.decode(data)
  name=dat.name
  if dat.menu then
    params:set("allowmenu",1)
  else
    params:set("allowmenu",0)
  end
  if dat.keepawake then
    params:set("keepawake",1)
  else
    params:set("keepawake",0)
  end
end

function update()
  os.execute("curl "+LATEST_RELEASE+" -o /home/we/dust/code/norns.online/norns.online")
end

function start()
  write_settings()
  os.execute(START_FILE)
  redraw()
end

function stop()
  os.execute(KILL_FILE)
  redraw()
end
