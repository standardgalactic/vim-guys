local reload = require("vim-guys.reload")
reload.reload_all()

local ws = require("vim-guys.socket")
local frame = require("vim-guys.socket.frame")

--- @type WS | nil
Client = Client or nil

if Client ~= nil then
    Client:close()
end

Client = ws.connect("localhost", 42000)
Client:on_status_change(function (s)
    if s == "connected" then
    end
end)
Client:on_action(function(s)
    print("message received")
end)








































