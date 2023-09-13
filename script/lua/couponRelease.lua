--[[/*
* KEYS[1] 订单key
* KEYS[2] 优惠券key
* KEYS[3] 释放key
* ARGV[1] 订单号
* result  数组 v[1]:key v[2]:field v[3]:数量 v[4]:总数量 v[5]:当前数量
*/]]
local r = redis.call('hget', KEYS[1], ARGV[1])
if not r then
    return redis.error_reply('not exists')
end
if (redis.call('hsetnx', KEYS[3], ARGV[1], r) == 0) then
    return redis.error_reply('already released')
end

local ARGV_T = cjson.decode(r)
for i, v in pairs(ARGV_T) do
    redis.call('hdel', KEYS[2], v)
end

return ARGV[1]