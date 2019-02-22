package pricing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/itsubaki/hermes/internal/pricing"
)

type Repository struct {
	Region   []string   `json:"region"`
	Internal RecordList `json:"internal"`
}

func New(region []string) (*Repository, error) {
	repo := NewRepository(region)
	return repo, repo.Fetch()
}

func NewRepository(region []string) *Repository {
	return &Repository{
		Region: region,
	}
}

func (repo *Repository) Fetch() error {
	return repo.FetchWithClient(http.DefaultClient)
}

func (repo *Repository) FetchWithClient(client *http.Client) error {
	if err := repo.fetchWithClient(pricing.ComputeURL, client); err != nil {
		return err
	}

	if err := repo.fetchWithClient(pricing.DatabseURL, client); err != nil {
		return err
	}

	if err := repo.fetchWithClient(pricing.CacheURL, client); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) fetchWithClient(url string, client *http.Client) error {
	for _, r := range repo.Region {
		price, err := pricing.Fetch(url, r)
		if err != nil {
			return fmt.Errorf("get price url=%s: %v", url, err)
		}

		for k := range price {
			v := price[k]
			repo.Internal = append(repo.Internal, &Record{
				Version:                 v.Version,
				InstanceType:            v.InstanceType,
				LeaseContractLength:     v.LeaseContractLength,
				NormalizationSizeFactor: v.NormalizationSizeFactor,
				OfferTermCode:           v.OfferTermCode,
				OfferingClass:           v.OfferingClass,
				OnDemand:                v.OnDemand,
				OperatingSystem:         v.OperatingSystem,
				Operation:               v.Operation,
				PreInstalled:            v.PreInstalled,
				PurchaseOption:          v.PurchaseOption,
				Region:                  v.Region,
				ReservedHrs:             v.ReservedHrs,
				ReservedQuantity:        v.ReservedQuantity,
				SKU:                     v.SKU,
				Tenancy:                 v.Tenancy,
				UsageType:               v.UsageType,
				CacheEngine:             v.CacheEngine,
				DatabaseEngine:          v.DatabaseEngine,
			})
		}
	}

	return nil
}

func Read(path string) (*Repository, error) {
	read, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %v", err)
	}

	repo := &Repository{}
	if err := repo.Deserialize(read); err != nil {
		return nil, fmt.Errorf("new repository: %v", err)
	}

	return repo, nil
}

func (repo *Repository) Write(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}

	bytes, err := repo.Serialize()
	if err != nil {
		return fmt.Errorf("serialize: %v", err)
	}

	if err := ioutil.WriteFile(path, bytes, os.ModePerm); err != nil {
		return fmt.Errorf("write file: %v", err)
	}

	return nil
}

func (repo *Repository) Serialize() ([]byte, error) {
	bytes, err := json.Marshal(repo)
	if err != nil {
		return []byte{}, fmt.Errorf("marshal: %v", err)
	}

	return bytes, nil
}

func (repo *Repository) Deserialize(bytes []byte) error {
	if err := json.Unmarshal(bytes, repo); err != nil {
		return fmt.Errorf("unmarshal: %v", err)
	}

	return nil
}

func (repo *Repository) SelectAll() RecordList {
	return repo.Internal
}

