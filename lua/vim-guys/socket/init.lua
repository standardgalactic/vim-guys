local bit = require("bit")
local uv = vim.loop
__Prevent_reconnect = true

--- @alias StatusChange "connected" | "disconnected"
--- @alias StatusChangeCB fun(s: StatusChange)
--- @alias ServerStatusChangeCB fun()
--- @alias ServerMsgCB fun(frame: string)

local PING = 0x9
local PONG = 0xA
local TEXT = 0x1
local BINARY = 0x2
local mask = {0x45, 0x45, 0x45, 0x45}

---@param ping WSFrame
---@return string
local function create_pong_frame(ping)
    local fin = 0x80
    local opcode = PONG
    local mask_bit = 0x00

    local payload_length = #ping.data
    local length_field = ""
    if payload_length <= 125 then
        length_field = string.char(payload_length)
    else
        error("unable to create pong with a large amount of data")
    end

    return string.char(fin + opcode) .. mask_bit .. length_field .. ping
end


local function complete_header(header)
    -- Byte 2: Mask bit and initial payload length
    local second_byte = string.byte(header, 2)
    local payload_length = bit.band(second_byte, 0x7F) -- Lower 7 bits of Byte 2
    local required_len = 2
    if payload_length == 126 then
        required_len = required_len + 2
    elseif payload_length == 127 then
        return 0, true
    end

    local mask_bit = bit.band(second_byte, 0x80) ~= 0 -- Mask bit (MSB of Byte 2)
    if mask_bit then
        required_len = required_len + 4
    end

    return required_len <= #header, nil
end

--- @class WSFrame
--- @field data string
--- @field opcode number
--- @field _len number
--- @field _mask string
--- @field _state string
--- @field _buf string
--- @field _fin boolean
--- @field errored boolean
local WSFrame = {}
WSFrame.__index = WSFrame

function WSFrame:new()
    return setmetatable({
        data = "",
        errored = false,
        _state = "init",
        _buf = "",
    }, self)
end

--- @param data string
--- @return string
function WSFrame:_parse_header(data)
    local first_byte = string.byte(data, 1)
    local second_byte = string.byte(data, 2)
    local payload_length = bit.band(second_byte, 0x7F) -- Lower 7 bits of Byte 2
    local opcode = bit.band(first_byte, 0x0F)          -- Lower 4 bits of Byte 1

    local offset = 2
    if payload_length == 126 then
        payload_length = bit.lshift(string.byte(data, 3), 8) + string.byte(data, 4)
        offset = offset + 2
    end

    local mask_bit = bit.band(second_byte, 0x80) ~= 0 -- Mask bit (MSB of Byte 2)
    local mask = ""
    if mask_bit then
        mask = data:sub(offset + 1, offset + 1 + 4)
        offset = offset + 4
    end

    self._fin = bit.band(first_byte, 0x80) ~= 0
    self._len = payload_length
    self._mask = mask
    self.opcode = opcode

    return data:sub(offset + 1)
end

function WSFrame:done()
    return self._state == "done"
end

function WSFrame:mask()
    return self._mask
end

--- @param txt string
--- @return string
function WSFrame.text_frame(txt)
    local payload_length = #txt

    local first_byte = bit.band(BINARY, 0x0F)
    first_byte = bit.bor(first_byte, 0x80)

    local out_string = string.char(first_byte)

    if payload_length > 125 then
        local len = 126
        len = bit.bor(len, 0x80)
        out_string = out_string .. string.char(len)

        local high_byte = bit.rshift(payload_length, 8)
        local low_byte = bit.band(payload_length, 0xFF)
        out_string = out_string .. string.char(high_byte)
        out_string = out_string .. string.char(low_byte)
    else
        local len = payload_length
        len = bit.bor(len, 0x80)
        out_string = out_string .. string.char(len)
    end

    for i = 1, #mask do
        out_string = out_string .. string.char(mask[i])
    end

    local out_txt = ""
    for i = 1, #txt do
        local mask_byte = mask[i % 4 + 1]
        out_txt = out_txt .. string.char(bit.bxor(txt:byte(i, i), mask_byte))
    end

    out_string = out_string .. out_txt
    return out_string
end

--- @param data string
--- @return string
function WSFrame:push(data)
    data = self._buf .. data
    self._buf = ""

    while self._state ~= "done" and #data > 0 do
        if self._state == "init" then
            local finished, err = complete_header(data)
            if err ~= nil then
                self.errored = true
            end

            if not finished then
                self._buf = self._buf .. data
                return ""
            end

            data = self:_parse_header(data)
            self._state = "body"
        elseif self._state == "body" then
            if self._len <= #data then
                self.data = data:sub(1, self._len)
                data = data:sub(self._len + 1)
                if self._fin then
                    self._state = "done"
                else
                    error("I haven't programmed in multiframe frames")
                end
            end
        end
    end

    return data
end

--- @class TCPSocket
--- @field close fun(self: TCPSocket)
--- @field connect fun(self: TCPSocket, addr: string, port: number, cb: fun(e: unknown))
--- @field is_closing fun(self: TCPSocket): boolean
--- @field write fun(self: TCPSocket, msg: string)
--- @field read_start fun(self: TCPSocket, cb: fun(err: unknown, data: string))

