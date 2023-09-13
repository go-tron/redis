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