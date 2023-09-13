--[[/*
* KEYS[1] 库存Key
* KEYS[2] 订单Key
* ARGV[1] 订单号
* ARGV[2] 数组 库存编号
* ARGV[3] 出库数量
* result 数组 v[1]:库存编号 v[2]:出库数量 v[3]:库存数量
*/]]
local number = tonumber(ARGV[3])
if number <= 0 then
    return redis.error_reply("number invalid")
end

local result = {}
local ARGV_T = cjson.decode(ARGV[2])
for i, w in pairs(ARGV_T) do
    local curr = redis.call('hget', KEYS[1], w)
    if curr then
        local val = tonumber(curr)
        if val > 0 then
            if val >= number then
                table.insert(result, 1, { w, number })
                number = 0
                break
            else
                table.insert(result, 1, { w, val })
                number = number - val
            end
        end
    end
end

if number > 0 then
    return 0
end

if (redis.call('hsetnx', KEYS[2], ARGV[1], cjson.encode(result)) == 0) then
    return -1
end

for i, v in pairs(result) do
    v[3] = redis.call('hincrby', KEYS[1], v[1], 0 - v[2])
end

return result