package seed

import (
	"context"
	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/model"
	"go-api-starter/internal/platform/auth"
	"log"
	"strings"
	"time"
)

var permMeta = map[string][2]string{"user.create": {"创建用户", "允许创建新用户"}, "user.read": {"查看用户", "允许查看用户列表和详情"}, "user.update": {"编辑用户", "允许编辑用户信息"}, "user.delete": {"删除用户", "允许删除用户"}, "role.create": {"创建角色", "允许创建角色"}, "role.read": {"查看角色", "允许查看角色"}, "role.update": {"编辑角色", "允许编辑角色"}, "role.delete": {"删除角色", "允许删除角色"}, "permission.create": {"创建权限", "允许创建权限"}, "permission.read": {"查看权限", "允许查看权限"}, "permission.update": {"编辑权限", "允许编辑权限"}, "permission.delete": {"删除权限", "允许删除权限"}, "file.upload": {"上传文件", "允许上传文件"}, "file.delete": {"删除文件", "允许删除文件"}}
var presetCodes = []string{"user.create", "user.read", "user.update", "user.delete", "role.create", "role.read", "role.update", "role.delete", "permission.create", "permission.read", "permission.update", "permission.delete", "file.upload", "file.delete"}

func SyncPermissions(db *sqlx.DB, codes []string) {
	set := map[string]bool{}
	for _, v := range append(codes, presetCodes...) {
		set[v] = true
	}
	ctx := context.Background()
	for code := range set {
		var id uint
		if db.GetContext(ctx, &id, "SELECT id FROM permissions WHERE code=? LIMIT 1", code) == nil {
			continue
		}
		module := strings.SplitN(code, ".", 2)[0]
		space := "system"
		if module == "file" {
			space = "content"
		}
		var sid uint
		if db.GetContext(ctx, &sid, "SELECT id FROM permission_spaces WHERE name=? LIMIT 1", space) != nil {
			r, e := db.ExecContext(ctx, "INSERT INTO permission_spaces(name,description,is_active,created_at,updated_at) VALUES(?,?,?,?,?)", space, space+" 权限空间", true, modelTime(), modelTime())
			if e != nil {
				continue
			}
			x, _ := r.LastInsertId()
			sid = uint(x)
		}
		var count int
		_ = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM permissions WHERE space_id=?", sid)
		meta := permMeta[code]
		if meta[0] == "" {
			meta = [2]string{code, code}
		}
		_, _ = db.ExecContext(ctx, "INSERT INTO permissions(code,name,description,space_id,position,value,module,is_active,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)", code, meta[0], meta[1], sid, count, 1<<count, module, true, modelTime(), modelTime())
	}
	log.Printf("[seed] permissions synchronized")
}
func SyncAdminUser(db *sqlx.DB, email, password string) {
	if email == "" {
		return
	}
	var id uint
	if db.Get(&id, "SELECT id FROM users WHERE email=? LIMIT 1", email) == nil {
		return
	}
	hash, e := auth.HashPassword(password)
	if e != nil {
		return
	}
	_, e = db.Exec("INSERT INTO users(uid,lp_id,username,email,password,sex,freezed,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)", model.GenerateUID(), model.GenerateLPID(), model.GenerateUsername(), email, hash, 0, false, modelTime(), modelTime())
	if e != nil {
		log.Printf("[seed] admin create failed: %v", e)
	}
}
func SyncAdminRole(db *sqlx.DB, email string) {
	var roleID uint
	if db.Get(&roleID, "SELECT id FROM roles WHERE name='admin' LIMIT 1") != nil {
		x, e := db.Exec("INSERT INTO roles(name,description,is_active,is_system,created_at,updated_at) VALUES(?,?,?,?,?,?)", "admin", "超级管理员", true, true, modelTime(), modelTime())
		if e != nil {
			return
		}
		id, _ := x.LastInsertId()
		roleID = uint(id)
	}
	var rows []struct {
		ID      uint
		SpaceID uint
		Value   uint64
	}
	_ = db.Select(&rows, "SELECT id,space_id,value FROM permissions WHERE is_active=?", true)
	for _, p := range rows {
		_, _ = db.Exec("INSERT INTO role_permissions(role_id,permission_id,space_id,value,created_at,updated_at) VALUES(?,?,?,?,?,?)", roleID, p.ID, p.SpaceID, p.Value, modelTime(), modelTime())
	}
	if email != "" {
		var uid uint
		if db.Get(&uid, "SELECT id FROM users WHERE email=?", email) == nil {
			_, _ = db.Exec("INSERT INTO user_roles(user_id,role_id,created_at,updated_at) VALUES(?,?,?,?)", uid, roleID, modelTime(), modelTime())
		}
	}
}
func modelTime() time.Time { return time.Now() }
