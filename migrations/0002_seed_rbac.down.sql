-- Remove sample roles and permissions
DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE name IN ('admin', 'deleter'));

DELETE FROM permissions
WHERE name IN ('user.create', 'user.read', 'user.update', 'user.delete');

DELETE FROM roles
WHERE name IN ('admin', 'deleter');
