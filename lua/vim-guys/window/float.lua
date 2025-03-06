
--- @class Float
local Float = {}
Float.__index = Float

function Float:new()
end

--- @param width number
---@param height number
---@return boolean if the recalculation is successful
function Float:recalculate(width, height)

    return true
end

return Float

