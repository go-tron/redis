package script

import "github.com/redis/go-redis/v9"

var QuotaApply = redis.NewScript(`
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
    return redis.error_reply("exists:" ..  ARGV[1])
end

for i, v in pairs(ARGV_T) do
    local curr = redis.call('hincrbyfloat', KEYS[2] .. ':' .. v[1], v[2], v[3])
    v[5] = tonumber(curr)
end

return cjson.encode(ARGV_T)
`)

var QuotaRelease = redis.NewScript(`
--[[/*
* KEYS[1] 订单key
* KEYS[2] 名额key
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

if r == "" then
    return ""
end

local ARGV_T = cjson.decode(r)
for i, v in pairs(ARGV_T) do
    local curr = redis.call('hincrbyfloat', KEYS[2] .. ':' .. v[1], v[2], -v[3])
    v[5] = tonumber(curr)
end

return cjson.encode(ARGV_T)
`)

var QuotaApplyRandom = redis.NewScript(`
--[[/*
* KEYS[1] 订单key
* KEYS[2] 名额key
* ARGV[1] 订单号
* ARGV[2] 数组 v[1]:key v[2]:field v[3]:数量 v[4]:总数量
* ARGV[3] 随机数据{{id,w,q}}
* ARGV[4] 权重随机数组
* result  数组 v[1]:key v[2]:field v[3]:数量 v[4]:总数量 v[5]:当前数量
*/]]

local select = {}
if ARGV[3] then
    local ARGV_3 = cjson.decode(ARGV[3])
    local availables = {}
    local total = 0
    for i, val in pairs(ARGV_3) do
        local a = 1
        if val.q then
            for j, v in pairs(val.q) do
                local curr = redis.pcall('hget', KEYS[2] .. ':' .. v[1], v[2])
                local currNumber = 0
                if curr then
                    currNumber = tonumber(curr)
                end
                if currNumber + v[3] > v[4] then
                    a = 0
                    break
                end
            end
        end

        if a == 1 then
            total = total + val.w
            table.insert(availables, val)
        end
    end

--    redis.log(redis.LOG_NOTICE, "availables", #availables)

    if #availables == 0 then
        return redis.error_reply('none available')
    elseif #availables == 1 then
        select = availables[1]
    else
        local ARGV_4 = cjson.decode(ARGV[4])
        local w = 0
        for i, v in pairs(ARGV_4) do
           if v[1] == total then
               w = v[2]
           end
        end
--        redis.log(redis.LOG_NOTICE, "w", w)
        for i, v in pairs(availables) do
            if w <= v.w then
                select = v
                break
            else
                w = w - v.w
            end
        end
--        redis.log(redis.LOG_NOTICE, "select", select.id)
    end
end

local ARGV_2 = cjson.decode(ARGV[2])
for i, v in pairs(ARGV_2) do
    local curr = redis.pcall('hget', KEYS[2] .. ':' .. v[1], v[2])
    local currNumber = 0
    if curr then
        currNumber = tonumber(curr)
    end
    if currNumber + v[3] > v[4] then
        return redis.error_reply('reach limit:' .. v[1] .. ',' .. v[2])
    end
end

if select and select.id then
    for i, v in pairs(select.q) do
        table.insert(ARGV_2, v)
    end
    ARGV[2] = cjson.encode(ARGV_2)
end

if (redis.call('hsetnx', KEYS[1], ARGV[1], ARGV[2]) == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

for i, v in pairs(ARGV_2) do
    local curr = redis.call('hincrbyfloat', KEYS[2] .. ':' .. v[1], v[2], v[3])
    v[5] = tonumber(curr)
end

return cjson.encode(ARGV_2)
`)

var CalculateApplyRandom = redis.NewScript(`
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
`)

var CalculateApply = redis.NewScript(`
--[[/*
* KEYS[1] 订单key
* KEYS[2] 名额key
* ARGV[1] 订单号
* ARGV[2] 数组 v[1]:key v[2]:field v[3]:数量 v[4]:总数量
* result  数组 v[1]:key v[2]:field v[3]:数量 v[4]:总数量 v[5]:当前数量
*/]]
local ARGV_2 = cjson.decode(ARGV[2])
local curr = redis.pcall('hget', KEYS[2] .. ':' .. ARGV_2[1], ARGV_2[2])
local currNumber = 0
if curr then
	currNumber = tonumber(curr)
end
if currNumber + ARGV_2[3] > ARGV_2[4] then
	return redis.error_reply('reach limit:' .. ARGV_2[1] .. ',' .. ARGV_2[2])
end

if (redis.call('hsetnx', KEYS[1], ARGV[1], ARGV[2]) == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

local curr = redis.call('hincrbyfloat', KEYS[2] .. ':' .. ARGV_2[1], ARGV_2[2], ARGV_2[3])
ARGV_2[5] = tonumber(curr)

return cjson.encode(ARGV_2)
`)

var CalculateRelease = redis.NewScript(`
--[[/*
* KEYS[1] 订单key
* KEYS[2] 名额key
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

local v = cjson.decode(r)
local curr = redis.call('hincrbyfloat', KEYS[2] .. ':' .. v[1], v[2], -v[3])
v[5] = tonumber(curr)

return cjson.encode(v)
`)

var CouponApply = redis.NewScript(`
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
`)

var CouponRelease = redis.NewScript(`
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

return r
`)
