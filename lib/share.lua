-- share.lua
local share={
  SHARE_DATA_DIR=_path.data.."norns.online/",
  CONFIG_FILE=_path.data.."norns.online/config.json",
  VIRTUAL_DIR=_path.data.."norns.online/virtualdir/",
  server_name="https://norns.online",
}
local json=include("norns.online/lib/json")


os.execute("mkdir -p "..share.VIRTUAL_DIR)


--
-- virtual directory
--

share.get_remote_directory=function()
  curl_url=share.server_name.."/directory.json"
  curl_cmd="curl -s -m 5 "..curl_url
  result=os.capture(curl_cmd)
  print(result)
  if result=="" then
    do return nil end
  end
  return json.decode(result)
end

share.get_virtual_directory=function(datatype)
  if datatype==nil then
    do return share.VIRTUAL_DIR end
  end
  return share.VIRTUAL_DIR..datatype.."/"
end

share.make_virtual_directory=function()
  -- get remove files
  remotedir=share.get_remote_directory()
  if remotedir==nil then
    do return nil end
  end

  -- get current virutal directory
  curdir={}
  for _,f in ipairs(list_files(share.VIRTUAL_DIR)) do
    curdir[share.trim_prefix(f,share.VIRTUAL_DIR)]=true
  end

  -- make new ones
  for k,_ in pairs(remotedir) do
    if curdir[k]==nil then
      path,filename,_=share.split_path(k)
      print("making "..share.VIRTUAL_DIR..path..filename)
      -- os.execute("mkdir -p "..share.VIRTUAL_DIR..path)
      -- os.execute("touch "..share.VIRTUAL_DIR..path..filename)
    end
  end

  -- delete missing ones
  for k,_ in pairs(curdir) do
    if remotedir[k]==nil then
      print("removing "..share.VIRTUAL_DIR..k)
      -- os.execute("rm -rf "..share.VIRTUAL_DIR..k)
    end
  end

  return share.VIRTUAL_DIR
end

