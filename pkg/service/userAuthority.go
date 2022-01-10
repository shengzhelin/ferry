package service

import (
	"encoding/json"
	"ferry/global/orm"
	"ferry/models/process"
	"ferry/models/system"
	"ferry/tools"

	"github.com/gin-gonic/gin"
)

/*
  @Author : lanyulei
  @todo: 添加新的處理人時候，需要修改（先完善功能，後續有時間的時候優化一下這部分。）
*/

func JudgeUserAuthority(c *gin.Context, workOrderId int, currentState string) (status bool, err error) {
	/*
		person 人員
		persongroup 人員組
		department 部門
		variable 變量
	*/
	var (
		userDept          system.Dept
		workOrderInfo     process.WorkOrderInfo
		userInfo          system.SysUser
		cirHistoryList    []process.CirculationHistory
		stateValue        map[string]interface{}
		processInfo       process.Info
		processState      ProcessState
		currentStateList  []map[string]interface{}
		currentStateValue map[string]interface{}
		currentUserInfo   system.SysUser
	)
	// 獲取工單訊息
	err = orm.Eloquent.Model(&workOrderInfo).
		Where("id = ?", workOrderId).
		Find(&workOrderInfo).Error
	if err != nil {
		return
	}

	// 獲取流程訊息
	err = orm.Eloquent.Model(&process.Info{}).Where("id = ?", workOrderInfo.Process).Find(&processInfo).Error
	//if err != nil {
	//	return
	//}

	if processInfo.Structure != nil && len(processInfo.Structure) > 0 {
		err = json.Unmarshal(processInfo.Structure, &processState.Structure)
		if err != nil {
			return
		}
	}

	stateValue, err = processState.GetNode(currentState)
	if err != nil {
		return
	}

	err = json.Unmarshal(workOrderInfo.State, &currentStateList)
	if err != nil {
		return
	}

	for _, v := range currentStateList {
		if v["id"].(string) == currentState {
			currentStateValue = v
			break
		}
	}

	// 獲取當前用戶訊息
	err = orm.Eloquent.Model(&currentUserInfo).
		Where("user_id = ?", tools.GetUserId(c)).
		Find(&currentUserInfo).
		Error
	if err != nil {
		return
	}

	// 會簽
	if currentStateValue["processor"] != nil && len(currentStateValue["processor"].([]interface{})) >= 1 {
		if isCounterSign, ok := stateValue["isCounterSign"]; ok {
			if isCounterSign.(bool) {
				err = orm.Eloquent.Model(&process.CirculationHistory{}).
					Where("work_order = ?", workOrderId).
					Order("id desc").
					Find(&cirHistoryList).Error
				if err != nil {
					return
				}
				for _, cirHistoryValue := range cirHistoryList {
					if cirHistoryValue.Source != stateValue["id"] {
						break
					} else if cirHistoryValue.Source == stateValue["id"] {
						if currentStateValue["process_method"].(string) == "person" {
							// 驗證個人會簽
							if cirHistoryValue.ProcessorId == tools.GetUserId(c) {
								return
							}
						} else if currentStateValue["process_method"].(string) == "role" {
							// 驗證角色會簽
							if stateValue["fullHandle"].(bool) {
								if cirHistoryValue.ProcessorId == tools.GetUserId(c) {
									return
								}
							} else {
								var roleUserInfo system.SysUser
								err = orm.Eloquent.Model(&roleUserInfo).
									Where("user_id = ?", cirHistoryValue.ProcessorId).
									Find(&roleUserInfo).
									Error
								if err != nil {
									return
								}
								if roleUserInfo.RoleId == tools.GetRoleId(c) {
									return
								}
							}
						} else if currentStateValue["process_method"].(string) == "department" {
							// 部門會簽
							if stateValue["fullHandle"].(bool) {
								if cirHistoryValue.ProcessorId == tools.GetUserId(c) {
									return
								}
							} else {
								var (
									deptUserInfo system.SysUser
								)
								err = orm.Eloquent.Model(&deptUserInfo).
									Where("user_id = ?", cirHistoryValue.ProcessorId).
									Find(&deptUserInfo).
									Error
								if err != nil {
									return
								}

								if deptUserInfo.DeptId == currentUserInfo.DeptId {
									return
								}
							}
						}
					}
				}
			}
		}
	}

	switch currentStateValue["process_method"].(string) {
	case "person":
		for _, processorValue := range currentStateValue["processor"].([]interface{}) {
			if int(processorValue.(float64)) == tools.GetUserId(c) {
				status = true
			}
		}
	case "role":
		for _, processorValue := range currentStateValue["processor"].([]interface{}) {
			if int(processorValue.(float64)) == tools.GetRoleId(c) {
				status = true
			}
		}
	case "department":
		for _, processorValue := range currentStateValue["processor"].([]interface{}) {
			if int(processorValue.(float64)) == currentUserInfo.DeptId {
				status = true
			}
		}
	case "variable":
		for _, p := range currentStateValue["processor"].([]interface{}) {
			switch int(p.(float64)) {
			case 1:
				if workOrderInfo.Creator == tools.GetUserId(c) {
					status = true
				}
			case 2:
				err = orm.Eloquent.Model(&userInfo).Where("user_id = ?", workOrderInfo.Creator).Find(&userInfo).Error
				if err != nil {
					return
				}
				err = orm.Eloquent.Model(&userDept).Where("dept_id = ?", userInfo.DeptId).Find(&userDept).Error
				if err != nil {
					return
				}

				if userDept.Leader == tools.GetUserId(c) {
					status = true
				}
			}
		}
	}
	return
}
