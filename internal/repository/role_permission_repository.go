package repository

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/model"
)

var ErrRolePermissionNotFound = errors.New("role permission not found")
var _ RolePermissionRepositoryInterface = (*RolePermissionRepository)(nil)

type RolePermissionRepository struct{ db *sqlx.DB }

func NewRolePermissionRepository(db *sqlx.DB) *RolePermissionRepository {
	return &RolePermissionRepository{db}
}

const rpColumns = `id,role_id,permission_id,space_id,value,created_at,updated_at`

func (r *RolePermissionRepository) Create(c context.Context, v *model.RolePermission) error {
	n := modelTime()
	v.CreatedAt = n
	v.UpdatedAt = n
	x, e := r.db.ExecContext(c, `INSERT INTO role_permissions(role_id,permission_id,space_id,value,created_at,updated_at) VALUES(?,?,?,?,?,?)`, v.RoleID, v.PermissionID, v.SpaceID, v.Value, n, n)
	if e == nil {
		if id, z := x.LastInsertId(); z == nil {
			v.ID = uint(id)
		}
	}
	return e
}
func (r *RolePermissionRepository) Update(c context.Context, v *model.RolePermission) error {
	v.UpdatedAt = modelTime()
	_, e := r.db.ExecContext(c, `UPDATE role_permissions SET role_id=?,permission_id=?,space_id=?,value=?,updated_at=? WHERE id=?`, v.RoleID, v.PermissionID, v.SpaceID, v.Value, v.UpdatedAt, v.ID)
	return e
}
func (r *RolePermissionRepository) Delete(c context.Context, role, p uint) error {
	x, e := r.db.ExecContext(c, `DELETE FROM role_permissions WHERE role_id=? AND permission_id=?`, role, p)
	if e != nil {
		return e
	}
	n, _ := x.RowsAffected()
	if n == 0 {
		return ErrRolePermissionNotFound
	}
	return nil
}
func (r *RolePermissionRepository) DeleteByRoleID(c context.Context, id uint) error {
	_, e := r.db.ExecContext(c, `DELETE FROM role_permissions WHERE role_id=?`, id)
	return e
}
func (r *RolePermissionRepository) FindByRoleID(c context.Context, id uint) ([]model.RolePermission, error) {
	var v []model.RolePermission
	e := r.db.SelectContext(c, &v, `SELECT `+rpColumns+` FROM role_permissions WHERE role_id=?`, id)
	return v, e
}
func (r *RolePermissionRepository) FindByRoleAndSpace(c context.Context, role, space uint) (*model.RolePermission, error) {
	return r.find(c, `role_id=? AND space_id=?`, role, space)
}
func (r *RolePermissionRepository) FindByRoleAndPermission(c context.Context, role, p uint) (*model.RolePermission, error) {
	return r.find(c, `role_id=? AND permission_id=?`, role, p)
}
func (r *RolePermissionRepository) find(c context.Context, w string, a ...any) (*model.RolePermission, error) {
	var v model.RolePermission
	e := r.db.GetContext(c, &v, `SELECT `+rpColumns+` FROM role_permissions WHERE `+w+` LIMIT 1`, a...)
	if noRows(e) {
		return nil, nil
	}
	return &v, e
}
func (r *RolePermissionRepository) Exists(c context.Context, role, p uint) (bool, error) {
	var n int
	e := r.db.GetContext(c, &n, `SELECT COUNT(*) FROM role_permissions WHERE role_id=? AND permission_id=?`, role, p)
	return n > 0, e
}
func (r *RolePermissionRepository) GetSpaceValuesByRoleID(c context.Context, id uint) (map[uint]uint64, error) {
	var v []struct {
		SpaceID uint   `db:"space_id"`
		Value   uint64 `db:"value"`
	}
	e := r.db.SelectContext(c, &v, `SELECT space_id,SUM(value) value FROM role_permissions WHERE role_id=? GROUP BY space_id`, id)
	m := map[uint]uint64{}
	for _, x := range v {
		m[x.SpaceID] = x.Value
	}
	return m, e
}
