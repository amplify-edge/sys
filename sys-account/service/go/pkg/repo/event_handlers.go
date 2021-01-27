package repo

import (
	"context"
	"fmt"

	"github.com/getcouragenow/sys-share/sys-account/service/go/pkg"

	sharedCore "github.com/getcouragenow/sys-share/sys-core/service/go/pkg"
	sharedBus "github.com/getcouragenow/sys-share/sys-core/service/go/pkg/bus"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

func (ad *SysAccountRepo) onDeleteProject(ctx context.Context, in *sharedCore.EventRequest) (map[string]interface{}, error) {
	const projectIdKey = "project_id"
	deleteRequestMap, err := getEventIdMap(in, projectIdKey)
	if err != nil {
		return nil, err
	}
	projId := deleteRequestMap[projectIdKey].(string)
	_, err = ad.DeleteProject(ctx, &pkg.IdRequest{Id: projId})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"success":    true,
		"successMsg": fmt.Sprintf("successfully deleted project with id: %s", projId),
	}, nil
}

func (ad *SysAccountRepo) onDeleteOrg(ctx context.Context, in *sharedCore.EventRequest) (map[string]interface{}, error) {
	const orgIdKey = "org_id"
	deleteRequestMap, err := getEventIdMap(in, orgIdKey)
	if err != nil {
		return nil, err
	}
	orgId := deleteRequestMap[orgIdKey].(string)
	_, err = ad.DeleteOrg(ctx, &pkg.IdRequest{Id: orgId})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"success":    true,
		"successMsg": fmt.Sprintf("successfully deleted org with id: %s", orgId),
	}, nil
}

func (ad *SysAccountRepo) onDeleteAccount(ctx context.Context, in *sharedCore.EventRequest) (map[string]interface{}, error) {
	const accountIdKey = "account_id"
	deleteRequestMap, err := getEventIdMap(in, accountIdKey)
	if err != nil {
		return nil, err
	}
	accountId := deleteRequestMap[accountIdKey].(string)
	_, err = ad.DisableAccount(ctx, &pkg.DisableAccountRequest{AccountId: accountId})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"success":    true,
		"successMsg": fmt.Sprintf("successfully deleted account with id: %s", accountId),
	}, nil
}

func getEventIdMap(in *sharedCore.EventRequest, key string) (map[string]interface{}, error) {
	requestMap, err := coredb.UnmarshalToMap(in.JsonPayload)
	if err != nil {
		return nil, err
	}
	if requestMap[key] == nil || requestMap[key].(string) == "" {
		return nil, sharedBus.Error{
			Reason: sharedBus.ErrInvalidEventPayload,
			Err:    fmt.Errorf("%s id is not valid", key),
		}
	}
	return requestMap, nil
}

func (ad *SysAccountRepo) onCheckProjectExists(ctx context.Context, in *sharedCore.EventRequest) (map[string]interface{}, error) {
	const projectIdKey = "sys_account_project_ref_id"
	const projectNameKey = "sys_account_project_ref_name"
	requestMap, err := coredb.UnmarshalToMap(in.JsonPayload)
	if err != nil {
		return nil, err
	}
	rmap := map[string]interface{}{}
	if requestMap[projectIdKey] != nil && requestMap[projectIdKey].(string) != "" {
		rmap["id"] = requestMap[projectIdKey]
	}
	if requestMap[projectNameKey] != nil && requestMap[projectNameKey].(string) != "" {
		rmap["name"] = requestMap[projectNameKey]
	}
	proj, err := ad.store.GetProject(&coredb.QueryParams{Params: rmap})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"exists":     true,
		projectIdKey: proj.Id,
	}, nil
}

func (ad *SysAccountRepo) onCheckOrgExists(_ context.Context, in *sharedCore.EventRequest) (map[string]interface{}, error) {
	const orgIdKey = "sys_account_org_ref_id"
	const orgNameKey = "sys_account_org_ref_name"
	requestMap, err := coredb.UnmarshalToMap(in.JsonPayload)
	if err != nil {
		return nil, err
	}
	rmap := map[string]interface{}{}
	if requestMap[orgIdKey] != nil && requestMap[orgIdKey].(string) != "" {
		rmap["id"] = requestMap[orgIdKey]
	}
	if requestMap[orgNameKey] != nil && requestMap[orgNameKey].(string) != "" {
		rmap["name"] = requestMap[orgNameKey]
	}
	org, err := ad.store.GetOrg(&coredb.QueryParams{Params: rmap})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"exists": true,
		orgIdKey: org.Id,
	}, nil
}

func (ad *SysAccountRepo) onCheckAccountExists(ctx context.Context, in *sharedCore.EventRequest) (map[string]interface{}, error) {
	const accountIdKey = "sys_account_user_ref_id"
	const accountNameKey = "sys_account_user_ref_name"
	requestMap, err := coredb.UnmarshalToMap(in.JsonPayload)
	if err != nil {
		return nil, err
	}
	rmap := map[string]interface{}{}
	if requestMap[accountIdKey] != nil && requestMap[accountIdKey].(string) != "" {
		rmap["id"] = requestMap[accountIdKey]
	}
	if requestMap[accountNameKey] != nil && requestMap[accountNameKey].(string) != "" {
		rmap["email"] = requestMap[accountNameKey]
	}
	acc, err := ad.store.GetAccount(&coredb.QueryParams{Params: rmap})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"exists":     true,
		accountIdKey: acc.ID,
	}, nil
}

func (ad *SysAccountRepo) onGetAccountEmail(ctx context.Context, in *sharedCore.EventRequest) (map[string]interface{}, error) {
	const accountIdKey = "sys_account_user_ref_id"
	requestMap, err := coredb.UnmarshalToMap(in.JsonPayload)
	if err != nil {
		return nil, err
	}
	rmap := map[string]interface{}{}
	if requestMap[accountIdKey] != nil && requestMap[accountIdKey].(string) != "" {
		rmap["id"] = requestMap[accountIdKey]
	}
	acc, err := ad.store.GetAccount(&coredb.QueryParams{Params: rmap})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"exists": true,
		"email":  acc.Email,
	}, nil
}

func (ad *SysAccountRepo) onResetAllSysAccount(ctx context.Context, in *sharedCore.EventRequest) (map[string]interface{}, error) {
	err := ad.store.ResetAll(ad.initialSuperusersMail)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{}, nil
}
