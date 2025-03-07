
local M = {}

--- @param num number[] | string
function M.to_hex_string(numsOrString)
    local str = "0x"
    if type(numsOrString) == "table" then
        for _, n in ipairs(num) do
            str = str .. string.format("%02x", n)
        end
    elseif type(numsOrString) == "string" then
        print("printing string", #numsOrString)
        for i = 1, #numsOrString do
            local part = string.format("%02x", numsOrString:byte(i, i))
            print("    part:", part)
            str = str .. part
        end
        print("done", #str)
    end

    return str
end

return M



