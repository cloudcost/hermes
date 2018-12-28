package costexp

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
)

type RecordList []*Record

func (list RecordList) String() string {
	bytea, err := json.Marshal(list)
	if err != nil {
		panic(err)
	}

	return string(bytea)
}

func (list RecordList) Pretty() string {
	bytea, err := json.Marshal(list)
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	if err := json.Indent(&out, bytea, "", " "); err != nil {
		panic(err)
	}

	return string(out.Bytes())
}

type Record struct {
	AccountID      string  `json:"account_id"`
	Description    string  `json:"description"`
	UsageType      string  `json:"usage_type"`
	Platform       string  `json:"platform,omitempty"`        // ec2
	DatabaseEngine string  `json:"database_engine,omitempty"` // rds
	CacheEngine    string  `json:"cache_engine,omitempty"`    // cache
	Date           string  `json:"date"`
	InstanceHour   float64 `json:"instance_hour"`
	InstanceNum    float64 `json:"instance_num"`
}

func (u *Record) String() string {
	bytea, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}

	return string(bytea)
}

func (list RecordList) Unique(fieldname string) []string {
	uniq := make(map[string]bool)
	for i := range list {
		ref := reflect.ValueOf(*list[i]).FieldByName(fieldname)
		val := ref.Interface().(string)
		if len(val) > 0 {
			uniq[val] = true
		}
	}

	out := []string{}
	for k := range uniq {
		out = append(out, k)
	}

	return out
}

func (list RecordList) InstanceNumAvg() float64 {
	if len(list) == 0 {
		return 0
	}

	sum := 0.0
	for i := range list {
		sum = sum + list[i].InstanceNum
	}

	return sum / float64(len(list))
}

func (list RecordList) InstanceHourAvg() float64 {
	if len(list) == 0 {
		return 0
	}

	sum := 0.0
	for i := range list {
		sum = sum + list[i].InstanceHour
	}

	return sum / float64(len(list))
}

func (list RecordList) AccountID(accountID string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].AccountID != accountID {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) UsageType(usageType string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].UsageType != usageType {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) Platform(platform string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].Platform != platform {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) Date(date string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].Date != date {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) CacheEngine(engine string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].CacheEngine != engine {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) DatabaseEngine(engine string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].DatabaseEngine != engine {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (r RecordList) Sort() RecordList {
	list := append(RecordList{}, r...)

	sort.SliceStable(list, func(i, j int) bool { return list[i].UsageType < list[j].UsageType })
	sort.SliceStable(list, func(i, j int) bool { return list[i].Platform < list[j].Platform })
	sort.SliceStable(list, func(i, j int) bool { return list[i].CacheEngine < list[j].CacheEngine })
	sort.SliceStable(list, func(i, j int) bool { return list[i].DatabaseEngine < list[j].DatabaseEngine })
	sort.SliceStable(list, func(i, j int) bool { return list[i].AccountID < list[j].AccountID })

	return list
}
