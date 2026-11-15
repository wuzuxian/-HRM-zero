package custom

import (
	"context"
	"encoding/json"
	"time"

	"github.com/casbin/casbin/v2"

	"github.com/jzero-io/jzero-admin/server/internal/global"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/resignation"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/transfer"
	"github.com/jzero-io/jzero-admin/server/internal/model"
	"github.com/jzero-io/jzero-admin/server/internal/scheduler"
	menutypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/manage/menu"
)

// Custom is the service-group hook for server-lifecycle logic that is not a
// REST route. It owns the in-process scheduler (服务内嵌定时调度骨架): domain
// packages expose NewXxxJob constructors and Start registers them; later
// tickets (HRM-10 离职生效、假勤日结/月报锁定) append their jobs here.
type Custom struct {
	scheduler *scheduler.Scheduler
}

func New() *Custom {
	return &Custom{}
}

// Init Please add custom logic here.
func (c *Custom) Init() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// auto gen casbin rules
	if err := InitCasbinRule(ctx, global.ServiceContext.Model, global.ServiceContext.CasbinEnforcer); err != nil {
		return err
	}

	return nil
}

// Start launches the in-process scheduler. Called by service.ServiceGroup in
// its own goroutine; Stop is invoked on server shutdown.
func (c *Custom) Start() {
	s := scheduler.New()
	s.Register(transfer.NewEffectiveScanJob(global.ServiceContext.SqlxConn, time.Now))
	s.Register(resignation.NewResignedScanJob(global.ServiceContext.SqlxConn, time.Now))
	c.scheduler = s
	s.Start()
}

// Stop shuts the scheduler down (cancels job contexts, waits for in-flight runs).
func (c *Custom) Stop() {
	if c.scheduler != nil {
		c.scheduler.Stop()
	}
}

func InitCasbinRule(ctx context.Context, model model.Model, enforcer *casbin.Enforcer) error {
	// get all role
	allRoles, err := model.ManageRole.FindByCondition(ctx, nil)
	if err != nil {
		return err
	}

	allMenus, err := model.ManageMenu.FindByCondition(ctx, nil)
	if err != nil {
		return err
	}

	// get role menu
	allRoleMenus, err := model.ManageRoleMenu.FindByCondition(ctx, nil)
	if err != nil {
		return err
	}

	var casbinRules [][]string
	for _, v := range allRoles {
		for _, arm := range allRoleMenus {
			if v.Uuid == arm.RoleUuid {
				for _, am := range allMenus {
					if arm.MenuUuid == am.Uuid {
						var permissions []menutypes.Permission
						err = json.Unmarshal([]byte(am.Permissions), &permissions)
						if err == nil {
							for _, perm := range permissions {
								if perm.Code != "" {
									if hasPolicy, _ := enforcer.HasPolicy(v.Uuid, perm.Code); !hasPolicy {
										casbinRules = append(casbinRules, []string{v.Uuid, perm.Code})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if len(casbinRules) > 0 {
		_, err = enforcer.AddPolicies(casbinRules)
		if err != nil {
			return err
		}
	}
	return nil
}