--- @class WS
--- @field host string
--- @field port number
--- @field status StatusChange
--- @field _upgraded boolean
--- @field _running boolean
--- @field _client nil | TCPSocket i don't know what this is suppose to be so i made my own
--- @field _currentFrame nil | WSFrame i don't know what this is suppose to be so i made my own
--- @field _on_messages table<ServerActionCB>
--- @field _on_status_change table<StatusChangeCB>
--- @field _queued_messages table<string>
local WS = {}
WS.__index = WS

---@param host string
---@param port number
---@return WS
function WS:new(host, port)
    return setmetatable({
        host = host,
        port = port,
        status = "disconnected",
        _upgraded = false,
        _currentFrame = WSFrame:new(),
        _client = nil,
        _running = false,
        _on_messages = {},
        _on_status_change = {},
        _queued_messages = {},
    }, self)
end

--- @param status StatusChange
function WS:_status_change(status)
    self.status = status
    for _, cb in ipairs(self._on_status_change) do
        cb(status)
    end
end

--- @param txt string
function WS:msg(txt)
    if self._upgraded == false then
        table.insert(self._queued_messages, txt)
        return
    end
    self._client:write(WSFrame.text_frame(txt))
end

function WS:_flush()
    if self._upgraded == false then
        error("called _flush but i am not upgraded... wtf")
    end

    for _, value in ipairs(self._queued_messages) do
        self._client:write(WSFrame.text_frame(value))
    end
    self._queued_messages = {}
end

function WS:close()
    self._running = false
    if self._client ~= nil then
        self._client:close()
        self:_status_change("disconnected")
    end
end

function WS:_connect()
    self._running = true

    -- Handle cleanup when the Neovim process exits
    vim.api.nvim_create_autocmd("VimLeavePre", {
        callback = function()
            if self._client ~= nil and not self._client:is_closing() then
                self._client:close()
                self:_status_change("disconnected")
            end
        end,
    })

    self:_run_connect()
end

function WS:_reconnect()
    self._upgraded = false
    if not self._running then
        return
    end

    if self._client ~= nil then
        self._client:close()
    end

    print("server connection broken... reconnecting in 5 seconds")
    vim.defer_fn(function()
        if __Prevent_reconnect then
            return
        end

        self._client = nil
        self:_run_connect()
    end, 50)
end

--- @param msg WSFrame
function WS:_emit(msg)
    if msg.opcode == PING then
        local out = create_pong_frame(msg)
        self._client:write(out)
    elseif msg.opcode == BINARY then
        local mask = msg:mask()
        if #mask > 0 then
            error("i haven't implemented mask... wtf server")
        end

        for _, cb in ipairs(self._on_messages) do
            cb(msg.data)
        end
    else
        error("i don't currently support this frame.. what is this? " .. vim.inspect(msg))
    end
end

function WS:_run()

    local ws_upgrade = {
        "GET /socket HTTP/1.1",
        string.format("Host: %s:%d", self.host, self.port),
        string.format("origin: http://%s:%d", self.host, self.port),
        "Upgrade: websocket",
        "Connection: Upgrade",
        "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==",
        "Sec-WebSocket-Version: 13",
        "",
        ""
    }

    -- Send a message to the server
    self._client:write(table.concat(ws_upgrade, "\r\n"))

    -- Read data from the server
    self._client:read_start(function(err, data)
        if err then
            print("Read error:", err)
            return
        end

        if data then
            if self._upgraded then
                local remaining = data

                while #remaining > 0 do
                    remaining = self._currentFrame:push(remaining)
                    if self._currentFrame.errored then
                        error("OH NO")
                        self:_reconnect()
                        return
                    end
                    if self._currentFrame:done() then
                        self:_emit(self._currentFrame)
                        self._currentFrame = WSFrame:new()
                    end
                end
            else
                local str = "HTTP/1.1 101 Switching Protocols"
                local sub = data:sub(1, #str)
                self._upgraded = str == sub
                if self._upgraded then
                    self:_status_change("connected")
                    self:_flush()
                else
                    print("upgrade protocol was unsuccessful...")
                    self:_reconnect()
                end
            end
        else
            self:_reconnect()
        end
    end)
end

function WS:_run_connect()
    -- Resolve the hostname
    uv.getaddrinfo(self.host, nil, {}, function(err, res)
        if err then
            error("DNS resolution error for WS:", err)
            return
        end

        -- Extract the first resolved address
        local resolved_address = res[1].addr
        self._client = uv.new_tcp()
        print("resolved addr", resolved_address)

        -- Connect to the resolved address and port
        self._client:connect(resolved_address, self.port, function(e)
            if not e then
                self:_run()
            else
                print("error occurred", e)
                self:_reconnect()
            end
        end)
    end)
end

--- @param key string
function WS:authorize(key)
    -- From golang def... probably need to look into protobuf
    -- all comes are in a type box as well with the type of message
    --
    -- type SocketAuth struct {
    --     Type SocketAuthType
    --     Key  string
    -- }

    local data = vim.json.encode({
        type = "socket-auth",
        data = {
            type = "admin",
            key = key,
        },
    })

    self:msg(data)
end


--- @param cb StatusChangeCB
function WS:on_status_change(cb)
    table.insert(self._on_status_change, cb)
    cb(self.status)
end

--- @param cb ServerMsgCB
function WS:on_action(cb)
    table.insert(self._on_messages, cb)
end

---@param host string
---@param port number
---@return WS
local connect = function(host, port)
    local ws = WS:new(host, port)
    ws:_connect()
    return ws
end

return {
    connect = connect,
    WSFrame = WSFrame,
}


