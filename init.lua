box.cfg{
    listen = 3301,
    log_level = 5,
    memtx_memory = 512 * 1024 * 1024
}

box.once("init", function()
    -- Создаём пользователя, если его нет
    if not box.schema.user.exists('admin') then
        box.schema.user.create('admin', { password = 'secret' })
    end
    box.schema.user.grant('admin', 'super', nil, nil, { if_not_exists = true })

    -- Создаём пространство votes, если его нет
    local votes = box.schema.space.create('votes', { if_not_exists = true })

    -- Исправляем тип индекса на unsigned
    votes:create_index('primary', { type = 'hash', parts = {1, 'string'}, if_not_exists = true })
end)

print("Tarantool started and configured")

-- Не создаём пространство снова, просто получаем ссылку на него
local votes = box.space.votes

print("Tarantool started and votes space initialized!")
