local ping = redis.call("ping")
local key = KEYS[1]
local max_key = KEYS[2]

local request_id = ARGV[1]
local capacity = tonumber(ARGV[2])

local current_len = tonumber(redis.call("SCARD", key))
redis.call("SET", max_key, current_len)
if current_len < capacity then
    local result = redis.call("SADD", key, request_id)

    if result == 1 then 
        return true
    end
end

return false
