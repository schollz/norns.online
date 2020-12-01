-- share.lua
local share={debug=true}
local json=include("norns.online/lib/json")

datadir="/home/we/dust/data/norns.online/"
server_name="https://norns.online"

share.log=function(...)
  local arg={...}
  if share.debug and arg~=nil then
    printResult=""
    for i,v in ipairs(arg) do
      printResult=printResult..tostring(v).." "
    end
    print(printResult)
  end
end

share.key_established=function()
  if file_exists(datadir.."username") and file_exists(datadir.."key.public") then
    return os.capture("cat "..datadir.."username")
  end
  return false
end

share.generate_keypair=function(username)
  os.execute("mkdir -p "..datadir)
  os.execute("openssl genrsa -out "..datadir.."key.private 2048")
  os.execute("openssl rsa -in "..datadir.."key.private -pubout -out "..datadir.."key.public")
  f=io.open(datadir.."username","w")
  f:write(username)
  f:close()
end

share.download=function(datatype,username,dataname)
  -- check signature
  result=os.capture("curl -s "..server_name.."/share/"..datatype.."/"..username.."/"..dataname.."/metadata.json")
  print(result)
  metadata=json.decode(result)
  if metadata==nil then
    return "bad metadata"
  end
  for _,file in ipairs(metadata.files) do
    filename=file.name
    result=os.capture("curl -s -o /dev/shm/"..filename.." "..server_name.."/share/"..datatype.."/"..username.."/"..dataname.."/"..file.name)
    if ends_with(filename,".wav.flac") then
      -- convert back to wav
      new_filename=filename:gsub(".flac".."$","")
      os.execute("ffmpeg -y -i /dev/shm/"..filename.." -ar 48000 -c:a pcm_s24le /dev/shm/"..new_filename)
      os.remove("/dev/shm/"..filename)
      filename=new_filename
    end
    -- TODO: verify
    -- make target directory
    os.execute("mkdir -p "..file.targetdir)
    os.execute("mv /dev/shm/"..filename.." "..file.targetdir.."/"..filename)
  end
  return "..downloaded"
end

share.is_registered=function()
  local username=os.capture("cat "..datadir.."username")
  if username==nil then
    return
  end
  local publickey=os.capture("cat "..datadir.."key.public")
  if publickey==nil then
    return
  end
  curl_url=server_name.."/share/keys/"..username
  curl_cmd="curl -s "..curl_url
  result=os.capture(curl_cmd)
  return result==publickey
end

share.directory=function()
  curl_url=server_name.."/directory.json"
  curl_cmd="curl -s "..curl_url
  result=os.capture(curl_cmd)
  print(result)
  return json.decode(result)
end

share.register=function()
  tmp_signature=temp_file_name()
  tmp_username=temp_file_name()
  local username=os.capture("cat "..datadir.."username")
  if username==nil then
    return
  end

  -- sign the username
  local f=io.open(tmp_username,"w")
  f:write(username)
  f:close()
  os.execute("openssl dgst -sign "..datadir.."key.private -out "..tmp_signature.." "..tmp_username)
  signature=os.capture("base64 -w 0 "..tmp_signature)


  curl_url=server_name.."/register?username="..username.."&signature="..signature
  curl_cmd="curl -s --upload-file "..datadir.."key.public "..'"'..curl_url..'"'
  print(curl_cmd)
  result=os.capture(curl_cmd)
  print(result)
  os.remove(tmp_signature)
  os.remove(tmp_username)
  return result
end

share.unregister=function()
  tmp_signature=temp_file_name()
  tmp_username=temp_file_name()
  local username=os.capture("cat "..datadir.."username")
  if username==nil then
    return
  end

  -- sign the username
  f=io.open(tmp_username,"w")
  f:write(username)
  f:close()
  os.execute("openssl dgst -sign "..datadir.."key.private -out "..tmp_signature.." "..tmp_username)
  signature=os.capture("base64 -w 0 "..tmp_signature)

  -- send unregistration
  curl_url=server_name.."/unregister?username="..username.."&signature="..signature
  curl_cmd="curl -s --upload-file "..datadir.."key.public "..'"'..curl_url..'"'
  print(curl_cmd)
  result=os.capture(curl_cmd)
  print(result)

  os.remove(tmp_signature)
  os.remove(tmp_username)
  return result
end

share.upload=function(type,dataname,filename,targetdir)
  tmp_signature=temp_file_name()
  tmp_hash=temp_file_name()
  local username=os.capture("cat "..datadir.."username")
  if username==nil then
    return
  end

  -- convert wav to flac, if it is a wav
  flaced=false
  if ends_with(filename,".wav") then
    os.execute("ffmpeg -y -i "..filename.." -ar 48000 "..filename..".flac")
    filename=filename..".flac"
    flaced=true
  end

  -- hash the data
  hash=os.capture("sha256sum "..filename)
  hash=hash:firstword()
  print("hash: "..hash)
  f=io.open(tmp_hash,"w")
  f:write(hash)
  f:close()

  -- sign the hash
  os.execute("openssl dgst -sign "..datadir.."key.private -out "..tmp_signature.." "..tmp_hash)
  signature=os.capture("base64 -w 0 "..tmp_signature)

  -- upload the file and metadata
  curl_url=server_name.."/upload?type="..type.."&username="..username.."&dataname="..dataname.."&filename="..filename.."&targetdir="..targetdir.."&hash="..hash.."&signature="..signature
  curl_cmd="curl -s --upload-file "..filename..' "'..curl_url..'"'
  print(curl_cmd)
  result=os.capture(curl_cmd)
  print(result)

  -- clean up
  os.remove(tmp_signature)
  os.remove(tmp_hash)
  if flaced then
    os.remove(filename) -- remove if we converted
  end
  return result
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


--
-- utilities
--
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

function temp_file_name()
  return "/dev/shm/"..randomString(5)
end

function file_exists(fname)
  local f=io.open(fname,"r")
  if f~=nil then io.close(f) return true else return false end
end


return share
