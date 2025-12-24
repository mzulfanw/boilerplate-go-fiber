-- RBAC management permissions
INSERT INTO permissions (name, description)
VALUES
  ('role.read', 'Read roles'),
  ('role.create', 'Create roles'),
  ('role.update', 'Update roles'),
  ('role.delete', 'Delete roles'),
  ('permission.read', 'Read permissions'),
  ('permission.create', 'Create permissions'),
  ('permission.update', 'Update permissions'),
  ('permission.delete', 'Delete permissions'),
  ('role.permission.read', 'Read role permissions'),
  ('role.permission.update', 'Update role permissions')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
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
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;
