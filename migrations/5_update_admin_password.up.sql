-- Обновление пароля дефолтного администратора на admin123
UPDATE users
SET password = '$2a$10$a.j80JB172F4rnV/xmZsFeTHVl4AqdJcSizsCMlNOzFvZM5nBESMK'
WHERE email = 'admin@example.com';
