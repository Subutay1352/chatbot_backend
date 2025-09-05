-- Chatbot Backend Database Setup
-- PostgreSQL için manuel tablo oluşturma script'i

-- Veritabanını oluştur (eğer yoksa)
-- CREATE DATABASE chatbot;

-- Veritabanına bağlan
-- \c chatbot;

-- 1. Sessions Tablosu
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    is_favorite BOOLEAN DEFAULT FALSE
);

-- 2. Messages Tablosu
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(255) PRIMARY KEY,
    content TEXT NOT NULL,
    sender VARCHAR(50) NOT NULL CHECK (sender IN ('user', 'bot')),
    timestamp TIMESTAMP NOT NULL,
    message_type VARCHAR(50) DEFAULT 'text',
    is_typing BOOLEAN DEFAULT FALSE,
    is_favorite BOOLEAN DEFAULT FALSE,
    is_regenerated BOOLEAN DEFAULT FALSE,
    original_message_id VARCHAR(255),
    session_id VARCHAR(255) NOT NULL,
    language VARCHAR(10),
    code_block BOOLEAN DEFAULT FALSE,
    link_title VARCHAR(255),
    link_description TEXT,
    link_image VARCHAR(500),
    link_url VARCHAR(500),
    link_domain VARCHAR(255),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- 3. Reactions Tablosu
CREATE TABLE IF NOT EXISTS reactions (
    id VARCHAR(255) PRIMARY KEY,
    emoji VARCHAR(10) NOT NULL,
    count INTEGER DEFAULT 0,
    users TEXT, -- JSON array as string
    message_id VARCHAR(255) NOT NULL,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

-- 4. Performans için İndeksler
CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender);
CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at);
CREATE INDEX IF NOT EXISTS idx_sessions_is_favorite ON sessions(is_favorite);
CREATE INDEX IF NOT EXISTS idx_reactions_message_id ON reactions(message_id);

-- 5. Örnek veri ekleme (isteğe bağlı)
-- INSERT INTO sessions (id, title, created_at, updated_at, is_favorite) 
-- VALUES ('demo-session-1', 'Demo Chat', NOW(), NOW(), false);

-- INSERT INTO messages (id, content, sender, timestamp, session_id) 
-- VALUES 
-- ('demo-msg-1', 'Merhaba!', 'user', NOW(), 'demo-session-1'),
-- ('demo-msg-2', 'Merhaba! Size nasıl yardımcı olabilirim?', 'bot', NOW(), 'demo-session-1');

-- Tabloları kontrol et
SELECT table_name, column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_schema = 'public' 
ORDER BY table_name, ordinal_position;
