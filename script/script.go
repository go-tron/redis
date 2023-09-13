package script

import "github.com/redis/go-redis/v9"

var Test = redis.NewScript(`
return redis.call('hdel', 'test1','a')
`)

var GetDel = redis.NewScript(`
local current = redis.call('get', KEYS[1]);
if (current) then
	redis.call('del', KEYS[1]);
end
return current;
`)

var HGetDel = redis.NewScript(`
local current = redis.call('hget', KEYS[1], KEYS[2]);
if (current) then
	redis.call('hdel', KEYS[1], KEYS[2]);
end
return current;
`)

var HIncrLimit = redis.NewScript(`
--[[/*
* KEYS[1] Key
* KEYS[2] filed
* ARGV[1] 增加数量
* ARGV[2] 上限数量
* ARGV[3] 锁定时间
*/]]
if ARGV[3] and tonumber(ARGV[3]) > 0 then
    if not redis.call('set', KEYS[1] .. '-lock:' .. KEYS[2], ARGV[1], 'NX', 'EX', ARGV[3]) then
        return redis.error_reply("locked")
    end
end

local incr = tonumber(ARGV[1])
if incr <= 0 then
    return redis.error_reply("incr must be positive")
end
local total = tonumber(ARGV[2])
if total <= 0 then
    return redis.error_reply("limit must be positive")
end
local curr = redis.call('hget', KEYS[1], KEYS[2])
if curr then
    if tonumber(curr) + tonumber(ARGV[1]) > tonumber(ARGV[2]) then
        return redis.error_reply("reach limit")
    end
end
return redis.call('hincrby', KEYS[1], KEYS[2], ARGV[1])
`)

var IncrExpire = redis.NewScript(`
--[[/*
* KEYS[1] Key
* ARGV[1] 过期时间
* redis.log(redis.LOG_NOTICE,"log text")
*/]]
local result = redis.call('incr', KEYS[1])
if result == 1 and ARGV[1] and tonumber(ARGV[1]) > 0 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return result
`)

var FrequencyLimit = redis.NewScript(`
--[[/*
* KEYS[1] Key
* ARGV[1] max
* ARGV[2] 过期时间
* redis.log(redis.LOG_NOTICE,"log text")
*/]]
if ARGV[1] and tonumber(ARGV[1]) > 0 then
	local curr = redis.call('get', KEYS[1])
	if curr then
		if tonumber(curr) + 1 > tonumber(ARGV[1]) then
			return redis.error_reply("reach limit")
		end
	end
end
local result = redis.call('incr', KEYS[1])
if result == 1 and ARGV[2] and tonumber(ARGV[2]) > 0 then
    redis.call('EXPIRE', KEYS[1], ARGV[2])
end
return result
`)

var BatchGet = redis.NewScript(`
--[[/*
* KEYS[1] key
* KEYS[2] 名额key
* result  数组
*/]]

local result = {}
for i, v in pairs(ARGV) do
    result[i] = redis.call('get', KEYS[1] .. ':' .. v)
	if not result[i] then
		result[i] = ''
	end
end

return cjson.encode(result)
`)

var BatchLock = redis.NewScript(`
--[[/*
* KEYS 数组
* ARGV[1] 过期时间
*/]]

for i, v in pairs(KEYS) do
    if redis.call('get', v) then
        return v
    end
end

for i, v in pairs(KEYS) do
	if ARGV[1] and tonumber(ARGV[1]) > 0 then
		redis.call('set', v, 1, 'NX', 'EX', ARGV[1])
	else
    	redis.call('setnx', v, 1)
	end
end

return ""
`)

var BatchUnlock = redis.NewScript(`
--[[/*
* KEYS 数组
*/]]

for i, v in pairs(KEYS) do
    redis.call('del', v)
end
`)

var HDelIsKeyDeleted = redis.NewScript(`
--[[/*
* KEYS[1] key
* ARGV[1] field
*/]]
redis.call('hdel', KEYS[1], ARGV[1])
if (redis.call('exists', KEYS[1]) == 1) then
	return 0
else
	return 1
end
`)
