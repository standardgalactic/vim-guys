local M = {}

M.reload_all = function()
	for module_name in pairs(package.loaded) do
		if module_name:match("^vim-guys") then
			package.loaded[module_name] = nil
		end
	end
end

return M


