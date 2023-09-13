package script

import "github.com/redis/go-redis/v9"

var InventoryGet = redis.NewScript(`
--[[/*
* KEYS[1] 库存Key
* ARGV[1] 商品编号
* ARGV[2] sku编号
*/]]
local val = redis.call('hget', KEYS[1] .. ":" .. ARGV[1], ARGV[2])
local total = 0
if val then
    total = tonumber(val)
end
return total
`)

var InventoryList = redis.NewScript(`
--[[/*
* KEYS[1] 库存Key
* ARGV[1] 数组 v[1]:商品编号 v[2]:sku编号
* result 数组 v[1]:商品编号 v[2]:sku编号 v[3]:库存数量
*/]]
local ARGV_T = cjson.decode(ARGV[1])
for i, v in pairs(ARGV_T) do
    local val = redis.call('hget', KEYS[1] .. ":" .. v[1], v[2])
    local total = 0
    if val then
        total = tonumber(val)
    end
    v[3] = total
end
return ARGV_T
`)

var InventoryApply = redis.NewScript(`
--[[/*
* KEYS[1] 库存Key
* KEYS[2] ApplyKey
* ARGV[1] 订单号
* ARGV[2] 库存数组 v[1]:商品编号 v[2]:sku编号 v[3]:增减数量
* result 库存数组 v[1]:商品编号 v[2]:sku编号 v[3]:增减数量 v[4]:增减后库存数量
*/]]
local ARGV_T = cjson.decode(ARGV[2])
for i, v in pairs(ARGV_T) do
    if v[3] >= 0 then
        return redis.error_reply("number must be negative")
    end
    local val = redis.call('hget', KEYS[1] .. ":" .. v[1], v[2])
    local total = 0
    if val then
        total = tonumber(val)
    end
    if total + v[3] < 0 then
        return redis.error_reply('not enough:' .. v[2] .. ',' .. total)
    end
end

if (redis.call('hsetnx', KEYS[2], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

for i, v in pairs(ARGV_T) do
    redis.call('hincrby', KEYS[1] .. '-freeze' .. ":" .. v[1], "T", 0 - v[3])
    redis.call('hincrby', KEYS[1] .. '-freeze' .. ":" .. v[1], v[2], 0 - v[3])
    redis.call('hincrby', KEYS[1] .. ":" .. v[1], "T", v[3])
    v[4] = redis.call('hincrby', KEYS[1] .. ":" .. v[1], v[2], v[3])
end
redis.call('hset', KEYS[2], ARGV[1], cjson.encode(ARGV_T))
return ARGV_T
`)

var InventoryApplyConfirm = redis.NewScript(`
--[[/*
* KEYS[1] 库存Key
* KEYS[2] ApplyKey
* KEYS[3] ApplyResultKey
* ARGV[1] 订单号
* result v[1]:商品编号 v[2]:sku编号 v[3]:增减数量 v[4]:增减后库存数量
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
    redis.call('hincrby', KEYS[1] .. '-freeze' .. ":" .. v[1], "T", v[3])
    redis.call('hincrby', KEYS[1] .. '-freeze' .. ":" .. v[1], v[2], v[3])
    v[4] = tonumber(redis.call('hget', KEYS[1] .. ":" .. v[1], v[2]))
end

redis.call('hset', KEYS[3], ARGV[1], cjson.encode(ARGV_T))
redis.call('hdel', KEYS[2], ARGV[1])
return ARGV_T
`)

var InventoryApplyCancel = redis.NewScript(`
--[[/*
* KEYS[1] 库存Key
* KEYS[2] ApplyKey
* KEYS[3] ApplyResultKey
* ARGV[1] 订单号
* result v[1]:商品编号 v[2]:sku编号 v[3]:增减数量 v[4]:增减后库存数量
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
    redis.call('hincrby', KEYS[1] .. '-freeze' .. ":" .. v[1], "T", v[3])
    redis.call('hincrby', KEYS[1] .. '-freeze' .. ":" .. v[1], v[2], v[3])
    v[3] = 0 - v[3]
    redis.call('hincrby', KEYS[1] .. ":" .. v[1], "T", v[3])
    v[4] = redis.call('hincrby', KEYS[1] .. ":" .. v[1], v[2], v[3])
end

redis.call('hset', KEYS[3], ARGV[1], cjson.encode(ARGV_T))
redis.call('hdel', KEYS[2], ARGV[1])
return ARGV_T
`)

