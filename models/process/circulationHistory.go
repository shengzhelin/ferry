package process

import (
	"ferry/models/base"
)

/*
  @Author : lanyulei
*/

// 工單流轉歷史
type CirculationHistory struct {
	base.Model
	Title        string `gorm:"column:title; type: varchar(128)" json:"title" form:"title"`                    // 工單標題
	WorkOrder    int    `gorm:"column:work_order; type: int(11)" json:"work_order" form:"work_order"`          // 工單ID
	State        string `gorm:"column:state; type: varchar(128)" json:"state" form:"state"`                    // 工單狀態
	Source       string `gorm:"column:source; type: varchar(128)" json:"source" form:"source"`                 // 源節點ID
	Target       string `gorm:"column:target; type: varchar(128)" json:"target" form:"target"`                 // 目標節點ID
	Circulation  string `gorm:"column:circulation; type: varchar(128)" json:"circulation" form:"circulation"`  // 流轉ID
	Status       int    `gorm:"column:status; type: int(11)" json:"status" form:"status"`                      // 流轉狀態 1 同意， 0 拒絕， 2 其他
	Processor    string `gorm:"column:processor; type: varchar(45)" json:"processor" form:"processor"`         // 處理人
	ProcessorId  int    `gorm:"column:processor_id; type: int(11)" json:"processor_id" form:"processor_id"`    // 處理人ID
	CostDuration int64  `gorm:"column:cost_duration; type: int(11)" json:"cost_duration" form:"cost_duration"` // 處理時長
	Remarks      string `gorm:"column:remarks; type: longtext" json:"remarks" form:"remarks"`                  // 備註
}

func (CirculationHistory) TableName() string {
	return "p_work_order_circulation_history"
}
