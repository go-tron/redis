package script

import "github.com/redis/go-redis/v9"

var PointGet = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* ARGV[1] 积分编号
*/]]
local val = redis.call('get', KEYS[1] .. ":" .. ARGV[1])
local total = 0
if val then 
	total = tonumber(val)
end
return total
`)

var PointList = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* ARGV[1] 数组 []积分编号
* result 数组 v[1]:积分编号 v[2]:积分数量
*/]]
local ARGV_T = cjson.decode(ARGV[1])
local result = {}
for i, v in pairs(ARGV_T) do
    local val = redis.call('get', KEYS[1] .. ":" .. v)
	local total = 0
    if val then
        total = tonumber(val)
    end
    result[i] = {v, total}
end
return result
`)

var PointBatchEdit = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* KEYS[2] EditKey
* ARGV[1] 积分ID
* ARGV[2] v[1]:积分编号 v[2]:增减数量
* result v[1]:积分编号 v[2]:增减数量 v[3]:增减后总数量
*/]]
local ARGV_T = cjson.decode(ARGV[2])
for i, v in pairs(ARGV_T) do
	if v[2] == 0 then
        return redis.error_reply("数量不能为0")
    end
    if v[2] < 0 then
        local val = redis.call('get', KEYS[1] .. ":" .. v[1])
		local total = 0
		if val then
			total = tonumber(val)
		end
		if total + v[2] < 0 then
			return redis.error_reply('not enough:' .. v[1] .. ',' .. total)
		end
    end
end

if (redis.call('hsetnx', KEYS[2], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" ..  ARGV[1])
end

for i, v in pairs(ARGV_T) do
	v[3] = redis.call('incrby', KEYS[1] .. ":" .. v[1], v[2])
end

redis.call('hset', KEYS[2], ARGV[1], cjson.encode(ARGV_T))
return ARGV_T
`)

var PointBatchRevoke = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* KEYS[2] EditKey
* KEYS[3] EditRevokeKey
* ARGV[1] 积分ID
* result  v[1]:积分编号 v[2]:增加数量 v[3]:撤销后积分
*/]]
local record = redis.call('hget', KEYS[2], ARGV[1])
if not record then
    return redis.error_reply("not exists:" .. ARGV[1])
end

local ARGV_T = cjson.decode(record)
for i, v in pairs(ARGV_T) do
    local val = redis.call('get', KEYS[1] .. ":" .. v[1])
    local total = 0
    if val then
        total = tonumber(val)
    end
    if total - v[2] < 0 then
        return redis.error_reply('not enough:' .. v[1] .. ',' .. total)
    end
    v[2] = 0 - v[2]
end

if (redis.call('hsetnx', KEYS[3], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" ..  ARGV[1])
end

for i, v in pairs(ARGV_T) do
    v[3] = redis.call('incrby', KEYS[1] .. ":" .. v[1], v[2])
end

redis.call('hset', KEYS[3], ARGV[1], cjson.encode(ARGV_T))
return ARGV_T
`)

var PointBatchApply = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* KEYS[2] ApplyKey
* ARGV[1] 订单号
* ARGV[2] v[1]:积分编号 v[2]:增减数量
* result v[1]:积分编号 v[2]:增减数量 v[3]:增减后积分数量
*/]]
local ARGV_T = cjson.decode(ARGV[2])
for i, v in pairs(ARGV_T) do
    if v[2] >= 0 then
        return redis.error_reply("数量必须小于0")
    end
    local val = redis.call('get', KEYS[1] .. ":" .. v[1])
    local total = 0
    if val then
        total = tonumber(val)
    end
    if total + v[2] < 0 then
        return redis.error_reply('not enough:' .. v[1] .. ',' .. total)
    end
end

if (redis.call('hsetnx', KEYS[2], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

for i, v in pairs(ARGV_T) do
    v[3] = redis.call('incrby', KEYS[1] .. ":" .. v[1], v[2])
end
redis.call('hset', KEYS[2], ARGV[1], cjson.encode(ARGV_T))
return ARGV_T
`)

var PointBatchApplyConfirm = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* KEYS[2] ApplyKey
* KEYS[3] ApplyResultKey
* ARGV[1] 订单号
* result v[1]:积分编号 v[2]:增减数量 v[3]:增减后积分数量
*/]]
local record = redis.call('hget', KEYS[2], ARGV[1])
if not record then
    return redis.error_reply('not exists:' .. ARGV[1])
end

if (redis.call('hsetnx', KEYS[3], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

local ARGV_T = cjson.decode(record)
for i, v in pairs(ARGV_T) do
    v[3] = tonumber(redis.call('get', KEYS[1] .. ":" .. v[1]))
end

redis.call('hset', KEYS[3], ARGV[1], cjson.encode(ARGV_T))
redis.call('hdel', KEYS[2], ARGV[1])
return ARGV_T
`)

var PointBatchApplyCancel = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* KEYS[2] ApplyKey
* KEYS[3] ApplyResultKey
* ARGV[1] 订单号
* result v[1]:积分编号 v[3]:增减数量 v[4]:增减后积分数量
*/]]
local record = redis.call('hget', KEYS[2], ARGV[1])
if not record then
    return redis.error_reply('not exists:' .. ARGV[1])
