package repository

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/model"
)

var (
	ErrPermissionSpaceNotFound   = errors.New("permission space not found")
	ErrPermissionSpaceNameExists = errors.New("permission space name already exists")
)
var _ PermissionSpaceRepositoryInterface = (*PermissionSpaceRepository)(nil)

type PermissionSpaceRepository struct{ db *sqlx.DB }

func NewPermissionSpaceRepository(db *sqlx.DB) *PermissionSpaceRepository {
	return &PermissionSpaceRepository{db}
}

const spaceColumns = `id,name,description,is_active,created_at,updated_at`

func (r *PermissionSpaceRepository) Create(c context.Context, v *model.PermissionSpace) error {
	n := modelTime()
	v.CreatedAt = n
	v.UpdatedAt = n
	x, e := r.db.ExecContext(c, `INSERT INTO permission_spaces(name,description,is_active,created_at,updated_at) VALUES(?,?,?,?,?)`, v.Name, v.Description, v.IsActive, v.CreatedAt, v.UpdatedAt)
	if e == nil {
		if id, z := x.LastInsertId(); z == nil {
			v.ID = uint(id)
		}
	}
	return e
}
func (r *PermissionSpaceRepository) FindByName(c context.Context, n string) (*model.PermissionSpace, error) {
	return r.find(c, "name = ?", n)
}
func (r *PermissionSpaceRepository) FindByID(c context.Context, id uint) (*model.PermissionSpace, error) {
	return r.find(c, "id = ?", id)
}
func (r *PermissionSpaceRepository) find(c context.Context, p string, a any) (*model.PermissionSpace, error) {
	var v model.PermissionSpace
	e := r.db.GetContext(c, &v, `SELECT `+spaceColumns+` FROM permission_spaces WHERE `+p+` LIMIT 1`, a)
	if noRows(e) {
		return nil, ErrPermissionSpaceNotFound
	}
	return &v, e
}
func (r *PermissionSpaceRepository) FindAll(c context.Context) ([]model.PermissionSpace, error) {
	var v []model.PermissionSpace
	e := r.db.SelectContext(c, &v, `SELECT `+spaceColumns+` FROM permission_spaces ORDER BY id`)
	return v, e
}
func (r *PermissionSpaceRepository) FindAllWithCount(c context.Context) ([]model.SpaceWithCount, error) {
	var v []model.SpaceWithCount
	e := r.db.SelectContext(c, &v, `SELECT s.id,s.name,s.description,s.is_active,COUNT(p.id) permission_count FROM permission_spaces s LEFT JOIN permissions p ON p.space_id=s.id GROUP BY s.id ORDER BY s.id`)
	return v, e
}
func (r *PermissionSpaceRepository) Exists(c context.Context, n string) (bool, error) {
	var x int
	e := r.db.GetContext(c, &x, `SELECT COUNT(*) FROM permission_spaces WHERE name=?`, n)
	return x > 0, e
}
func (r *PermissionSpaceRepository) Update(c context.Context, v *model.PermissionSpace) error {
	v.UpdatedAt = modelTime()
	_, e := r.db.ExecContext(c, `UPDATE permission_spaces SET name=?,description=?,is_active=?,updated_at=? WHERE id=?`, v.Name, v.Description, v.IsActive, v.UpdatedAt, v.ID)
	return e
}
func (r *PermissionSpaceRepository) Delete(c context.Context, id uint) error {
	_, e := r.db.ExecContext(c, `DELETE FROM permission_spaces WHERE id=?`, id)
	return e
}
