--[[/*
* KEYS[1] 订单key
* KEYS[2] 名额key
* ARGV[1] 订单号
* ARGV[2] 数组 v[1]:key v[2]:field v[3]:数量 v[4]:总数量
* ARGV[3] 随机种子
* result  数组 v[1]:key v[2]:field v[3]:数量 v[4]:总数量 v[5]:当前数量
*/]]

local ARGV_2 = cjson.decode(ARGV[2])
local a = {}
local total = 0
for i, v in pairs(ARGV_2) do
    local curr = redis.pcall('hget', KEYS[2] .. ':' .. v[1], v[2])
    local currNumber = 0
    if curr then
        currNumber = tonumber(curr)
    end
    if currNumber + v[3] <= v[4] then
        total = total + v[4]
        table.insert(a, v)
    end
end

local s = {}
if #a == 0 then
    return redis.error_reply('none available')
elseif #a == 1 then
    s = a[1]
else
    math.randomseed(ARGV[3])
    local w = math.random(total)
    --    redis.log(redis.LOG_NOTICE, "w", w)

    for i, v in pairs(a) do
        if w <= v[4] then
            s = v
            break
        else
            w = w - v[4]
        end
    end
    --    redis.log(redis.LOG_NOTICE, "s", s[1])
end

if (redis.call('hsetnx', KEYS[1], ARGV[1], cjson.encode(s)) == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

local curr = redis.call('hincrby', KEYS[2] .. ':' .. s[1], s[2], s[3])
s[5] = tonumber(curr)

return cjson.encode(s)