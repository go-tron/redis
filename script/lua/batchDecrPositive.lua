--[[/*
* KEYS[1] 库存Key
* KEYS[2] 订单Key
* ARGV[1] 订单号
* ARGV[2] 数组 v[1]:库存编号 v[2]:出库数量
* result 数组 v[1]:库存编号 v[2]:出库数量 v[3]:库存数量
*/]]

local ARGV_T = cjson.decode(ARGV[2])
for i, v in pairs(ARGV_T) do
    if v[2] <= 0 then
        return redis.error_reply("number invalid")
    end
    local curr = redis.call('hget', KEYS[1], v[1])
    if not curr then
        return v[1]
    end
    if tonumber(curr) - tonumber(v[2]) < 0 then
        return v[1]
    end
end

if (redis.call('hsetnx', KEYS[2], ARGV[1], ARGV[2]) == 0) then
    return -1
end

for i, v in pairs(ARGV_T) do
    v[3] = redis.call('hincrby', KEYS[1], v[1], 0 - v[2])
end

return ARGV_T