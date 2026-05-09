package seed

import (
	"context"
	"log"
	"strings"

	"go-api-starter/internal/model"
	"go-api-starter/pkg/auth"

	"gorm.io/gorm"
)

// permMeta 为权限 code 提供可读名称和描述
// 如果 code 不在这里，会自动用 code 本身作为 name
var permMeta = map[string][2]string{
	// [0]=中文名称  [1]=描述
	"user.create": {"创建用户", "允许创建新用户"},
	"user.read":   {"查看用户", "允许查看用户列表和详情"},
	"user.update": {"编辑用户", "允许编辑用户信息"},
	"user.delete": {"删除用户", "允许删除用户"},
	"role.manage": {"角色管理", "允许管理角色、权限和用户角色分配"},
	"file.upload": {"上传文件", "允许上传和编辑文件"},
	"file.delete": {"删除文件", "允许删除文件"},
}

// moduleToSpace 将 module 映射到权限空间
var moduleToSpace = map[string]string{
	"user": "system",
	"role": "system",
	"file": "content",
}

// SyncPermissions 根据路由中实际使用的权限 code 自动同步到数据库
// codes 来自 PermissionMiddleware.CollectedCodes()，是路由注册时自动收集的
func SyncPermissions(db *gorm.DB, codes []string) {
	if len(codes) == 0 {
		return
	}
	ctx := context.Background()

	// 1. 查出数据库已有的 code，避免重复创建
	var existingPerms []model.Permission
	db.WithContext(ctx).Select("code").Find(&existingPerms)
	existingSet := make(map[string]struct{}, len(existingPerms))
	for _, p := range existingPerms {
		existingSet[p.Code] = struct{}{}
	}

	// 过滤出需要新建的 code
	var newCodes []string
	for _, code := range codes {
		if _, exists := existingSet[code]; !exists {
			newCodes = append(newCodes, code)
		}
	}
	if len(newCodes) == 0 {
		log.Println("[seed] 权限已全部同步，无需新增")
		return
	}

	// 2. 确保权限空间存在
	spaceCache := make(map[string]*model.PermissionSpace)
	for _, code := range newCodes {
		spaceName := spaceNameFromCode(code)
		if _, ok := spaceCache[spaceName]; ok {
			continue
		}
		var space model.PermissionSpace
		if err := db.WithContext(ctx).Where("name = ?", spaceName).First(&space).Error; err != nil {
			space = model.PermissionSpace{Name: spaceName, Description: spaceName + " 权限空间", IsActive: true}
			if err := db.WithContext(ctx).Create(&space).Error; err != nil {
				log.Printf("[seed] 创建权限空间 %s 失败: %v", spaceName, err)
				continue
			}
			log.Printf("[seed] 创建权限空间: %s", spaceName)
		}
		spaceCache[spaceName] = &space
	}

	// 3. 统计各空间已有权限数量（用于 position）
	posMap := make(map[uint]int)
	for _, space := range spaceCache {
		var count int64
		db.WithContext(ctx).Model(&model.Permission{}).Where("space_id = ?", space.ID).Count(&count)
		posMap[space.ID] = int(count)
	}

	// 4. 创建新权限
	for _, code := range newCodes {
		spaceName := spaceNameFromCode(code)
		space, ok := spaceCache[spaceName]
		if !ok {
			continue
		}

		name, desc := metaFromCode(code)
		module := moduleFromCode(code)
		pos := posMap[space.ID]

		perm := model.Permission{
			Code:        code,
			Name:        name,
			Description: desc,
			SpaceID:     space.ID,
			Position:    uint8(pos),
			Value:       1 << uint(pos),
			Module:      module,
			IsActive:    true,
		}
		if err := db.WithContext(ctx).Create(&perm).Error; err != nil {
			log.Printf("[seed] 创建权限 %s 失败: %v", code, err)
			continue
		}
		posMap[space.ID] = pos + 1
		log.Printf("[seed] 自动创建权限: %s (%s)", code, name)
	}
}

// moduleFromCode 从 "topic.create" 提取 "topic"
func moduleFromCode(code string) string {
	parts := strings.SplitN(code, ".", 2)
	return parts[0]
}

