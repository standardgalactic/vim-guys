local ws = require("vim-guys.socket")

local M = {}
local VERSION = 1

---@enum Types
local Types = {
    Authentication = 1,
}

setmetatable(Types, {
    __index = Types,
    --- @param t Types
    to_string = function(t)
        if t == Types.Authentication then
            return "Authentication"
        end
        return "Unknown"
    end,
    __newindex = function() error("Cannot modify enum Colors") end,
})

--- @param token string
local function authentication(token)
end

return {
    Types = Types,
    authentication = authentication,
}

