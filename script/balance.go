package script

import "github.com/redis/go-redis/v9"

var BalanceGet = redis.NewScript(`
--[[/*
* KEYS[1] 余额Key
* ARGV[1] 余额编号
*/]]
local val = redis.call('hget', KEYS[1], ARGV[1])
local total = 0
if val then 
	total = val
end
return total
`)

var BalanceList = redis.NewScript(`
--[[/*
* KEYS[1] 余额Key
* ARGV[1] 余额编号
* ARGV[2] List[n]
*/]]
local ARGV2 = cjson.decode(ARGV[2])
local result = {}
for i, v in pairs(ARGV2) do
    local val = redis.call('hget', KEYS[1]..v, ARGV[1])
	local total = '0'
    if val then
        total = val
    end
    result[i] = {v, total}
end
return result
`)

var BalanceConsume = redis.NewScript(`
--[[/*
* KEYS[1] 余额Key
* KEYS[2] ConsumeKey
* ARGV[1] 订单号
* ARGV[2] 余额数组 v[1]:余额编号 v[2]:使用数量
* result 使用后余额
*/]]
local ARGV2 = cjson.decode(ARGV[2])
local n = tonumber(ARGV2[2])
if n <= 0 then
	return redis.error_reply("number must be positive")
end
local val = redis.call('hget', KEYS[1], ARGV2[1])
local total = 0
if val then
	total = tonumber(val)
end
if total - n < 0 then
	return redis.error_reply('not enough:' .. ARGV2[1] .. ',' .. total)
end

if (redis.call('hsetnx', KEYS[2], ARGV[1], ARGV[2]) == 0) then
    return redis.error_reply("exists:" ..  ARGV[1])
end

return redis.call('hincrbyfloat', KEYS[1], ARGV2[1], 0 - ARGV2[2])
`)

var BalanceConsumeRevoke = redis.NewScript(`
--[[/*
* KEYS[1] 余额Key
* KEYS[2] ConsumeKey
* KEYS[3] RevokeKey
* ARGV[1] 订单号
* result  v[1]:余额编号 v[2]:充值数量 v[3]:撤销后余额
*/]]
local record = redis.call('hget', KEYS[2], ARGV[1])
if not record then
    return redis.error_reply("not exists")
end

if (redis.call('hsetnx', KEYS[3], ARGV[1], record) == 0) then
    return redis.error_reply("exists:" ..  ARGV[1])
end

local ARGV2 = cjson.decode(record)
ARGV2[3] = tonumber(redis.call('hincrbyfloat', KEYS[1], ARGV2[1], ARGV2[2]))
return cjson.encode(ARGV2)
`)

var BalanceCharge = redis.NewScript(`
--[[/*
* KEYS[1] 余额Key
* KEYS[2] ChargeKey
* ARGV[1] 订单号
* ARGV[2] v[1]:余额编号 v[2]:充值数量
* result 充值后余额
*/]]
local ARGV2 = cjson.decode(ARGV[2])
local n = tonumber(ARGV2[2])
if n <= 0 then
	return redis.error_reply("number must be positive")
end

if (redis.call('hsetnx', KEYS[2], ARGV[1], ARGV[2]) == 0) then
    return redis.error_reply("exists:" ..  ARGV[1])
end

return redis.call('hincrbyfloat', KEYS[1], ARGV2[1], ARGV2[2])
`)

var BalanceChargeRevoke = redis.NewScript(`
--[[/*
* KEYS[1] 余额Key
* KEYS[2] ChargeKey
* KEYS[3] RevokeKey
* ARGV[1] 订单号
* result  v[1]:余额编号 v[2]:充值数量 v[3]:撤销后余额
*/]]
local record = redis.call('hget', KEYS[2], ARGV[1])
if not record then
    return redis.error_reply("not exists")
end

if (redis.call('hsetnx', KEYS[3], ARGV[1], record) == 0) then
    return redis.error_reply("exists:" ..  ARGV[1])
end

local ARGV2 = cjson.decode(record)
ARGV2[3] = tonumber(redis.call('hincrbyfloat', KEYS[1], ARGV2[1], 0 - ARGV2[2]))
return cjson.encode(ARGV2)
`)