func (repo *Repository) Normalize(record *Record) (*Record, error) {
	if strings.Contains(record.InstanceType, "cache") {
		return record, nil
	}

	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/apply_ri.html
	// Instance size flexibility does not apply to Reserved Instances
	// that are purchased for a specific Availability Zone,
	// bare metal instances,
	// Reserved Instances with dedicated tenancy,
	// and Reserved Instances for Windows,
	// Windows with SQL Standard,
	// Windows with SQL Server Enterprise,
	// Windows with SQL Server Web,
	// RHEL, and SLES.
	if strings.Contains(record.OSEngine(), "Windows") {
		return record, nil
	}

	if strings.Contains(record.OSEngine(), "Red Hat Enterprise Linux") {
		return record, nil
	}

	if strings.Contains(record.OSEngine(), "SUSE Linux") {
		return record, nil
	}

	if strings.Contains(record.Tenancy, "dedicated") {
		return record, nil
	}

	defined := []string{
		"nano",
		"micro",
		"small",
		"medium",
		"large",
		"xlarge",
	}

	instanceType := record.InstanceType
	familiy := instanceType[:strings.LastIndex(instanceType, ".")]

	if record.IsInstance() {
		tmp := RecordList{}
		for i := range defined {
			suspect := fmt.Sprintf("%s.%s", familiy, defined[i])
			for j := range repo.Internal {
				if repo.Internal[j].InstanceType == suspect &&
					strings.LastIndex(repo.Internal[j].UsageType, ".") > 0 {
					tmp = append(tmp, repo.Internal[j])
				}
			}
			if len(tmp) > 0 {
				break
			}
		}

		if len(tmp) < 1 {
			return nil, fmt.Errorf("undefined instance type=%s family=%s. defined=%v", instanceType, familiy, defined)
		}

		usageType := fmt.Sprintf("%s%s",
			record.UsageType[:strings.LastIndex(record.UsageType, ".")],
			tmp[0].UsageType[strings.LastIndex(tmp[0].UsageType, "."):],
		)

		rs := tmp.UsageType(usageType).
			OperatingSystem(record.OperatingSystem).
			LeaseContractLength(record.LeaseContractLength).
			PurchaseOption(record.PurchaseOption).
			PreInstalled(record.PreInstalled).
			OfferingClass(record.OfferingClass).
			Region(record.Region)

		if len(rs) != 1 {
			return nil, fmt.Errorf("invalid ec2 result set=%v", rs)
		}

		return rs[0], nil
	}

	if record.IsDatabase() {
		tmp := RecordList{}
		for i := range defined {
			suspect := fmt.Sprintf("%s.%s", familiy, defined[i])
			for j := range repo.Internal {
				if repo.Internal[j].InstanceType == suspect &&
					repo.Internal[j].DatabaseEngine == record.DatabaseEngine &&
					strings.LastIndex(repo.Internal[j].UsageType, ".") > 0 {
					tmp = append(tmp, repo.Internal[j])
				}
			}
			if len(tmp) > 0 {
				break
			}
		}

		if len(tmp) < 1 {
			return nil, fmt.Errorf("undefined instance type. defined=%v", defined)
		}

		usageType := fmt.Sprintf("%s%s",
			record.UsageType[:strings.LastIndex(record.UsageType, ".")],
			tmp[0].UsageType[strings.LastIndex(tmp[0].UsageType, "."):],
		)

		rs := tmp.UsageType(usageType).
			DatabaseEngine(record.DatabaseEngine).
			LeaseContractLength(record.LeaseContractLength).
			PurchaseOption(record.PurchaseOption).
			Region(record.Region)

		if len(rs) != 1 {
			return nil, fmt.Errorf("invalid database result set=%v", rs)
		}

		return rs[0], nil
	}

	return nil, fmt.Errorf("invalid record=%v", record)
}

func (repo *Repository) FindByInstanceType(tipe string) RecordList {
	out := RecordList{}
	for i := range repo.Internal {
		if repo.Internal[i].InstanceType == tipe {
			out = append(out, repo.Internal[i])
		}
	}

	return out
}

func (repo *Repository) FindByUsageType(tipe string) RecordList {
	out := RecordList{}
	for i := range repo.Internal {
		if repo.Internal[i].UsageType == tipe {
			out = append(out, repo.Internal[i])
		}
	}

	return out
}

func (repo *Repository) Recommend(record *Record, forecast ForecastList, strategy ...string) (*Recommended, error) {
	min, err := repo.Normalize(record)
	if err != nil {
		return nil, fmt.Errorf("normalize record=%v: %v", record, err)
	}

	out := record.Recommend(forecast, strategy...)
	out.NormalizedRecord = min
	out.NormalizedInstanceNum = float64(out.ReservedInstanceNum) * 1.0

	// cache hasnt normalization size factor
	if record.IsCacheNode() {
		return out, nil
	}

	rf64, err := strconv.ParseFloat(record.NormalizationSizeFactor, 64)
	if err != nil {
		return nil, fmt.Errorf("parse float normalization size factor in record: %v", err)
	}

	mf64, err := strconv.ParseFloat(min.NormalizationSizeFactor, 64)
	if err != nil {
		return nil, fmt.Errorf("parse float normalization size factor in normalized record: %v", err)
	}

	scale := rf64 / mf64
	out.NormalizedInstanceNum = float64(out.ReservedInstanceNum) * scale

	return out, nil
}

func Download(region []string, dir string) error {
	path := fmt.Sprintf("%s/pricing", dir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}

	for _, r := range region {
		cache := fmt.Sprintf("%s/%s.out", path, r)
		if _, err := os.Stat(cache); !os.IsNotExist(err) {
			continue
		}

		repo := NewRepository([]string{r})
		if err := repo.Fetch(); err != nil {
			return fmt.Errorf("fetch pricing (region=%s): %v", r, err)
		}

		if err := repo.Write(cache); err != nil {
			return fmt.Errorf("write pricing (region=%s): %v", r, err)
		}

		fmt.Printf("write: %v\n", cache)
	}

	return nil
}
