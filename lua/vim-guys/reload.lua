local M = {}

local vim_guys = "vim-guys"
M.reload_all = function()
	for module_name in pairs(package.loaded) do
		if module_name:sub(1, #vim_guys) == vim_guys then
			package.loaded[module_name] = nil
		end
	end
end

return M


