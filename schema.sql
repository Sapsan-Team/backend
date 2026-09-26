-- 1. Создание перечислений (Enums) для типов чатов и ролей
CREATE TYPE chat_type AS ENUM ('private', 'group');
CREATE TYPE member_role AS ENUM ('owner', 'admin', 'member');

-- 2. Таблица пользователей
-- Хранит профиль и хеш пароля. Авторизация идет по phone_number.
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE,
    display_name VARCHAR(100),
    bio TEXT,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. Таблица чатов (Сущность чата, независимая от его участников)
-- Для private чатов title может быть NULL.
CREATE TABLE chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type chat_type NOT NULL,
    title VARCHAR(255),
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. Таблица участников чата (Связующая таблица Many-to-Many)
CREATE TABLE chat_members (
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role member_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Оптимизация для мессенджеров: ID последнего прочитанного сообщения пользователем в этом чате.
    last_read_message_id UUID, 
    PRIMARY KEY (chat_id, user_id)
);

-- 5. Таблица сообщений
-- Хранит текст и метаданные. При удалении пользователя его сообщения остаются (ON DELETE SET NULL).
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id UUID REFERENCES users(id) ON DELETE SET NULL, 
    content TEXT,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    is_edited BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 6. Добавляем внешний ключ для last_read_message_id после создания таблицы messages
ALTER TABLE chat_members
ADD CONSTRAINT fk_last_read_message
FOREIGN KEY (last_read_message_id) REFERENCES messages(id) ON DELETE SET NULL;

-- 7. Индексы (Критически важны для масштабирования и скорости выборок)

-- Поиск пользователя по номеру телефона (для авторизации)
CREATE UNIQUE INDEX idx_users_phone ON users(phone_number);

-- Получение списка чатов конкретного пользователя
CREATE INDEX idx_chat_members_user_id ON chat_members(user_id);

-- Быстрая выборка сообщений чата с пагинацией (сортировка по дате)
CREATE INDEX idx_messages_chat_id_created_at ON messages(chat_id, created_at DESC);

-- 8. Триггерная функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Навешиваем триггеры
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_chats_updated_at BEFORE UPDATE ON chats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
