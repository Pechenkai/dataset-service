-- Возврат старого хэша (если нужно откатить)
UPDATE users
SET password = '$2a$10$X9hfXSY.VS1APYp4nAIGkuVCehu6mne6tClYI1MLC8ccI/Bs6HAwO'
WHERE email = 'admin@example.com';
