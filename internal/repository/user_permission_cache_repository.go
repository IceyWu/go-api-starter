package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/model"
)

var _ UserPermissionCacheRepositoryInterface = (*UserPermissionCacheRepository)(nil)

type UserPermissionCacheRepository struct{ db *sqlx.DB }

func NewUserPermissionCacheRepository(db *sqlx.DB) *UserPermissionCacheRepository {
	return &UserPermissionCacheRepository{db}
}

const cacheColumns = `id,user_id,space_id,value,expires_at,created_at,updated_at`

func (r *UserPermissionCacheRepository) Upsert(c context.Context, v *model.UserPermissionCache) error {
	n := modelTime()
	if v.CreatedAt.IsZero() {
		v.CreatedAt = n
	}
	v.UpdatedAt = n
	query := `INSERT INTO user_permission_caches(user_id,space_id,value,expires_at,created_at,updated_at) VALUES(?,?,?,?,?,?) ON CONFLICT(user_id,space_id) DO UPDATE SET value=excluded.value,expires_at=excluded.expires_at,updated_at=excluded.updated_at`
	if r.db.DriverName() == "mysql" {
		query = `INSERT INTO user_permission_caches(user_id,space_id,value,expires_at,created_at,updated_at) VALUES(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE value=VALUES(value),expires_at=VALUES(expires_at),updated_at=VALUES(updated_at)`
	}
	_, e := r.db.ExecContext(c, query, v.UserID, v.SpaceID, v.Value, v.ExpiresAt, v.CreatedAt, v.UpdatedAt)
	return e
}
func (r *UserPermissionCacheRepository) FindByUserAndSpace(c context.Context, u, s uint) (*model.UserPermissionCache, error) {
	var v model.UserPermissionCache
	e := r.db.GetContext(c, &v, `SELECT `+cacheColumns+` FROM user_permission_caches WHERE user_id=? AND space_id=? LIMIT 1`, u, s)
	if noRows(e) {
		return nil, nil
	}
	return &v, e
}
func (r *UserPermissionCacheRepository) FindByUserID(c context.Context, u uint) ([]model.UserPermissionCache, error) {
	var v []model.UserPermissionCache
	e := r.db.SelectContext(c, &v, `SELECT `+cacheColumns+` FROM user_permission_caches WHERE user_id=?`, u)
	return v, e
}
func (r *UserPermissionCacheRepository) DeleteByUserID(c context.Context, u uint) error {
	_, e := r.db.ExecContext(c, `DELETE FROM user_permission_caches WHERE user_id=?`, u)
	return e
}
func (r *UserPermissionCacheRepository) DeleteByUserIDs(c context.Context, u []uint) error {
	if len(u) == 0 {
		return nil
	}
	q, a, e := sqlx.In(`DELETE FROM user_permission_caches WHERE user_id IN (?)`, u)
	if e != nil {
		return e
	}
	_, e = r.db.ExecContext(c, r.db.Rebind(q), a...)
	return e
}
func (r *UserPermissionCacheRepository) GetUserSpaceValues(c context.Context, u uint) (map[uint]uint64, error) {
	v, e := r.FindByUserID(c, u)
	m := map[uint]uint64{}
	for _, x := range v {
		m[x.SpaceID] = x.Value
	}
	return m, e
}
