-- Sample roles and permissions
INSERT INTO roles (name, description)
VALUES
  ('admin', 'Full access'),
  ('deleter', 'Delete only')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, description)
VALUES
  ('user.create', 'Create user'),
  ('user.read', 'Read user'),
  ('user.update', 'Update user'),
  ('user.delete', 'Delete user')
ON CONFLICT (name) DO NOTHING;

-- Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Deleter gets delete permission only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name = 'user.delete'
WHERE r.name = 'deleter'
ON CONFLICT DO NOTHING;
