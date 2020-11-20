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

cs=require 'controlspec'

px=48
py=16

function init()
  params:add{type='binary',name='allow menu',id='menu',behavior='toggle',allow_pmap=true,action=function(v) print('tog: '..tostring(v)) redraw() end}
  params:add{type='binary',name='keep awake',id='awake',behavior='toggle',allow_pmap=true,action=function(v) print('tog: '..tostring(v)) redraw() end}
  
  redraw()
end

function key(n,z)
  if n==1 and z==1 then
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
  screen.level(params:get('menu')==1 and 15 or 2)
  screen.text('allow menu')
  screen.move(py+px,py*2)
  screen.level(params:get('awake')==1 and 15 or 2)
  screen.text('keep awake')
  screen.move(py-px,py*0.5)
  if util.file_exists("/tmp/norns.online.kill") then
    screen.text('stop')
  else
    screen.text('start')
  end
  screen.update()
end
