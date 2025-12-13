DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'ppo_readonly') THEN
            CREATE ROLE ppo_readonly LOGIN PASSWORD 'readonly';
        END IF;
    END
$$;

GRANT CONNECT ON DATABASE ppo TO ppo_readonly;
GRANT USAGE ON SCHEMA public TO ppo_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO ppo_readonly;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO ppo_readonly;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO ppo_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON SEQUENCES TO ppo_readonly;
