package models

import (
	"bytes"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

//ExpInspect 按組巡檢報告導出
func ExpInspect(groupname string, Insp []Insp) ([]byte, error) {

	f := excelize.NewFile()
	// Create a new sheet.
	index := f.NewSheet("Sheet1")
	stylecenter, err := f.NewStyle(`{"alignment":{"horizontal":"center","vertical":"center","wrap_text":true}}`)
	if err != nil {
		return []byte{}, nil
	}

	style1111, err := f.NewStyle(`{"font":{"bold":true,"size":36,"color":"#777777"},
		"alignment":{"horizontal":"center","vertical":"center"}}`)
	if err != nil {
		return []byte{}, nil
	}
	f.MergeCell("Sheet1", "A1", "H1")
	f.SetCellStyle("Sheet1", "A1", "H1", style1111)
	f.SetCellValue("Sheet1", "A1", "服務器設備日常巡檢")
	//標題
	f.SetColWidth("Sheet1", "A", "A", 4)
	f.MergeCell("Sheet1", "A2", "A3")
	f.SetCellStyle("Sheet1", "G2", "G3", stylecenter)
	f.MergeCell("Sheet1", "B2", "B3")
	f.SetCellStyle("Sheet1", "B2", "B3", stylecenter)
	f.SetCellValue("Sheet1", "B2", "設備型號")
	f.MergeCell("Sheet1", "C2", "E2")
	f.SetCellStyle("Sheet1", "C2", "E2", stylecenter)
	f.SetCellStyle("Sheet1", "C3", "E3", stylecenter)
	f.SetColWidth("Sheet1", "C", "E", 12)
	f.SetCellValue("Sheet1", "C2", "運行狀態檢查")
	f.SetCellValue("Sheet1", "C3", "CPU使用率")
	f.SetCellValue("Sheet1", "D3", "內存使用率")
	f.SetCellValue("Sheet1", "E3", "硬碟狀態")
	f.SetCellValue("Sheet1", "E4", "正常□\n異常□")
	f.SetRowHeight("Sheet1", 2, 32)
	f.SetRowHeight("Sheet1", 3, 32)
	f.SetCellValue("Sheet1", "F2", "基本安全\n檢查")
	f.SetCellValue("Sheet1", "F3", "登入、系\n統安全")
	f.SetCellStyle("Sheet1", "F2", "F2", stylecenter)
	f.SetCellStyle("Sheet1", "F3", "F3", stylecenter)
	f.SetColWidth("Sheet1", "G", "G", 23.67)
	f.SetColWidth("Sheet1", "H", "H", 9.67)
	f.SetCellValue("Sheet1", "G2", "硬件檢查")
	f.SetCellValue("Sheet1", "G3", "指示燈、電源、風扇情況")
	f.SetCellValue("Sheet1", "H2", "網卡檢查")
	f.SetCellValue("Sheet1", "H3", "網卡狀態")
	f.SetCellStyle("Sheet1", "H2", "H3", stylecenter)

	i := 4
	for i = 4; i < len(Insp)*3+4; {
		f.MergeCell("Sheet1", "B"+strconv.Itoa(i), "B"+strconv.Itoa(i+2))
		f.MergeCell("Sheet1", "C"+strconv.Itoa(i), "C"+strconv.Itoa(i+2))
		f.MergeCell("Sheet1", "D"+strconv.Itoa(i), "D"+strconv.Itoa(i+2))
		f.MergeCell("Sheet1", "E"+strconv.Itoa(i), "E"+strconv.Itoa(i+2))
		f.SetCellStyle("Sheet1", "E"+strconv.Itoa(i), "E"+strconv.Itoa(i+2), stylecenter)
		f.SetCellValue("Sheet1", "E"+strconv.Itoa(i), "正常□\n異常□")
		f.MergeCell("Sheet1", "F"+strconv.Itoa(i), "F"+strconv.Itoa(i+2))
		f.SetCellStyle("Sheet1", "F"+strconv.Itoa(i), "F"+strconv.Itoa(i+2), stylecenter)
		f.SetCellValue("Sheet1", "F"+strconv.Itoa(i), "正常□\n異常□")
		f.SetCellValue("Sheet1", "G"+strconv.Itoa(i), "指示燈:正常□ 異常□")
		f.SetCellValue("Sheet1", "G"+strconv.Itoa(i+1), "電源:正常□ 異常□")
		f.SetCellValue("Sheet1", "G"+strconv.Itoa(i+2), "風扇:正常□ 異常□")
		f.MergeCell("Sheet1", "H"+strconv.Itoa(i), "H"+strconv.Itoa(i+2))
		f.SetCellStyle("Sheet1", "H"+strconv.Itoa(i), "H"+strconv.Itoa(i+2), stylecenter)
		f.SetCellValue("Sheet1", "H"+strconv.Itoa(i), "正常□\n異常□")
		i = i + 3
	}

	for i := 0; i < len(Insp); {
		f.SetCellValue("Sheet1", "B"+strconv.Itoa(i*3+4), Insp[i].HostName)
		f.SetCellStyle("Sheet1", "B"+strconv.Itoa(i*3+4), "B"+strconv.Itoa(i*3+4), stylecenter)
		f.SetCellValue("Sheet1", "C"+strconv.Itoa(i*3+4), Insp[i].CPULoad)
		f.SetCellStyle("Sheet1", "C"+strconv.Itoa(i*3+4), "C"+strconv.Itoa(i*3+4), stylecenter)
		f.SetCellValue("Sheet1", "D"+strconv.Itoa(i*3+4), Insp[i].MemPct)
		f.SetCellStyle("Sheet1", "D"+strconv.Itoa(i*3+4), "D"+strconv.Itoa(i*3+4), stylecenter)
		i = i + 1
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	StartUnix := time.Now().In(loc)
	StrStart := StartUnix.Format("2006-01-02 15:04:05")

	f.MergeCell("Sheet1", "A4", "A"+strconv.Itoa(i))
	f.SetCellValue("Sheet1", "A4", groupname)
	f.SetColWidth("Sheet1", "A", "A", 10)
	f.SetCellStyle("Sheet1", "A4", "A"+strconv.Itoa(i), stylecenter)

	f.SetCellValue("Sheet1", "D"+strconv.Itoa(i+1), "巡檢人員")
	f.SetCellValue("Sheet1", "F"+strconv.Itoa(i+1), "巡檢日期")
	f.SetCellValue("Sheet1", "G"+strconv.Itoa(i+1), StrStart)
	f.SetCellValue("Sheet1", "H"+strconv.Itoa(i+1), "早  □")
	f.SetCellValue("Sheet1", "H"+strconv.Itoa(i+2), "中  □")
	f.SetCellValue("Sheet1", "H"+strconv.Itoa(i+3), "晚  □")

	f.SetActiveSheet(index)
	var b bytes.Buffer
	err = f.Write(&b)
	if err != nil {
		return []byte{}, nil
	}
	return b.Bytes(), nil
}
