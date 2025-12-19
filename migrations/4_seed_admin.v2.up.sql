-- Сид администратора через PL/pgSQL + проверка наличия
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@example.com') THEN
        INSERT INTO users (username, email, password, country, role, registration_date, is_blocked)
        VALUES (
                'Admin',
                'admin@example.com',
                -- bcrypt hash для "admin123"
                '$2a$10$X9hfXSY.VS1APYp4nAIGkuVCehu6mne6tClYI1MLC8ccI/Bs6HAwO',
                'RU',
                'admin',
                now(),
                FALSE
            );
    END IF;
END $$;
