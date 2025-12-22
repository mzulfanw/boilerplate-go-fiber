-- Seed system admin user
INSERT INTO users (email, password_hash, is_active)
VALUES ('admin@boiler.com', crypt('abcd5dasar', gen_salt('bf', 12)), true)
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.name = 'admin'
WHERE u.email = 'admin@boiler.com'
ON CONFLICT DO NOTHING;
