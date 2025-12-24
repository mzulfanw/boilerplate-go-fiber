DELETE FROM user_roles
WHERE user_id = (SELECT id FROM users WHERE email = 'admin@boiler.com');

DELETE FROM users
WHERE email = 'admin@boiler.com';
