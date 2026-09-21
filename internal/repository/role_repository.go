package repository

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/model"
)

var (
	ErrRoleNotFound   = errors.New("role not found")
	ErrRoleNameExists = errors.New("role name already exists")
)
var _ RoleRepositoryInterface = (*RoleRepository)(nil)

type RoleRepository struct{ db *sqlx.DB }

func NewRoleRepository(db *sqlx.DB) *RoleRepository { return &RoleRepository{db} }

const roleColumns = `id,name,description,is_active,is_system,created_at,updated_at`

func (r *RoleRepository) Create(c context.Context, v *model.Role) error {
	n := modelTime()
	v.CreatedAt = n
	v.UpdatedAt = n
	x, e := r.db.ExecContext(c, `INSERT INTO roles(name,description,is_active,is_system,created_at,updated_at) VALUES(?,?,?,?,?,?)`, v.Name, v.Description, v.IsActive, v.IsSystem, n, n)
	if e == nil {
		if id, z := x.LastInsertId(); z == nil {
			v.ID = uint(id)
		}
	}
	return e
}
func (r *RoleRepository) FindByName(c context.Context, v string) (*model.Role, error) {
	return r.find(c, `name = ?`, v)
}
func (r *RoleRepository) FindByID(c context.Context, v uint) (*model.Role, error) {
	return r.find(c, `id = ?`, v)
}
func (r *RoleRepository) find(c context.Context, p string, a any) (*model.Role, error) {
	var v model.Role
	e := r.db.GetContext(c, &v, `SELECT `+roleColumns+` FROM roles WHERE `+p+` LIMIT 1`, a)
	if noRows(e) {
		return nil, ErrRoleNotFound
	}
	return &v, e
}
func (r *RoleRepository) FindByIDWithPermissions(c context.Context, id uint) (*model.Role, error) {
	v, e := r.FindByID(c, id)
	if e != nil {
		return nil, e
	}
	v.RolePermissions, e = (&RolePermissionRepository{r.db}).FindByRoleID(c, id)
	return v, e
}
func (r *RoleRepository) FindAll(c context.Context) ([]model.Role, error) {
	var v []model.Role
	e := r.db.SelectContext(c, &v, `SELECT `+roleColumns+` FROM roles ORDER BY id`)
	return v, e
}
func (r *RoleRepository) Update(c context.Context, v *model.Role) error {
	v.UpdatedAt = modelTime()
	_, e := r.db.ExecContext(c, `UPDATE roles SET name=?,description=?,is_active=?,is_system=?,updated_at=? WHERE id=?`, v.Name, v.Description, v.IsActive, v.IsSystem, v.UpdatedAt, v.ID)
	return e
}
func (r *RoleRepository) Delete(c context.Context, id uint) error {
	x, e := r.db.ExecContext(c, `DELETE FROM roles WHERE id=?`, id)
	if e != nil {
		return e
	}
	n, _ := x.RowsAffected()
	if n == 0 {
		return ErrRoleNotFound
	}
	return nil
}
func (r *RoleRepository) Exists(c context.Context, n string) (bool, error) {
	var x int
	e := r.db.GetContext(c, &x, `SELECT COUNT(*) FROM roles WHERE name=?`, n)
	return x > 0, e
}
