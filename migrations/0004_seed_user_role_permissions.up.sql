-- User role management permissions
INSERT INTO permissions (name, description)
VALUES
  ('user.role.read', 'Read user roles'),
  ('user.role.update', 'Update user roles')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
  'user.role.read',
  'user.role.update'
)
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;
