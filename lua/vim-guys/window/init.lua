
--- @class Layout
--- @field
local Layout = {}
Layout.__index = Layout

function Layout:menu()
end

function Layout:gameplay()
end

function Layout:billboard()
end

function Layout:close_billboard()
end

function Layout:display_message()
end

function Layout:close()
end

return {
    new = function()
    end
}
