-- Remove RBAC management permissions
DELETE FROM role_permissions
WHERE permission_id IN (
  SELECT id
  FROM permissions
  WHERE name IN (
    'role.read',
    'role.create',
    'role.update',
    'role.delete',
    'permission.read',
    'permission.create',
    'permission.update',
    'permission.delete',
    'role.permission.read',
    'role.permission.update'
  )
);

DELETE FROM permissions
WHERE name IN (
  'role.read',
  'role.create',
  'role.update',
  'role.delete',
  'permission.read',
  'permission.create',
  'permission.update',
  'permission.delete',
  'role.permission.read',
  'role.permission.update'
);