end

if (redis.call('hsetnx', KEYS[3], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

local ARGV_T = cjson.decode(record)
for i, v in pairs(ARGV_T) do
    v[2] = 0 - v[2]
    v[3] = redis.call('incrby', KEYS[1] .. ":" .. v[1], v[2])
end

redis.call('hset', KEYS[3], ARGV[1], cjson.encode(ARGV_T))
redis.call('hdel', KEYS[2], ARGV[1])
return ARGV_T
`)

var PointEdit = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* KEYS[2] EditKey
* ARGV[1] 积分记录ID
* ARGV[2] 积分编号
* ARGV[3] 增减数量
* result 增减后总数量
*/]]
local quantity = tonumber(ARGV[3])
if quantity == 0 then
	return redis.error_reply("数量不能为0")
end

if quantity < 0 then
	local val = redis.call('get', KEYS[1] .. ":" .. ARGV[2])
	local total = 0
	if val then
		total = tonumber(val)
	end
	if total + quantity < 0 then
		return redis.error_reply('not enough:' .. ARGV[2] .. ',' .. total)
	end
end

if (redis.call('hsetnx', KEYS[2], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" ..  ARGV[1])
end

local result = redis.call('incrby', KEYS[1] .. ":" .. ARGV[2], quantity)
redis.call('hset', KEYS[2], ARGV[1], cjson.encode({ARGV[2], quantity, result}))
return result
`)

var PointRevoke = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* KEYS[2] EditKey
* KEYS[3] EditRevokeKey
* ARGV[1] 积分记录ID
* result  撤销后积分
*/]]
local record = redis.call('hget', KEYS[2], ARGV[1])
if not record then
    return redis.error_reply("not exists:" .. ARGV[1])
end

local ARGV_T = cjson.decode(record)
local val = redis.call('get', KEYS[1] .. ":" .. ARGV_T[1])
local total = 0
if val then
	total = tonumber(val)
end
if total - ARGV_T[2] < 0 then
	return redis.error_reply('not enough:' .. ARGV_T[1] .. ',' .. total)
end
ARGV_T[2] = 0 - ARGV_T[2]

if (redis.call('hsetnx', KEYS[3], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" ..  ARGV[1])
end

ARGV_T[3] = redis.call('incrby', KEYS[1] .. ":" .. ARGV_T[1], ARGV_T[2])

redis.call('hset', KEYS[3], ARGV[1], cjson.encode(ARGV_T))
return ARGV_T[3]
`)

var PointApply = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* KEYS[2] ApplyKey
* ARGV[1] 订单号
* ARGV[2] 积分编号
* ARGV[3] 增减数量
* ARGV[4] data
* result 增减后积分数量
*/]]
local quantity = tonumber(ARGV[3])
if quantity == 0 then
	return redis.error_reply("数量不能为0")
end
if quantity >= 0 then
	return redis.error_reply("数量必须小于0")
end
local val = redis.call('get', KEYS[1] .. ":" .. ARGV[2])
local total = 0
if val then
	total = tonumber(val)
end
if total + quantity < 0 then
	return redis.error_reply('not enough:' .. ARGV[2] .. ',' .. total)
end

if (redis.call('hsetnx', KEYS[2], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

local result = redis.call('incrby', KEYS[1] .. ":" .. ARGV[2], quantity)
redis.call('hset', KEYS[2], ARGV[1], cjson.encode({ARGV[2], quantity, result, ARGV[4]}))
return result
`)

var PointApplyConfirm = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* KEYS[2] ApplyKey
* KEYS[3] ApplyResultKey
* KEYS[4] EditKey
* ARGV[1] 订单号
* ARGV[2] 积分记录ID
* result 增减后积分数量
*/]]
local record = redis.call('hget', KEYS[2], ARGV[1])
if not record then
    return redis.error_reply('not exists:' .. ARGV[1])
end

if (redis.call('hsetnx', KEYS[3], ARGV[1], ARGV[2]) == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

local ARGV_T = cjson.decode(record)
ARGV_T[3] = tonumber(redis.call('get', KEYS[1] .. ":" .. ARGV_T[1]))
local data = cjson.decode(ARGV_T[4])

redis.call('hdel', KEYS[2], ARGV[1])
redis.call('hset', KEYS[4], ARGV[2], cjson.encode(ARGV_T))
data["validTotal"] = ARGV_T[3]
return cjson.encode(data)
`)

var PointApplyCancel = redis.NewScript(`
--[[/*
* KEYS[1] 积分Key
* KEYS[2] ApplyKey
* KEYS[3] ApplyResultKey
* ARGV[1] 订单号
* result 增减后积分数量
*/]]
local record = redis.call('hget', KEYS[2], ARGV[1])
if not record then
    return redis.error_reply('not exists:' .. ARGV[1])
end

if (redis.call('hsetnx', KEYS[3], ARGV[1], '0') == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

local ARGV_T = cjson.decode(record)
ARGV_T[2] = 0 - ARGV_T[2]
ARGV_T[3] = redis.call('incrby', KEYS[1] .. ":" .. ARGV_T[1], ARGV_T[2])

redis.call('hdel', KEYS[2], ARGV[1])
return ARGV_T[3]
`)
