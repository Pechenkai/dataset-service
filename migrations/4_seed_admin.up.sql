DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@example.com') THEN
        INSERT INTO users (username, email, password, country, role, registration_date, is_blocked)
        VALUES (
                'Admin',
                'admin@example.com',
                -- bcrypt hash for "admin123" (cost=10)
                '$2a$10$X9hfXSY.VS1APYp4nAIGkuVCehu6mne6tClYI1MLC8ccI/Bs6HAwO',
                'RU',
                'admin',
                now(),
                FALSE
            );
    END IF;
END $$;
