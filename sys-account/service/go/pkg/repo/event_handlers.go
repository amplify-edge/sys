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
	if requestMap[projectIdKey] != "" {
		rmap["id"] = requestMap[projectIdKey]
	}
	if requestMap[projectNameKey] != "" {
		rmap["name"] = requestMap[projectNameKey]
	}
	_, err = ad.store.GetProject(&coredb.QueryParams{Params: rmap})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"exists": true,
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
	if requestMap[accountIdKey] != "" {
		rmap["id"] = requestMap[accountIdKey]
	}
	if requestMap[accountNameKey] != "" {
		rmap["name"] = requestMap[accountNameKey]
	}
	_, err = ad.store.GetAccount(&coredb.QueryParams{Params: rmap})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"exists": true,
	}, nil
}
