-- share.norns.online v0.0.0
-- 
--
-- 
-- 
-- 
-- 
--    ▼ instructions below ▼
--

local json=include("lib/json")
local textentry=require 'textentry'
local fileselect = require 'fileselect'
local share = include("share.norns.online/lib/share")
local textentry=require 'textentry'

username=""
uimessage=""
dir = {}
mode = 1
function init()

end


function download_callback(path)
	if path=="cancel" then 
		do return end 
	end
	p = "/dev/shm/share.norns.online/"
	local path = (path:sub(0, #p) == p) and path:sub(#p+1) or path
	print(path)
	foo = splitstr(path,"/")
	datatype = foo[1]
	username = foo[2]
	dataname = foo[3]
  	uimessage="downloading "..dataname.."..."
  	redraw()
	msg = share.download(datatype,username,dataname)
	show_message(msg)
end

function upload_callback(path)
	if path=="cancel" then 
		do return end 
	end
	-- https://stackoverflow.com/questions/5243179/what-is-the-neatest-way-to-split-out-a-path-name-into-its-components-in-lua
	_,dataname,_ = string.match(path, "(.-)([^\\/]-%.?([^%.\\/]*))$")
  	uimessage="uploading "..dataname.."..."
  	targetdir="/home/we/dust/audio/share/"..username
  	redraw()
	show_message(share.upload("tape",dataname,path,targetdir))
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
	  end
	end,"","enter public name:")
end


function key(k,z)
	if z == 0 then 
		do return end
	end
	if k==3 then
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
		if mode == 1 then 
			username = share.key_established()
			if not username then 
				server_generate_key()
			else
				-- upload
				if not share.is_registered() then
			      	uimessage="registering..."
			      	redraw()
					msg = share.register(x)
					show_message(msg)
					if not string.match(msg,"OK") then 
						do return end
					end
			    end
			    fileselect.enter("/home/we/dust/audio", upload_callback)
			end
		else
			-- download
			-- make fake folder structure in /dev/shm
	      	uimessage="getting directory..."
	      	redraw()
			dir = share.directory()
	      	uimessage=""
	      	redraw()
			os.execute("rm -rf /dev/shm/share.norns.online")
			for _,s in ipairs(dir) do
				tab.print(s)
				os.execute("mkdir -p /dev/shm/share.norns.online/"..s.type.."/"..s.username)
				os.execute("touch /dev/shm/share.norns.online/"..s.type.."/"..s.username.."/"..s.dataname)
			end
			fileselect.enter("/dev/shm/share.norns.online/", download_callback)
		end
	end
end

function enc(k,z)
	mode = util.clamp(mode+z,1,2)
	redraw()
end


function redraw()
  screen.clear() 

 screen.move(20,20)
 if mode ==1 then 
 	screen.level(15)
 else
 	screen.level(4)
 end
 screen.text("upload")
 if mode == 2 then 
 	screen.level(15)
 else
 	screen.level(4)
 end
 screen.move(20,40)
 screen.text("download")

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

function splitstr(inputstr, sep)
        if sep == nil then
                sep = "%s"
        end
        local t={}
        for str in string.gmatch(inputstr, "([^"..sep.."]+)") do
                table.insert(t, str)
        end
        return t
end