// spaceNameFromCode 从 code 推断权限空间名称
func spaceNameFromCode(code string) string {
	module := moduleFromCode(code)
	if space, ok := moduleToSpace[module]; ok {
		return space
	}
	return "system" // 默认归到 system
}

// metaFromCode 获取权限的中文名称和描述
func metaFromCode(code string) (name, desc string) {
	if meta, ok := permMeta[code]; ok {
		return meta[0], meta[1]
	}
	// 没有预定义元数据时，用 code 本身
	return code, code
}

// SyncAdminUser 确保默认管理员账号存在（首次初始化时自动创建）
func SyncAdminUser(db *gorm.DB, email, password string) {
	ctx := context.Background()
	var user model.User
	if err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err == nil {
		return // 已存在，跳过
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("[seed] 密码加密失败: %v", err)
		return
	}

	user = model.User{
		SecUID:   model.GenerateSecUID(),
		Email:    &email,
		Password: &hashedPassword,
	}
	if err := db.WithContext(ctx).Create(&user).Error; err != nil {
		log.Printf("[seed] 创建管理员账号失败: %v", err)
		return
	}
	log.Printf("[seed] 已创建默认管理员账号: %s", email)
}

// SyncAdminRole 确保 admin 角色拥有所有权限，并分配给指定邮箱的用户
func SyncAdminRole(db *gorm.DB, adminEmail string) {
	ctx := context.Background()

	// 1. 确保 admin 角色存在
	var role model.Role
	if err := db.WithContext(ctx).Where("name = ?", "admin").First(&role).Error; err != nil {
		role = model.Role{Name: "admin", Description: "超级管理员，拥有所有权限", IsActive: true, IsSystem: true}
		if err := db.WithContext(ctx).Create(&role).Error; err != nil {
			log.Printf("[seed] 创建 admin 角色失败: %v", err)
			return
		}
		log.Println("[seed] 创建 admin 角色")
	}

	// 2. 获取所有权限
	var allPerms []model.Permission
	db.WithContext(ctx).Where("is_active = ?", true).Find(&allPerms)

	// 3. 获取 admin 角色已有的权限 ID
	var existingRP []model.RolePermission
	db.WithContext(ctx).Where("role_id = ?", role.ID).Find(&existingRP)
	existingPermIDs := make(map[uint]struct{}, len(existingRP))
	for _, rp := range existingRP {
		existingPermIDs[rp.PermissionID] = struct{}{}
	}

	// 4. 补齐缺失的权限
	added := 0
	for _, perm := range allPerms {
		if _, exists := existingPermIDs[perm.ID]; exists {
			continue
		}
		rp := model.RolePermission{
			RoleID:       role.ID,
			PermissionID: perm.ID,
			SpaceID:      perm.SpaceID,
			Value:        perm.Value,
		}
		if err := db.WithContext(ctx).Create(&rp).Error; err != nil {
			log.Printf("[seed] 给 admin 角色添加权限 %s 失败: %v", perm.Code, err)
			continue
		}
		added++
	}
	if added > 0 {
		log.Printf("[seed] 给 admin 角色补充了 %d 条权限", added)
	}

	// 5. 确保指定用户拥有 admin 角色
	if adminEmail == "" {
		return
	}
	var user model.User
	if err := db.WithContext(ctx).Where("email = ?", adminEmail).First(&user).Error; err != nil {
		log.Printf("[seed] 未找到用户 %s，跳过角色分配", adminEmail)
		return
	}

	var existingUR model.UserRole
	if err := db.WithContext(ctx).Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&existingUR).Error; err != nil {
		ur := model.UserRole{UserID: user.ID, RoleID: role.ID}
		if err := db.WithContext(ctx).Create(&ur).Error; err != nil {
			log.Printf("[seed] 给用户 %s 分配 admin 角色失败: %v", adminEmail, err)
			return
		}
		log.Printf("[seed] 已将 admin 角色分配给 %s", adminEmail)
	}

	// 6. 清除该用户的权限缓存，使新权限立即生效
	db.WithContext(ctx).Where("user_id = ?", user.ID).Delete(&model.UserPermissionCache{})
}
