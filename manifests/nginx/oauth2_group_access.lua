local jwt = require "jwt"
local validators = require "jwt-validators"
-- local inspect = require "inspect"
local cjson = require("cjson")
local _M = {}

local function tableEmpty (t)
    for _, _ in ipairs(t) do
        return false
    end
    return true
end

local function joinTable(t)
    local x = {}
    for k, v in pairs(t) do
        table.insert(x, v..', ')
    end
    local s = table.concat(x)
    return s:sub(1, #s - 2)..' '
end

function _M.verify_authorization(self, authorization, authorized_groups_list)
  if authorization == nil or authorization == '' then
    ngx.log(ngx.ERR, "empty http authorization")
  elseif authorized_groups_list == nil or authorized_groups_list == '' or tableEmpty(authorized_groups_list) then
    ngx.log(ngx.ERR, "no authorized groups")
  else
    if authorization:find("Bearer ") ~= 1 then
      ngx.status = ngx.HTTP_BAD_REQUEST
      ngx.say("invalid authorization, not Bearer")
      ngx.exit(ngx.HTTP_OK)
    end

    local _, _, token = string.find(authorization, "Bearer%s+(.+)")

    local jwt_obj = jwt:load_jwt(token)
    if not jwt_obj.valid then
      ngx.status = ngx.HTTP_BAD_REQUEST
      ngx.say("invalid jwt")
      ngx.exit(ngx.HTTP_OK)
    end

    local groups = jwt_obj.payload.groups

    if groups == nil then
      ngx.status = ngx.HTTP_BAD_REQUEST
      ngx.say("groups claim not present in access token")
      ngx.exit(ngx.HTTP_OK)
    end

    local authorized_groups_set = {}
    for i, group in ipairs(authorized_groups_list) do
      authorized_groups_set[group] = true
    end

    local found = false
    -- Parse the groups and check if they match any of our authorized groups
    for i, group in ipairs(groups) do
      if authorized_groups_set[group] == true then
        -- If we found an authorized group, say so and break the loop
        found = true
        break
      end
    end


    -- If we didn't find out group in our list, then return forbidden
    if not found then
        -- If not, throw a forbidden
        ngx.status = ngx.HTTP_FORBIDDEN
        ngx.say("User is not in required group(s) (" .. joinTable(authorized_groups_list) .. ") in: " .. joinTable(groups))
        ngx.exit(ngx.HTTP_FORBIDDEN)
    end
  end
end

return _M
