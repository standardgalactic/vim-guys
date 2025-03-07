local reload = require("vim-guys.reload")
reload.reload_all()

local ws = require("vim-guys.socket")
local frame = require("vim-guys.socket.frame")
local test_utils = require("vim-guys.test_utils")

--- @type WS | nil
Client = Client or nil

if Client ~= nil then
    pcall(Client.close, Client)
end

Client = ws.connect("localhost", 42000)
Client:on_status_change(function (s)
    print("status change", s)
    if s == "connected" then
        local auth = frame.authentication("07669e6d-2857-486a-8208-ce64172875f7")
        print("sending message", test_utils.to_hex_string(auth))
        Client:msg(auth)
    end
end)
Client:on_action(function(s)
    print("message received", test_utils.to_hex_string(s))
end)
