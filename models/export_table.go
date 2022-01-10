package models

import (
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

//GetHostData func
//根據hostid獲取主機訊息和item相關訊息並返回
func GetHostData(hostid string) (hostd []HostsData, err error) {
	output := []string{"hostid", "host"}
	//log.Println(hostid)
	selectParentTemplates := []string{"templateid", "name"}
	selectItems := []string{"itemid", "key_", "name", "history", "trends", "state", "status"}
	rep, err := API.Call("host.get", Params{"output": output, "selectParentTemplates": selectParentTemplates,
		"hostids": hostid, "selectItems": selectItems})
	if err != nil {
		return []HostsData{}, err
	}

	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []HostsData{}, err
	}
	var hb []HostsData
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []HostsData{}, err
	}
	//log.Println(hb)
	return hb, nil
}

//GetExpTrendData func
//根據主機訊息數據和報表指標類型，找出需要的item列表
func GetExpTrendData(hostdata []HostsData, itemtype string) ([]Itm, string, string, error) {
	//hosttype string
	var hosttype, itemkey string
	//log.Println(hostdata[0].Host)
	for _, v := range hostdata[0].ParentTemplates {
		switch {
		case (strings.Contains(v.Name, "Linux") || strings.Contains(v.Name, "linux")):
			hosttype = "Linux"
		case (strings.Contains(v.Name, "Windows") || strings.Contains(v.Name, "windows")):
			hosttype = "Windows"
		default:
			hosttype = "Linux"
		}
	}
	var itemids []string
	switch itemtype {
	case "cpu":
		itemkey = "system.cpu.util[,idle]"
	case "mem":
		if hosttype == "Linux" {
			itemkey = "vm.memory.size[available]"
		} else {
			itemkey = "vm.memory.size[free]"
		}
	case "disk":
		itemkey = "vfs.fs.size"
	case "net_in":
		itemkey = "net.if.in"
	case "net_out":
		itemkey = "net.if.out"
	}
	//獲取主機訊息及itemtype
	var ity Itm
	var ityarr []Itm
	for _, v := range hostdata[0].Items {
		if strings.Contains(v.Key, itemkey) {
			itemids = append(itemids, v.Itemid)
			ity.ItemKey = v.Key
			ity.ItemName = v.Name
			ity.Itemids = v.Itemid
			ity.Status = v.Status
			ity.State = v.State
			ityarr = append(ityarr, ity)
		}
	}
	return ityarr, itemkey, hostdata[0].Host, nil

}

