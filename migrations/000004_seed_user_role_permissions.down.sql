-- Remove user role management permissions
DELETE FROM role_permissions
WHERE permission_id IN (
  SELECT id
  FROM permissions
  WHERE name IN ('user.role.read', 'user.role.update')
);

DELETE FROM permissions
WHERE name IN ('user.role.read', 'user.role.update');
