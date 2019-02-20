package pricing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
)

type RecordList []*Record

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

	sort.Slice(out, func(i, j int) bool { return out[i] > out[j] })
	return out
}

func (list RecordList) UsageType(usage string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].UsageType != usage {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) Region(region string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].Region != region {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) InstanceType(tipe string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].InstanceType != tipe {
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

func (list RecordList) LeaseContractLength(length string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].LeaseContractLength != length {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) Tenancy(tenancy string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].Tenancy != tenancy {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) PurchaseOption(purchase string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].PurchaseOption != purchase {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) PreInstalled(preinstalled string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].PreInstalled != preinstalled {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) OperatingSystem(os string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].OperatingSystem != os {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

func (list RecordList) OfferingClass(class string) RecordList {
	ret := RecordList{}

	for i := range list {
		if list[i].OfferingClass != class {
			continue
		}
		ret = append(ret, list[i])
	}

	return ret
}

type Record struct {
	Version                 string  `json:"version"`                             // common
	SKU                     string  `json:"sku"`                                 // common
	OfferTermCode           string  `json:"offer_term_code"`                     // common
	Region                  string  `json:"region"`                              // common
	InstanceType            string  `json:"instance_type"`                       // common
	UsageType               string  `json:"usage_type"`                          // common
	LeaseContractLength     string  `json:"lease_contract_length"`               // common
	PurchaseOption          string  `json:"purchase_option"`                     // common
	OnDemand                float64 `json:"ondemand"`                            // common
	ReservedQuantity        float64 `json:"reserved_quantity"`                   // common
	ReservedHrs             float64 `json:"reserved_hrs"`                        // common
	Tenancy                 string  `json:"tenancy,omitempty"`                   // ec2: Shared, Host, Dedicated
	PreInstalled            string  `json:"pre_installed,omitempty"`             // ec2: SQL Web, SQL Ent, SQL Std, NA
	OperatingSystem         string  `json:"operating_system,omitempty"`          // ec2: Windows, Linux, SUSE, RHEL
	Operation               string  `json:"operation,omitempty"`                 // ec2
	OfferingClass           string  `json:"offering_class,omitempty"`            // ec2, rds
	NormalizationSizeFactor string  `json:"normalization_size_factor,omitempty"` // ec2, rds
	DatabaseEngine          string  `json:"database_engine,omitempty"`           // rds
	CacheEngine             string  `json:"cache_engine,omitempty"`              // cache
}

func (r *Record) ID() string {
	return fmt.Sprintf("%s.%s", r.SKU, r.OfferTermCode)
}

func (r *Record) IsInstance() bool {
	if len(r.OperatingSystem) > 0 {
		return true
	}
	return false
}

func (r *Record) IsCacheNode() bool {
	if len(r.CacheEngine) > 0 {
		return true
	}
	return false
}

func (r *Record) IsDatabase() bool {
	if len(r.DatabaseEngine) > 0 {
		return true
	}
	return false
}

func (r *Record) OSEngine() string {
	if r.IsInstance() {
		return r.OperatingSystem
	}

	if r.IsDatabase() {
		return r.DatabaseEngine
	}

	if r.IsCacheNode() {
		return r.CacheEngine
	}

	return ""
}

func (r *Record) String() string {
	bytea, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return string(bytea)
}

func (r *Record) BreakevenPointInMonth() int {
	month := 12
	if r.LeaseContractLength == "3yr" {
		month = 12 * 3
	}

	breakevenPoint := 0
	res := r.ReservedQuantity
	ond := 0.0
	for i := 1; i < month+1; i++ {
		ond = ond + r.OnDemand*24*float64(GetDays(i))
		res = res + r.ReservedHrs*24*float64(GetDays(i))
		if ond > res {
			breakevenPoint = i
			break
		}
	}

	return breakevenPoint
}

type RecommendedList []*Recommended

func (list RecommendedList) Merge() RecommendedList {
	flat := make(map[string]*Recommended)
	for i := range list {
		if list[i].MinimumRecord == nil {
			key := fmt.Sprintf("%s_%s_%s_%s_%s",
				list[i].Record.Region,
				list[i].Record.UsageType,
				list[i].Record.OperatingSystem,
				list[i].Record.CacheEngine,
				list[i].Record.DatabaseEngine,
			)

			flat[key] = &Recommended{
				Record:                     list[i].Record,
				MinimumRecord:              list[i].Record,
				MinimumReservedInstanceNum: float64(list[i].ReservedInstanceNum),
			}
			continue
		}

		key := fmt.Sprintf("%s_%s_%s_%s_%s",
			list[i].MinimumRecord.Region,
			list[i].MinimumRecord.UsageType,
			list[i].MinimumRecord.OperatingSystem,
			list[i].MinimumRecord.CacheEngine,
			list[i].MinimumRecord.DatabaseEngine,
		)

		v, ok := flat[key]
		if ok {
			flat[key] = &Recommended{
				Record:                     v.Record,
				MinimumRecord:              v.MinimumRecord,
				MinimumReservedInstanceNum: v.MinimumReservedInstanceNum + list[i].MinimumReservedInstanceNum,
			}
			continue
		}

		flat[key] = list[i]
	}

	out := RecommendedList{}
	for _, v := range flat {
		out = append(out, v)
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Record.UsageType < out[j].Record.UsageType
	})

	return out
}

func (list RecommendedList) Header() []interface{} {
	return []interface{}{
		"usage_type",
		"os/engine",
		"ondemand_num_avg",
		"reserved_num",
		"full_ondemand_cost",
		"reserved_applied_cost",
		"reserved_applied_cost.ondemand",
		"reserved_applied_cost.reserved",
		"saving_cost",
		"discount_rate",
		"reserved_quantity",
	}
}

func (list RecommendedList) Array() [][]interface{} {
	out := [][]interface{}{}
	for _, r := range list {
		out = append(out, []interface{}{
			r.Record.UsageType,
			r.Record.OSEngine(),
			r.OnDemandInstanceNumAvg,
			r.ReservedInstanceNum,
			r.FullOnDemandCost,
			r.ReservedAppliedCost.Total,
			r.ReservedAppliedCost.OnDemand,
			r.ReservedAppliedCost.Reserved,
			r.SavingCost,
			r.DiscountRate,
			r.ReservedQuantity,
		})
	}

	return out
}

type Recommended struct {
	Record                     *Record `json:"record"`
	BreakevenPointInMonth      int     `json:"breakevenpoint_in_month"`
	Strategy                   string  `json:"strategy"`
	OnDemandInstanceNumAvg     float64 `json:"ondemand_instance_num_avg"`
	ReservedInstanceNum        int64   `json:"reserved_instance_num"`
	FullOnDemandCost           float64 `json:"full_ondemand_cost"`
	ReservedAppliedCost        Cost    `json:"reserved_applied_cost"`
	ReservedQuantity           float64 `json:"reserved_quantity"`
	SavingCost                 float64 `json:"saving_cost"`
	DiscountRate               float64 `json:"discount_rate"`
	MinimumRecord              *Record `json:"minimum_record,omitempty"`
	MinimumReservedInstanceNum float64 `json:"minimum_reserved_instance_num,omitempty"`
}

func (r *Recommended) String() string {
	bytea, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return string(bytea)
}

func (r *Recommended) Pretty() string {
	bytea, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	if err := json.Indent(&out, bytea, "", " "); err != nil {
		panic(err)
	}

	return string(out.Bytes())
}

type Forecast struct {
	Date        string  `json:"month"` // 2018-11
	InstanceNum float64 `json:"instance_num"`
}

type ForecastList []*Forecast

type ReservedAppliedCost struct {
	LeaseContractLength string  `json:"lease_contract_length"`
	PurchaseOption      string  `json:"purchase_option"`
	FullOnDemand        float64 `json:"full_ondemand"`
	ReservedApplied     Cost    `json:"reserved_applied"`
	SavingCost          float64 `json:"saving_cost"`
	DiscountRate        float64 `json:"discount_rate"`
	ReservedQuantity    float64 `json:"reserved_quantity"`
}

func (r *ReservedAppliedCost) String() string {
	bytea, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return string(bytea)
}

type Cost struct {
	OnDemand float64 `json:"ondemand"`
	Reserved float64 `json:"reserved"`
	Total    float64 `json:"total"`
}

func (r *Record) Recommend(forecast ForecastList, strategy ...string) *Recommended {
	actual, ondemand, reserved := r.GetInstanceNum(forecast, strategy...)
	cost := r.GetCost(ondemand, reserved)

	return &Recommended{
		Record:                 r,
		BreakevenPointInMonth:  r.BreakevenPointInMonth(),
		Strategy:               actual,
		OnDemandInstanceNumAvg: ondemand,
		ReservedInstanceNum:    reserved,
		FullOnDemandCost:       cost.FullOnDemand,
		ReservedAppliedCost:    cost.ReservedApplied,
		ReservedQuantity:       cost.ReservedQuantity,
		SavingCost:             cost.SavingCost,
		DiscountRate:           cost.DiscountRate,
	}
}

func (r *Record) GetInstanceNum(forecast ForecastList, strategy ...string) (string, float64, int64) {
	bep := r.BreakevenPointInMonth()
	if len(forecast) < bep {
		sum := 0.0
		for i := range forecast {
			sum = sum + forecast[i].InstanceNum
		}

		return "nothing", sum / float64(len(forecast)), 0
	}

	tmp := append(ForecastList{}, forecast...)
	sort.Slice(tmp, func(i, j int) bool { return tmp[i].InstanceNum > tmp[j].InstanceNum })

	// default strategy is breakevenpoint
	actual := "breakevenpoint"
	reserved := int64(math.Floor(tmp[bep-1].InstanceNum))
	if len(strategy) > 0 && strings.ToLower(strategy[0]) == "minimum" {
		actual = "minimum"
		reserved = int64(math.Floor(tmp[len(tmp)-1].InstanceNum))
	}

	// ondemand average in 1year/3year
	sum := 0.0
	for i := range tmp {
		ond := tmp[i].InstanceNum - float64(reserved)
		if ond > 0 {
			sum = sum + ond
		}
	}
	ondemand := sum / float64(len(forecast))

	return actual, ondemand, reserved
}

/*
GetCost returns full ondemand cost and reserved instance applied cost.

ondemandNum, reservedNum is Per Year  (LeaseContractLength=1yr)
ondemandNum, reservedNum is Per 3Year (LeaseContractLength=3yr)
*/
func (r *Record) GetCost(ondemandNum float64, reservedNum int64) *ReservedAppliedCost {
	full := r.GetAnnualCost().OnDemand * (ondemandNum + float64(reservedNum))
	ond := r.GetAnnualCost().OnDemand * ondemandNum
	res := r.GetAnnualCost().Reserved * float64(reservedNum)

	out := &ReservedAppliedCost{
		LeaseContractLength: r.LeaseContractLength,
		PurchaseOption:      r.PurchaseOption,
		FullOnDemand:        full,
		ReservedApplied: Cost{
			OnDemand: ond,
			Reserved: res,
			Total:    ond + res,
		},
		ReservedQuantity: r.ReservedQuantity * float64(reservedNum),
	}

	out.SavingCost = full - out.ReservedApplied.Total
	out.DiscountRate = 1.0 - (out.ReservedApplied.Total / full)

	return out
}

func (r *Record) GetAnnualCost() *AnnualCost {
	ret := &AnnualCost{
		LeaseContractLength: r.LeaseContractLength,
		PurchaseOption:      r.PurchaseOption,
	}

	hrs := 365 * 24
	if r.LeaseContractLength == "3yr" {
		hrs = hrs * 3
	}

	ret.OnDemand = r.OnDemand * float64(hrs)
	ret.Reserved = r.ReservedQuantity + r.ReservedHrs*float64(hrs)
	ret.SavingCost = ret.OnDemand - ret.Reserved
	ret.DiscountRate = 1.0 - ret.Reserved/ret.OnDemand
	ret.ReservedQuantity = r.ReservedQuantity

	return ret
}

type AnnualCost struct {
	LeaseContractLength string  `json:"lease_contract_length"`
	PurchaseOption      string  `json:"purchase_option"`
	OnDemand            float64 `json:"ondemand"`
	Reserved            float64 `json:"reserved"`
	SavingCost          float64 `json:"saving_cost"`
	DiscountRate        float64 `json:"discount_rate"`
	ReservedQuantity    float64 `json:"reserved_quantity"`
}

func (r *AnnualCost) String() string {
	bytea, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return string(bytea)
}