//GetTrenDataTable funcitemtype
//獲取主機數據訊息並輸出到[]bytes
func GetTrenDataTable(itd []Itm, itemkey, host, ItemType string, start, end int64) ([]FileSystemDataALL, []byte, error) {
	var filesystem []string
	var fileall []FileSystemDataALL
	switch {
	case strings.Contains(itemkey, "vfs.fs.size"):
		//獲取變量
		for _, v := range itd {
			if v.Status == "0" && v.State == "0" {
				mot := strings.Split(strings.Split(v.ItemKey, "[")[1], ",")
				filesystem = append(filesystem, mot[0])
			}
		}
		af := RemoveRepeatedElement(filesystem)

		//獲取掛在點對應的free空間
		for _, vv := range af {
			//filekey = append(filekey, "vfs.fs.size["+v+",/free]")
			freekey := "vfs.fs.size[" + vv + ",free]"
			//log.Println(freekey)
			var bcc FileSystemData
			var cc []FileSystemData
			for _, v := range itd {
				if v.ItemKey == freekey && v.Status == "0" && v.State == "0" {
					//獲取趨勢數據
					output := []string{"itemid", "clock", "num", "value_min", "value_avg", "value_max"}
					//now := strconv.FormatInt(time.Now().Unix(), 10)
					now := end
					//end := strconv.FormatInt(time.Now().Add(-200*time.Hour).Unix(), 10)
					end := start
					rep, err := API.Call("trend.get", Params{"output": output, "time_from": end,
						"time_till": now, "itemids": v.Itemids})
					if err != nil {
						return []FileSystemDataALL{}, []byte{}, err
					}
					//序列化
					hba, err := json.Marshal(rep.Result)
					if err != nil {
						return []FileSystemDataALL{}, []byte{}, err
					}

					var hb []Trend
					err = json.Unmarshal(hba, &hb)
					if err != nil {
						return []FileSystemDataALL{}, []byte{}, err
					}
					//數據組合
					for _, va := range hb {
						//	log.Println(va)
						bcc.ItemName = v.ItemName
						bcc.MountPoint = vv
						bcc.ItemKey = v.ItemKey
						//time start
						b, _ := strconv.ParseInt(va.Clock, 10, 64)
						bcc.Clock = time.Unix(b, 0).Format("2006-01-02 15:04:05")
						bcc.Num = va.Num
						bcc.ItemID = va.Itemid
						bcc.ValueAvg = va.ValueAvg
						bcc.ValueMax = va.ValueMax
						bcc.ValueMin = va.ValueMin
						cc = append(cc, bcc)
					}
				}
			}
			var ftt FileSystemDataALL
			ftt.FileSystemDataADD = cc
			ftt.MountPoint = vv
			fileall = append(fileall, ftt)
		}
		ma, err := Crt(fileall, host, ItemType, start, end)
		if err != nil {
			return []FileSystemDataALL{}, []byte{}, err
		}
		return fileall, ma, nil

	case strings.Contains(itemkey, "net.if.in"):
		//遍歷網卡
		for _, v := range itd {
			if v.Status == "0" && v.State == "0" {
				mot := strings.Split(strings.Split(v.ItemKey, "[")[1], "]")
				filesystem = append(filesystem, mot[0])
			}
		}

		af := RemoveRepeatedElement(filesystem)

		//獲取網卡in流量
		for _, vv := range af {
			//filekey = append(filekey, "vfs.fs.size["+v+",/free]")
			freekey := "net.if.in[" + vv + "]"
			log.Println(freekey)
			//log.Println(freekey)
			var bcc FileSystemData
			var cc []FileSystemData
			for _, v := range itd {
				if v.ItemKey == freekey && v.Status == "0" && v.State == "0" {
					//獲取趨勢數據
					output := []string{"itemid", "clock", "num", "value_min", "value_avg", "value_max"}
					//now := strconv.FormatInt(time.Now().Unix(), 10)
					now := end
					//end := strconv.FormatInt(time.Now().Add(-200*time.Hour).Unix(), 10)
					end := start
					rep, err := API.Call("trend.get", Params{"output": output, "time_from": end,
						"time_till": now, "itemids": v.Itemids})
					if err != nil {
						return []FileSystemDataALL{}, []byte{}, err
					}
					//序列化
					hba, err := json.Marshal(rep.Result)
					if err != nil {
						return []FileSystemDataALL{}, []byte{}, err
					}

					var hb []Trend
					err = json.Unmarshal(hba, &hb)
					if err != nil {
						return []FileSystemDataALL{}, []byte{}, err
					}
					//數據組合
					for _, va := range hb {
						//	log.Println(va)
						bcc.ItemName = v.ItemName
						bcc.MountPoint = vv
						bcc.ItemKey = v.ItemKey
						//time start
						b, _ := strconv.ParseInt(va.Clock, 10, 64)
						bcc.Clock = time.Unix(b, 0).Format("2006-01-02 15:04:05")
						bcc.Num = va.Num
						bcc.ItemID = va.Itemid
						bcc.ValueAvg = va.ValueAvg
						bcc.ValueMax = va.ValueMax
						bcc.ValueMin = va.ValueMin
						cc = append(cc, bcc)
					}
				}
			}
			var ftt FileSystemDataALL
			ftt.FileSystemDataADD = cc
			ftt.MountPoint = vv
			fileall = append(fileall, ftt)

		}
		ma, err := Crt(fileall, host, ItemType, start, end)
		if err != nil {
			return []FileSystemDataALL{}, []byte{}, err
		}
		return fileall, ma, nil
	case strings.Contains(itemkey, "net.if.out"):
		//遍歷網卡
		for _, v := range itd {
			if v.Status == "0" && v.State == "0" {
				mot := strings.Split(strings.Split(v.ItemKey, "[")[1], "]")
				filesystem = append(filesystem, mot[0])
			}
		}

		af := RemoveRepeatedElement(filesystem)

		//獲取網卡out流量
		for _, vv := range af {
			//filekey = append(filekey, "vfs.fs.size["+v+",/free]")
			freekey := "net.if.out[" + vv + "]"
			//log.Println(freekey)
			var bcc FileSystemData
			var cc []FileSystemData
			for _, v := range itd {
				if v.ItemKey == freekey && v.Status == "0" && v.State == "0" {
					//獲取趨勢數據
					output := []string{"itemid", "clock", "num", "value_min", "value_avg", "value_max"}
					//now := strconv.FormatInt(time.Now().Unix(), 10)
					now := end
					//end := strconv.FormatInt(time.Now().Add(-200*time.Hour).Unix(), 10)
					end := start
					rep, err := API.Call("trend.get", Params{"output": output, "time_from": end,
						"time_till": now, "itemids": v.Itemids})
					if err != nil {
						return []FileSystemDataALL{}, []byte{}, err
					}
					//序列化
					hba, err := json.Marshal(rep.Result)
					if err != nil {
						return []FileSystemDataALL{}, []byte{}, err
					}

					var hb []Trend
					err = json.Unmarshal(hba, &hb)
					if err != nil {
						return []FileSystemDataALL{}, []byte{}, err
					}
					//數據組合
					for _, va := range hb {
						//	log.Println(va)
						bcc.ItemName = v.ItemName
						bcc.MountPoint = vv
						bcc.ItemKey = v.ItemKey
						//time start
						b, _ := strconv.ParseInt(va.Clock, 10, 64)
						bcc.Clock = time.Unix(b, 0).Format("2006-01-02 15:04:05")
						bcc.Num = va.Num
						bcc.ItemID = va.Itemid
						bcc.ValueAvg = va.ValueAvg
						bcc.ValueMax = va.ValueMax
						bcc.ValueMin = va.ValueMin
						cc = append(cc, bcc)
					}
				}
			}
			var ftt FileSystemDataALL
			ftt.FileSystemDataADD = cc
			ftt.MountPoint = vv
			fileall = append(fileall, ftt)

		}
		ma, err := Crt(fileall, host, ItemType, start, end)
		if err != nil {
			return []FileSystemDataALL{}, []byte{}, err
		}
		return fileall, ma, nil

	case strings.Contains(itemkey, "system.cpu.util[,idle]"):
		//遍歷網卡
		for _, v := range itd {
			if v.Status == "0" && v.State == "0" {
				filesystem = append(filesystem, v.ItemKey)
			}
		}
		//獲取網卡out流量
		var bcc FileSystemData
		var cc []FileSystemData
		for _, v := range itd {
			// if v.ItemKey == freekey && v.Status == "0" && v.State == "0" {
			// 	//獲取趨勢數據
			output := []string{"itemid", "clock", "num", "value_min", "value_avg", "value_max"}
			now := end
			end := start
			rep, err := API.Call("trend.get", Params{"output": output, "time_from": end,
				"time_till": now, "itemids": v.Itemids})
			if err != nil {
				return []FileSystemDataALL{}, []byte{}, err
			}
			//序列化
			hba, err := json.Marshal(rep.Result)
			if err != nil {
				return []FileSystemDataALL{}, []byte{}, err
			}

			var hb []Trend
			err = json.Unmarshal(hba, &hb)
			if err != nil {
				return []FileSystemDataALL{}, []byte{}, err
			}
			//數據組合
			for _, va := range hb {
				//	log.Println(va)
				bcc.ItemName = v.ItemName
				bcc.MountPoint = v.ItemKey
				bcc.ItemKey = v.ItemKey
				//time start
				b, _ := strconv.ParseInt(va.Clock, 10, 64)
				bcc.Clock = time.Unix(b, 0).Format("2006-01-02 15:04:05")
				bcc.Num = va.Num
				bcc.ItemID = va.Itemid
				bcc.ValueAvg = va.ValueAvg
				bcc.ValueMax = va.ValueMax
				bcc.ValueMin = va.ValueMin
				cc = append(cc, bcc)
			}
			var ftt FileSystemDataALL
			ftt.FileSystemDataADD = cc
			ftt.MountPoint = v.ItemKey
			fileall = append(fileall, ftt)
		}
		ma, err := Crt(fileall, host, ItemType, start, end)
		if err != nil {
			return []FileSystemDataALL{}, []byte{}, err
		}
		return fileall, ma, nil
	case strings.Contains(itemkey, "vm.memory.size[available]") || strings.Contains(itemkey, "vm.memory.size[free]"):
		//遍歷key
		for _, v := range itd {
			if v.Status == "0" && v.State == "0" {
				filesystem = append(filesystem, v.ItemKey)
			}
		}
		var bcc FileSystemData
		var cc []FileSystemData
		for _, v := range itd {
			//獲取趨勢數據
			output := []string{"itemid", "clock", "num", "value_min", "value_avg", "value_max"}
			now := end
			end := start
			rep, err := API.Call("trend.get", Params{"output": output, "time_from": end,
				"time_till": now, "itemids": v.Itemids})
			if err != nil {
				return []FileSystemDataALL{}, []byte{}, err
			}
			//序列化
			hba, err := json.Marshal(rep.Result)
			if err != nil {
				return []FileSystemDataALL{}, []byte{}, err
			}

			var hb []Trend
			err = json.Unmarshal(hba, &hb)
			if err != nil {
				return []FileSystemDataALL{}, []byte{}, err
			}
			//數據組合
			for _, va := range hb {
				//	log.Println(va)
				bcc.ItemName = v.ItemName
				bcc.MountPoint = v.ItemKey
				bcc.ItemKey = v.ItemKey
				//time start
				b, _ := strconv.ParseInt(va.Clock, 10, 64)
				bcc.Clock = time.Unix(b, 0).Format("2006-01-02 15:04:05")
				bcc.Num = va.Num
				bcc.ItemID = va.Itemid
				bcc.ValueAvg = va.ValueAvg
				bcc.ValueMax = va.ValueMax
				bcc.ValueMin = va.ValueMin
				cc = append(cc, bcc)
			}
			var ftt FileSystemDataALL
			ftt.FileSystemDataADD = cc
			ftt.MountPoint = v.ItemKey
			fileall = append(fileall, ftt)
		}
		ma, err := Crt(fileall, host, ItemType, start, end)
		if err != nil {
			return []FileSystemDataALL{}, []byte{}, err
		}
		return fileall, ma, nil
	}
	return []FileSystemDataALL{}, []byte{}, nil
}

