package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/model"
)

var (
	ErrPermissionNotFound   = errors.New("permission not found")
	ErrPermissionCodeExists = errors.New("permission code already exists")
)
var _ PermissionRepositoryInterface = (*PermissionRepository)(nil)

type PermissionRepository struct{ db *sqlx.DB }

func NewPermissionRepository(db *sqlx.DB) *PermissionRepository { return &PermissionRepository{db} }

const permissionColumns = `id,code,name,description,space_id,position,value,module,is_active,created_at,updated_at`

func (r *PermissionRepository) Create(c context.Context, v *model.Permission) error {
	n := modelTime()
	v.CreatedAt = n
	v.UpdatedAt = n
	x, e := r.db.ExecContext(c, `INSERT INTO permissions(code,name,description,space_id,position,value,module,is_active,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, v.Code, v.Name, v.Description, v.SpaceID, v.Position, v.Value, v.Module, v.IsActive, n, n)
	if e == nil {
		if id, z := x.LastInsertId(); z == nil {
			v.ID = uint(id)
		}
	}
	return e
}
func (r *PermissionRepository) FindByCode(c context.Context, v string) (*model.Permission, error) {
	return r.find(c, `code = ?`, v)
}
func (r *PermissionRepository) FindByID(c context.Context, v uint) (*model.Permission, error) {
	return r.find(c, `id = ?`, v)
}
func (r *PermissionRepository) find(c context.Context, p string, a any) (*model.Permission, error) {
	var v model.Permission
	e := r.db.GetContext(c, &v, `SELECT `+permissionColumns+` FROM permissions WHERE `+p+` LIMIT 1`, a)
	if noRows(e) {
		return nil, ErrPermissionNotFound
	}
	return &v, e
}
func (r *PermissionRepository) FindAll(c context.Context) ([]model.Permission, error) {
	var v []model.Permission
	e := r.db.SelectContext(c, &v, `SELECT `+permissionColumns+` FROM permissions ORDER BY space_id,position`)
	return v, e
}
func (r *PermissionRepository) FindBySpaceID(c context.Context, id uint) ([]model.Permission, error) {
	var v []model.Permission
	e := r.db.SelectContext(c, &v, `SELECT `+permissionColumns+` FROM permissions WHERE space_id=? ORDER BY position`, id)
	return v, e
}
func (r *PermissionRepository) GetMaxPositionInSpace(c context.Context, id uint) (int, error) {
	var v sql.NullInt64
	e := r.db.GetContext(c, &v, `SELECT MAX(position) FROM permissions WHERE space_id=?`, id)
	if e != nil {
		return -1, e
	}
	if !v.Valid {
		return -1, nil
	}
	return int(v.Int64), nil
}
func (r *PermissionRepository) Update(c context.Context, v *model.Permission) error {
	v.UpdatedAt = modelTime()
	_, e := r.db.ExecContext(c, `UPDATE permissions SET code=?,name=?,description=?,space_id=?,position=?,value=?,module=?,is_active=?,updated_at=? WHERE id=?`, v.Code, v.Name, v.Description, v.SpaceID, v.Position, v.Value, v.Module, v.IsActive, v.UpdatedAt, v.ID)
	return e
}
func (r *PermissionRepository) Delete(c context.Context, id uint) error {
	x, e := r.db.ExecContext(c, `DELETE FROM permissions WHERE id=?`, id)
	if e != nil {
		return e
	}
	n, _ := x.RowsAffected()
	if n == 0 {
		return ErrPermissionNotFound
	}
	return nil
}
func (r *PermissionRepository) Exists(c context.Context, v string) (bool, error) {
	var n int
	e := r.db.GetContext(c, &n, `SELECT COUNT(*) FROM permissions WHERE code=?`, v)
	return n > 0, e
}
func (r *PermissionRepository) FindByCodes(c context.Context, codes []string) ([]model.Permission, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	q, args, e := sqlx.In(`SELECT `+permissionColumns+` FROM permissions WHERE code IN (?)`, codes)
	if e != nil {
		return nil, e
	}
	var v []model.Permission
	e = r.db.SelectContext(c, &v, r.db.Rebind(q), args...)
	return v, e
}
func (r *PermissionRepository) CountBySpaceID(c context.Context, id uint) (int64, error) {
	var n int64
	e := r.db.GetContext(c, &n, `SELECT COUNT(*) FROM permissions WHERE space_id=?`, id)
	return n, e
}
