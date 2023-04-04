local tokens_key = KEYS[1]
local timestamp_key = KEYS[2]

local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local last_tokens = tonumber(redis.call("get", tokens_key))
if last_tokens == nil then
    last_tokens = capacity
end

local last_refreshed = tonumber(redis.call("get", timestamp_key))
if last_refreshed == nil then
    last_refreshed = 0
end

local delta = math.max(0, now-last_refreshed)
local filled_tokens = math.min(capacity, last_tokens+(delta*rate))

local New_tokens = 0
local allowed = filled_tokens >= requested
if allowed then
    New_tokens = filled_tokens - requested
end

-- This part is optional. It has nothing to do with the actual algorithm.
-- Putting an expiry on a key is good to improve the space efficiency of the cache
-- This clear the key if there is no request from this user
local ttl = 20
redis.call("setex", tokens_key, ttl, New_tokens)
redis.call("setex", timestamp_key, ttl, now)

return { allowed }