//GetTrenDataFileName trend data
//獲取趨勢數據並輸出為[]bytes
func GetTrenDataFileName(v ListQueryAll, start, end int64) ([]byte, error) {

	rep, err := GetTrendDataByItemid(v.Item, start, end)
	if err != nil {
		return []byte{}, err
	}
	//序列化
	hba, err := json.Marshal(rep)
	if err != nil {
		return []byte{}, err
	}
	//log.Println(string(hba))
	var hb []Trend
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []byte{}, err
	}

	ma, err := CreateTrenXlsx(hb, v, start, end)
	if err != nil {
		return []byte{}, err
	}
	return ma, nil
}

//GetHistoryDataFileName trend data
//獲取詳情數據並輸出到excel文件，返回[]byte
func GetHistoryDataFileName(v ListQueryAll, start, end int64) ([]byte, error) {
	rep, err := GetHistoryByItemIDNew(v.Item, start, end)
	if err != nil {
		return []byte{}, err
	}
	//序列化
	hba, err := json.Marshal(rep)
	if err != nil {
		return []byte{}, err
	}
	var hb []History
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []byte{}, err
	}

	//生成xlsx文件
	ma, err := CreateHistoryXlsx(hb, v, start, end)
	if err != nil {
		return []byte{}, err
	}
	return ma, nil
}

//RemoveRepeatedElement 數組去重
func RemoveRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

//Round floa
func Round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}
