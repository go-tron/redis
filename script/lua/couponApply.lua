--[[/*
* KEYS[1] 订单key
* KEYS[2] 优惠券key
* ARGV[1] 订单号
* ARGV[2] 优惠券ID数组
*/]]
local ARGV_T = cjson.decode(ARGV[2])
for i, v in pairs(ARGV_T) do
    local curr = redis.pcall('hget', KEYS[2], v)
    if curr then
        return redis.error_reply('already applied:' .. v)
    end
end

if (redis.call('hsetnx', KEYS[1], ARGV[1], ARGV[2]) == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

for i, v in pairs(ARGV_T) do
    redis.call('hset', KEYS[2], v, ARGV[1])
end

return ARGV[1]