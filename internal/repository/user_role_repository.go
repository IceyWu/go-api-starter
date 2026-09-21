package repository

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/model"
)

var (
	ErrUserRoleNotFound      = errors.New("user role not found")
	ErrUserRoleAlreadyExists = errors.New("user already has this role")
)
var _ UserRoleRepositoryInterface = (*UserRoleRepository)(nil)

type UserRoleRepository struct{ db *sqlx.DB }

func NewUserRoleRepository(db *sqlx.DB) *UserRoleRepository { return &UserRoleRepository{db} }
func (r *UserRoleRepository) Create(c context.Context, v *model.UserRole) error {
	n := modelTime()
	_, e := r.db.ExecContext(c, `INSERT INTO user_roles(user_id,role_id,created_at,updated_at) VALUES(?,?,?,?)`, v.UserID, v.RoleID, n, n)
	return e
}
func (r *UserRoleRepository) Delete(c context.Context, u, role uint) error {
	x, e := r.db.ExecContext(c, `DELETE FROM user_roles WHERE user_id=? AND role_id=?`, u, role)
	if e != nil {
		return e
	}
	n, _ := x.RowsAffected()
	if n == 0 {
		return ErrUserRoleNotFound
	}
	return nil
}
func (r *UserRoleRepository) FindByUserID(c context.Context, id uint) ([]model.UserRole, error) {
	var v []model.UserRole
	e := r.db.SelectContext(c, &v, `SELECT id,user_id,role_id,created_at,updated_at FROM user_roles WHERE user_id=?`, id)
	return v, e
}
func (r *UserRoleRepository) FindByRoleID(c context.Context, id uint) ([]model.UserRole, error) {
	var v []model.UserRole
	e := r.db.SelectContext(c, &v, `SELECT id,user_id,role_id,created_at,updated_at FROM user_roles WHERE role_id=?`, id)
	return v, e
}
func (r *UserRoleRepository) Exists(c context.Context, u, role uint) (bool, error) {
	var n int
	e := r.db.GetContext(c, &n, `SELECT COUNT(*) FROM user_roles WHERE user_id=? AND role_id=?`, u, role)
	return n > 0, e
}
func (r *UserRoleRepository) GetUserIDsByRoleID(c context.Context, id uint) ([]uint, error) {
	var v []uint
	e := r.db.SelectContext(c, &v, `SELECT user_id FROM user_roles WHERE role_id=?`, id)
	return v, e
}
