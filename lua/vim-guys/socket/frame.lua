local ws = require("vim-guys.socket")

-- PROTOCOL
-- [ version(2) | type(2) | len(2) | player_id(4) | game_id(4) | data(len) ]

local VERSION = 1
local HEADER_LENGTH = 2 + 2 + 2 + 4 + 4;
local VERSION_OFFSET = 1
local TYPE_OFFSET = 3
local LEN_OFFSET = 5
local DATA_OFFSET = HEADER_LENGTH + 1

---@enum Type
local Type = {
    Authentication = 1,
}

-- Create a reverse lookup table (optional, for efficiency with large enums)
local typeValues = {}
for _, v in pairs(Type) do
    typeValues[v] = true
end

setmetatable(Type, {
    __index = Type,
    --- @param t Type
    to_string = function(t)
        if t == Type.Authentication then
            return "Authentication"
        end
        return "Unknown"
    end,

    --- @param t number
    --- @return boolean
    is_enum = function(t)
        return typeValues[t] or false
    end,

    __newindex = function() error("Cannot modify enum Colors") end,
})

local _0_str_32 = "\0\0\0\0"
--- @param num number
--- @return string
local function to_big_endian_16(num)
    assert(num < 65536, "number bigger than 2^16")
    local highByte = math.floor(num / 256)
    local lowByte = num % 256
    return string.char(highByte, lowByte)
end

--- @param data string
--- @return number
local function big_endian_16_to_num(data)
    assert(#data == 2, "big_endian_16_to_num requires str len 2")

    local high_byte = data:sub(1, 1):byte(1, 1) * 256
    local low_byte = data:sub(2, 2):byte(1, 1)

    return high_byte + low_byte
end

--- @class ProtocolFrame
--- @field type Type
--- @field len number
--- @field data string
local ProtocolFrame = {}
ProtocolFrame.__index = ProtocolFrame

--- @param type Type
---@param data string
function ProtocolFrame:new(type, data)
    return setmetatable({
        type = type,
        len = #data,
        data = data,
    }, self)
end

--- @param data_str string
--- @return ProtocolFrame
function ProtocolFrame:from_string(data_str)
    local version = big_endian_16_to_num(data_str:sub(VERSION_OFFSET, VERSION_OFFSET + 1))
    assert(version == VERSION, "unable to communicate with server, version mismatch.  please update your client")

    local type = big_endian_16_to_num(data_str:sub(TYPE_OFFSET, TYPE_OFFSET + 1))
    local len = big_endian_16_to_num(data_str:sub(LEN_OFFSET, LEN_OFFSET + 1))
    local data = data_str:sub(DATA_OFFSET)

    assert(#data == len, "malformed packet received", "data length", #data, "len expected", len)

    return setmetatable({
        type = type,
        len = len,
        data = data,
    }, self)
end

--- @return string
function ProtocolFrame:to_frame()
    local version = to_big_endian_16(VERSION)
    local type_str = to_big_endian_16(self.type)
    local len = to_big_endian_16(self.len)

    local header = version .. type_str .. len .. _0_str_32 .. _0_str_32
    return header .. self.data
end

--- @param token string
local function authentication(token)
    local frame = ProtocolFrame:new(Type.Authentication, token):to_frame()
    print("frame produced", #frame)
    return frame
end

return {
    Type = Type,
    authentication = authentication,
    to_big_endian_16 = to_big_endian_16,
    big_endian_16_to_num = big_endian_16_to_num,
    ProtocolFrame = ProtocolFrame,
}