var InventoryApplyRevoke = redis.NewScript(`
--[[/*
* KEYS[1] 库存Key
* KEYS[2] ConfirmKey
* KEYS[3] RevokeKey
* ARGV[1] 订单号
* result 数组 v[1]:商品编号 v[2]:sku编号 v[3]:增减数量 v[4]:增减后库存数量
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
    v[3] = 0 - v[3]
    redis.call('hincrby', KEYS[1] .. ":" .. v[1], "T", v[3])
    v[4] = redis.call('hincrby', KEYS[1] .. ":" .. v[1], v[2], v[3])
end

redis.call('hset', KEYS[3], ARGV[1], cjson.encode(ARGV_T))
return ARGV_T
`)

var InventoryEditCreate = redis.NewScript(`
--[[/*
* KEYS[1] 库存Key
* KEYS[2] EditKey
* ARGV[1] 订单号
* ARGV[2] v[1]:商品编号 v[2]:sku编号 v[3]:类型 v[4]:数量 
* result 数组 v[1]:商品编号 v[2]:sku编号 v[3]:增减数量 v[4]:增减后库存数量
*/]]
local ARGV_T = cjson.decode(ARGV[2])
for i, v in pairs(ARGV_T) do
    local val = redis.call('hget', KEYS[1] .. ":" .. v[1], v[2])
    local current = 0
    if val then
        current = tonumber(val)
    end
    if v[3] == 1 or v[3] == 2 then
        if not v[4] or v[4] == 0 then
            return redis.error_reply("请输入数量")
        end
		if v[3] == 1 then
            v[3] = v[4]
        else
            v[3] = 0 - v[4]
        end
        if current + v[3] < 0 then
            return redis.error_reply("库存数量不能为负")
        end 
    elseif v[3] == 3 then
        if not v[4] then
            return redis.error_reply("请输入库存数量")
        end
        if v[4] < 0 then
            return redis.error_reply("库存数量不能为负")
        end
  		if v[4] == current then
            return redis.error_reply("编辑数量与当前库存数量相同")
        end
        v[3] = v[4] - current
    else
        return redis.error_reply("库存类型错误")
    end
end

if (redis.call('hsetnx', KEYS[2], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

for i, v in pairs(ARGV_T) do
    redis.call('hincrby', KEYS[1] .. ":" .. v[1], "T", v[3])
    v[4] = redis.call('hincrby', KEYS[1] .. ":" .. v[1], v[2], v[3])
end

redis.call('hset', KEYS[2], ARGV[1], cjson.encode(ARGV_T))
return ARGV_T
`)

var InventoryEditRevoke = redis.NewScript(`
--[[/*
* KEYS[1] 库存Key
* KEYS[2] EditKey
* KEYS[3] RevokeKey
* ARGV[1] 订单号
* result 数组 v[1]:商品编号 v[2]:sku编号 v[3]:增减数量 v[4]:增减后库存数量
*/]]
local record = redis.call('hget', KEYS[2], ARGV[1])
if not record then
    return redis.error_reply('not exists:' .. ARGV[1])
end

local ARGV_T = cjson.decode(record)
for i, v in pairs(ARGV_T) do
    local val = redis.call('hget', KEYS[1] .. ":" .. v[1], v[2])
    local current = 0
    if val then
        current = tonumber(val)
    end
    if current - v[3] < 0 then
        return redis.error_reply('not enough:' .. v[2] .. ',' .. current)
    end
    v[3] = 0 - v[3]
end

if (redis.call('hsetnx', KEYS[3], ARGV[1], '1') == 0) then
    return redis.error_reply("exists:" .. ARGV[1])
end

for i, v in pairs(ARGV_T) do
    redis.call('hincrby', KEYS[1] .. ":" .. v[1], "T", v[3])
    v[4] = redis.call('hincrby', KEYS[1] .. ":" .. v[1], v[2], v[3])
end

redis.call('hset', KEYS[3], ARGV[1], cjson.encode(ARGV_T))
return ARGV_T
`)