share.trim_virtual_directory=function(path)
  local path=(path:sub(0,#share.VIRTUAL_DIR)==share.VIRTUAL_DIR) and path:sub(#share.VIRTUAL_DIR+1) or path
  return path
end

share.download_from_virtual_directory=function(path)
  if path=="cancel" then
    do return end
  end
  path=share.trim_virtual_directory(path)
  foo=share.splitstr(path,"/")
  datatype=foo[1]
  username=foo[2]
  dataname=foo[3]
  msg=share.download(datatype,username,dataname)
  print(msg)
  return msg
end

--
-- registration
--
share.get_username=function()
  -- returns username
  if not util.file_exists(share.CONFIG_FILE) then
    do return nil end
  end
  local f=assert(io.open(share.CONFIG_FILE,"rb"))
  local data=f:read("*all")
  f:close()
  settings=json.decode(data)
  return settings.name
end

share.generate_keypair=function(username)
  os.execute("mkdir -p "..share.SHARE_DATA_DIR)
  os.execute("openssl genrsa -out "..share.SHARE_DATA_DIR.."key.private 2048")
  os.execute("openssl rsa -in "..share.SHARE_DATA_DIR.."key.private -pubout -out "..share.SHARE_DATA_DIR.."key.public")
end

share.is_registered=function(username)
  local publickey=os.capture("cat "..share.SHARE_DATA_DIR.."key.public")
  if publickey==nil then
    return
  end
  curl_url=share.server_name.."/share/keys/"..username
  curl_cmd="curl -s -m 5 "..curl_url
  result=os.capture(curl_cmd)
  return result==publickey
end

share.register=function(username)
  tmp_signature=share.temp_file_name()
  tmp_username=share.temp_file_name()

  -- write username to file
  print("signing "..username)
  local f=io.open(tmp_username,"w")
  f:write(username)
  f:close()

  -- create signature
  os.execute("openssl dgst -sign "..share.SHARE_DATA_DIR.."key.private -out "..tmp_signature.." "..tmp_username)
  signature=os.capture("base64 -w 0 "..tmp_signature)

  curl_url=share.server_name.."/register?username="..username.."&signature="..signature
  curl_cmd="curl -s -m 5 --upload-file "..share.SHARE_DATA_DIR.."key.public "..'"'..curl_url..'"'
  print(curl_cmd)
  result=os.capture(curl_cmd)
  print(result)
  os.remove(tmp_signature)
  os.remove(tmp_username)
  return result
end

share.unregister=function(username)
  tmp_signature=share.temp_file_name()
  tmp_username=share.temp_file_name()

  -- sign the username
  f=io.open(tmp_username,"w")
  f:write(username)
  f:close()
  os.execute("openssl dgst -sign "..share.SHARE_DATA_DIR.."key.private -out "..tmp_signature.." "..tmp_username)
  signature=os.capture("base64 -w 0 "..tmp_signature)

  -- send unregistration
  curl_url=share.server_name.."/unregister?username="..username.."&signature="..signature
  curl_cmd="curl -s -m 5 --upload-file "..share.SHARE_DATA_DIR.."key.public "..'"'..curl_url..'"'
  print(curl_cmd)
  result=os.capture(curl_cmd)
  print(result)

  os.remove(tmp_signature)
  os.remove(tmp_username)
  return result
end

--
-- uploading/downloading
--

share._upload=function(username,type,dataname,pathtofile,target)
  -- type is the type, e.g. tape / barcode (name of script) / etc.
  -- dataname is how the group of data can be represented
  -- pathtofile is the path to the file on this norns
  -- target is the target path to file on any norns that downloads it
  tmp_signature=share.temp_file_name()
  tmp_hash=share.temp_file_name()

  _,filename,ext=share.split_path(pathtofile)
  print("ext: "..ext)

  -- convert wav to flac, if it is a wav
  flaced=false
  if ext=="wav" then
    os.execute("ffmpeg -y -i "..pathtofile.." -ar 48000 /dev/shm/"..filename..".flac")
    -- update the pathname and filename (but not the target path)
    pathtofile="/dev/shm/"..filename..".flac"
    _,filename,_=share.split_path(pathtofile)
    flaced=true
    ext="wav.flac"
  end

  -- hash the data
  hash=os.capture("sha256sum "..pathtofile)
  hash=hash:firstword()
  hashed_filename=string.sub(hash,1,9).."."..ext
  f=io.open(tmp_hash,"w")
  f:write(hashed_filename)
  f:write(target)
  f:write(hash)
  f:close()


  print(os.capture("cat "..tmp_hash))
  print("pathtofile: "..pathtofile)

  -- sign the hash
  os.execute("openssl dgst -sign "..share.SHARE_DATA_DIR.."key.private -out "..tmp_signature.." "..tmp_hash)
  signature=os.capture("base64 -w 0 "..tmp_signature)

  -- upload the file and metadata
  curl_url=share.server_name.."/upload?type="..type.."&username="..username.."&dataname="..dataname.."&filename="..hashed_filename.."&target="..target.."&hash="..hash.."&signature="..signature
  curl_cmd="curl -s -m 5 --upload-file "..pathtofile..' "'..curl_url..'"'
  print(curl_cmd)
  result=os.capture(curl_cmd)
  print(result)

  -- clean up
  os.remove(tmp_signature)
  os.remove(tmp_hash)
  if flaced then
    os.remove(pathtofile) -- remove if we converted
  end
  return result
end


share._delete=function(username,type,dataname)
  -- type is the type, e.g. tape / barcode (name of script) / etc.
  -- dataname is how the group of data can be represented
  -- pathtofile is the path to the file on this norns
  -- target is the target path to file on any norns that downloads it
  tmp_signature=share.temp_file_name()
  tmp_username=share.temp_file_name()

  f=io.open(tmp_username,"w")
  f:write(username)
  f:close()

  -- sign the hash
  os.execute("openssl dgst -sign "..share.SHARE_DATA_DIR.."key.private -out "..tmp_signature.." "..tmp_username)
  signature=os.capture("base64 -w 0 "..tmp_signature)

  -- upload the file and metadata
  curl_url=share.server_name.."/delete?type="..type.."&username="..username.."&dataname="..dataname.."&signature="..signature
  curl_cmd='curl -s -m 5 "'..curl_url..'"'
  print(curl_cmd)
  result=os.capture(curl_cmd)
  print(result)

  -- clean up
  os.remove(tmp_username)
  os.remove(tmp_username)
  return result
end


share.download=function(type,username,dataname)
  -- download metadata
  result=os.capture("curl -s -m 5 "..share.server_name.."/share/"..type.."/"..username.."/"..dataname.."/metadata.json")
  print(result)
  metadata=json.decode(result)
  if metadata==nil then
    return "bad metadata"
  end
  for _,file in ipairs(metadata.files) do
    target_dir,target_filename,_=share.split_path(file.target)
    -- make directory if it doesn't exist
    os.execute("mkdir -p "..target_dir)

    -- download
    result=""
    if ends_with(file.name,".wav.flac") then
      -- download to temp and convert to wav
      result=os.capture("curl -s -m 5 -o /tmp/"..file.name.." "..share.server_name.."/share/"..type.."/"..username.."/"..dataname.."/"..file.name)
      os.execute("ffmpeg -y -i /tmp/"..file.name.." -ar 48000 -c:a pcm_s24le "..file.target)
      os.remove("/tmp/"..file.name)
    else
      -- download directly to target
      result=os.capture("curl -s -m 5 -o "..file.target.." "..share.server_name.."/share/"..type.."/"..username.."/"..dataname.."/"..file.name)
    end
    -- TODO: verify
  end
  return "downloaded."
end



--
-- share uploader
--

share.new=function(self,o)
  -- uploader = share:new{script_name="oooooo"}
  -- defined parameters
  o=o or {}
  setmetatable(o,self)
  self.__index=self
  self.script_name=o.script_name
  self.upload_username=share.get_username()

  if self.upload_username==nil then
    print("not registered")
    do return nil end
  end
  if self.script_name==nil then
    print("no script_name defined")
    do return nil end
  end
  return o
end


share.upload=function(self,o)
  --
  -- uploader = share:new{script_name="hello"}
  -- uploader.upload{dataname=X,pathtofile=Y,target=Z}
  --
  if o.dataname==nil then
    print("need dataname")
    do return end
  end
  if o.pathtofile==nil then
    print("need pathtofile")
    do return end
  end
  if o.target==nil then
    print("need target")
    do return end
  end
  if self.upload_username==nil then
    print("not registered")
    do return end
  end
  if self.script_name==nil then
    print("no script_name defined")
    do return end
  end
  msg=share._upload(self.upload_username,self.script_name,o.dataname,o.pathtofile,o.target)
  print(msg)
  return msg
end

--
-- public utilities
--

share.trim_prefix=function(s,p)
  local t=(s:sub(0,#p)==p) and s:sub(#p+1) or s
  return t
end

share.write_file=function(fname,data)
  print("saving to "..fname)
  file=io.open(fname,"w+")
  io.output(file)
  io.write(data)
  io.close(file)
end

share.read_file=function(fname)
  local f=io.open(fname,"rb")
  local content=f:read("*all")
  f:close()
  return content
end

share.dump_table_to_json=function(fname,table_data)
  data=json.encode(table_data)
  share.write_file(fname,data)
end

share.load_table_from_json=function(json_file)
  data=share.read_file(json_file)
  if data=="" then
    do return nil end
  end
  return json.decode(data)
end

share.split_path=function(path)
  -- https://stackoverflow.com/questions/5243179/what-is-the-neatest-way-to-split-out-a-path-name-into-its-components-in-lua
  -- /home/zns/1.txt returns
  -- /home/zns/   1.txt   txt
  pathname,filename,ext=string.match(path,"(.-)([^\\/]-%.?([^%.\\/]*))$")
  return pathname,filename,ext
end

share.temp_file_name=function()
  return "/dev/shm/tempfile"..randomString(5)
end

share.splitstr=function(inputstr,sep)
  if sep==nil then
    sep="%s"
  end
  local t={}
  for str in string.gmatch(inputstr,"([^"..sep.."]+)") do
    table.insert(t,str)
  end
  return t
end



--
-- private utilities
--

function list_files(d,files,recursive)
  -- list files in a flat table
  if d=="." or d=="./" then
    d=""
  end
  if d~="" and string.sub(d,-1)~="/" then
    d=d.."/"
  end
  folders={}
  if recursive then
    local cmd="ls -ad "..d.."*/ 2>/dev/null"
    local f=assert(io.popen(cmd,'r'))
    local out=assert(f:read('*a'))
    f:close()
    for f in out:gmatch("%S+") do
      if not (string.match(f,"ls: ") or f=="../" or f=="./") then
        files=list_files(f,files)
      end
    end
  end
  do
    local cmd="ls -p "..d.." | grep -v /"
    local f=assert(io.popen(cmd,'r'))
    local out=assert(f:read('*a'))
    for f in out:gmatch("%S+") do
      table.insert(files,d..f)
    end
  end
  return files
end

function os.capture(cmd,raw)
  local f=assert(io.popen(cmd,'r'))
  local s=assert(f:read('*a'))
  f:close()
  if raw then return s end
  s=string.gsub(s,'^%s+','')
  s=string.gsub(s,'%s+$','')
  s=string.gsub(s,'[\n\r]+',' ')
  return s
end

function string:firstword()
  return self:match("^([%w]+)");-- matches the first word and returns it, or it returns nil
end

function ends_with(str,ending)
  return ending=="" or str:sub(-#ending)==ending
end

charset={} do -- [0-9a-zA-Z]
  for c=48,57 do table.insert(charset,string.char(c)) end
  for c=65,90 do table.insert(charset,string.char(c)) end
  for c=97,122 do table.insert(charset,string.char(c)) end
end

function randomString(length)
  if not length or length<=0 then return '' end
  math.randomseed(os.clock()^5)
  return randomString(length-1)..charset[math.random(1,#charset)]
end

return share
