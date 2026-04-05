CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                     email VARCHAR(255) NOT NULL UNIQUE,
                                     password_hash VARCHAR(255) NOT NULL,
                                     first_name VARCHAR(100),
                                     last_name VARCHAR(100),
                                     role VARCHAR(50) NOT NULL DEFAULT 'user',
                                     is_active BOOLEAN DEFAULT TRUE,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                     updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);

COMMENT ON TABLE users IS 'Platform users';
COMMENT ON COLUMN users.role IS 'user, admin';
COMMENT ON COLUMN users.password_hash IS 'bcrypt hashed password';