local eq = assert.are.same
local frame = require("vim-guys.socket.frame")

describe("frame", function()
    it("testing the big endian translations!", function()
        local test = "EA"
        local to = frame.big_endian_16_to_num(test)
        local from = frame.to_big_endian_16(to)
        eq(test, from)
    end)

    it("protocol frame", function()
        local p = frame.ProtocolFrame:new(frame.Type.Authentication, "1234-1234")
        local frame_str = p:to_frame()
        local out = frame.ProtocolFrame:from_string(frame_str)

        eq(p.data, out.data)
        eq(p.len, out.len)
        eq(p.type, out.type)
    end)
end)


