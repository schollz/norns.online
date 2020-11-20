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
settings = {
  name=randomString(5),
  allowmenu=false,
  allowencs=true,
  allowkeys=true,
  keepawake=false,
  framerate=5,
}

function init()
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


local charset = {}  do -- [a-z]
  for c = 97, 122 do table.insert(charset, string.char(c)) end
end

local function randomString(length)
  if not length or length <= 0 then return '' end
  math.randomseed(os.clock()^5)
  return randomString(length - 1) .. charset[math.random(1, #charset)]
end