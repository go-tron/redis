--[[/*
* KEYS[1] 订单key
* KEYS[2] 名额key
* ARGV[1] 订单号
* ARGV[2] 数组 v[1]:key v[2]:field v[3]:数量 v[4]:总数量
* result  数组 v[1]:key v[2]:field v[3]:数量 v[4]:总数量 v[5]:当前数量
*/]]
local ARGV_T = cjson.decode(ARGV[2])
for i, v in pairs(ARGV_T) do
    local curr = redis.pcall('hget', KEYS[2] .. ':' .. v[1], v[2])
    local currNumber = 0
    if curr then
        currNumber = tonumber(curr)
    end
    if currNumber + v[3] > v[4] then
        return redis.error_reply('reach limit:' .. v[1] .. ',' .. v[2])
    end
end

if (redis.call('hsetnx', KEYS[1], ARGV[1], ARGV[2]) == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

for i, v in pairs(ARGV_T) do
    local curr = redis.call('hincrbyfloat', KEYS[2] .. ':' .. v[1], v[2], v[3])
    v[5] = tonumber(curr)
end

return cjson.encode(ARGV_T)