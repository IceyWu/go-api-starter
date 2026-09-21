package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/model"
)

var ErrUserNotFound = errors.New("user not found")
var _ UserRepositoryInterface = (*UserRepository)(nil)

type UserRepository struct{ db *sqlx.DB }

func NewUserRepository(db *sqlx.DB) *UserRepository { return &UserRepository{db: db} }

const userColumns = `id, uid, lp_id, username, mobile, email, open_id, password, avatar_file_id, background_file_id, sex, birthday, city, job, company, signature, website, freezed, created_at, updated_at`

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	user.PrepareForCreate()
	now := user.CreatedAt
	if now.IsZero() {
		now = time.Now()
	}
	user.CreatedAt, user.UpdatedAt = now, now
	result, err := r.db.ExecContext(ctx, `INSERT INTO users (uid, lp_id, username, mobile, email, open_id, password, avatar_file_id, background_file_id, sex, birthday, city, job, company, signature, website, freezed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, user.UID, user.LPID, user.Username, user.Mobile, user.Email, user.OpenID, user.Password, user.AvatarFileID, user.BackgroundFileID, user.Sex, user.Birthday, user.City, user.Job, user.Company, user.Signature, user.Website, user.Freezed, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err == nil {
		user.ID = uint(id)
	}
	return err
}
func (r *UserRepository) FindAll(ctx context.Context, offset, limit int, sort string) ([]model.User, int64, error) {
	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM users`); err != nil {
		return nil, 0, err
	}
	if sort == "" || !safeSort(sort) {
		sort = "created_at DESC"
	}
	var users []model.User
	err := r.db.SelectContext(ctx, &users, fmt.Sprintf(`SELECT %s FROM users ORDER BY %s LIMIT ? OFFSET ?`, userColumns, sort), limit, offset)
	return users, total, err
}
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	return r.find(ctx, "id = ?", id)
}
func (r *UserRepository) FindByEmail(ctx context.Context, v string) (*model.User, error) {
	return r.find(ctx, "email = ?", v)
}
func (r *UserRepository) FindByMobile(ctx context.Context, v string) (*model.User, error) {
	return r.find(ctx, "mobile = ?", v)
}
func (r *UserRepository) FindByOpenID(ctx context.Context, v string) (*model.User, error) {
	return r.find(ctx, "open_id = ?", v)
}
func (r *UserRepository) FindByUID(ctx context.Context, v string) (*model.User, error) {
	return r.find(ctx, "uid = ?", v)
}
func (r *UserRepository) FindByUsername(ctx context.Context, v string) (*model.User, error) {
	return r.find(ctx, "username = ?", v)
}
func (r *UserRepository) FindByLPID(ctx context.Context, v string) (*model.User, error) {
	return r.find(ctx, "lp_id = ?", v)
}
func (r *UserRepository) FindByEmailForAuth(ctx context.Context, v string) (*model.User, error) {
	return r.find(ctx, "email = ?", v)
}
func (r *UserRepository) FindByMobileForAuth(ctx context.Context, v string) (*model.User, error) {
	return r.find(ctx, "mobile = ?", v)
}
func (r *UserRepository) find(ctx context.Context, predicate string, arg any) (*model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user, `SELECT `+userColumns+` FROM users WHERE `+predicate+` LIMIT 1`, arg)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return &user, err
}
func (r *UserRepository) FindByIDs(ctx context.Context, ids []uint) ([]model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	q, args, err := sqlx.In(`SELECT `+userColumns+` FROM users WHERE id IN (?)`, ids)
	if err != nil {
		return nil, err
	}
	var users []model.User
	err = r.db.SelectContext(ctx, &users, r.db.Rebind(q), args...)
	return users, err
}
func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	u.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, `UPDATE users SET uid=?,lp_id=?,username=?,mobile=?,email=?,open_id=?,password=?,avatar_file_id=?,background_file_id=?,sex=?,birthday=?,city=?,job=?,company=?,signature=?,website=?,freezed=?,updated_at=? WHERE id=?`, u.UID, u.LPID, u.Username, u.Mobile, u.Email, u.OpenID, u.Password, u.AvatarFileID, u.BackgroundFileID, u.Sex, u.Birthday, u.City, u.Job, u.Company, u.Signature, u.Website, u.Freezed, u.UpdatedAt, u.ID)
	return err
}
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}
func safeSort(sort string) bool {
	for _, r := range sort {
		if !(r == '_' || r == ',' || r == ' ' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return false
		}
	}
	return true
}
