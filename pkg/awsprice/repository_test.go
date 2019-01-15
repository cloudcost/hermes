package awsprice

import (
	"fmt"
	"os"
	"testing"
)

func TestFindMinimumDatabaseT2Medium(t *testing.T) {
	path := fmt.Sprintf(
		"%s/%s/%s",
		os.Getenv("GOPATH"),
		"src/github.com/itsubaki/hermes/internal/_serialized/awsprice",
		"ap-northeast-1.out",
	)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("file not found: %v", path)
	}

	repo, err := NewRepository(path)
	if err != nil {
		t.Errorf("%v", err)
	}

	rs := repo.FindByInstanceType("db.t2.medium").
		PurchaseOption("All Upfront").
		LeaseContractLength("1yr").
		DatabaseEngine("Aurora MySQL")

	min, err := repo.FindMinimumInstanceType(rs[0])
	if err != nil {
		t.Errorf("%v", err)
	}

	fmt.Println(min)
}

func TestFindMinimumDatabase(t *testing.T) {
	path := fmt.Sprintf(
		"%s/%s/%s",
		os.Getenv("GOPATH"),
		"src/github.com/itsubaki/hermes/internal/_serialized/awsprice",
		"ap-northeast-1.out",
	)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("file not found: %v", path)
	}

	repo, err := NewRepository(path)
	if err != nil {
		t.Errorf("%v", err)
	}

	rs := repo.FindByInstanceType("db.m4.4xlarge").
		PurchaseOption("All Upfront").
		LeaseContractLength("1yr").
		DatabaseEngine("PostgreSQL")

	r, err := repo.FindMinimumInstanceType(rs[0])
	if err != nil {
		t.Errorf("find minimum instance type: %v", err)
	}

	if r.InstanceType != "db.m4.large" {
		t.Errorf("invalid minimum instance type=%s", r.InstanceType)
	}
}

func TestFindMinimumCompute(t *testing.T) {
	path := fmt.Sprintf(
		"%s/%s/%s",
		os.Getenv("GOPATH"),
		"src/github.com/itsubaki/hermes/internal/_serialized/awsprice",
		"ap-northeast-1.out",
	)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("file not found: %v", path)
	}

	repo, err := NewRepository(path)
	if err != nil {
		t.Errorf("%v", err)
	}

	r := &Record{
		SKU:                     "XU2NYYPCRTK4T7CN",
		OfferTermCode:           "6QCMYABX3D",
		Region:                  "ap-northeast-1",
		InstanceType:            "m4.4xlarge",
		UsageType:               "APN1-BoxUsage:m4.4xlarge",
		LeaseContractLength:     "1yr",
		PurchaseOption:          "All Upfront",
		OnDemand:                1.032,
		ReservedQuantity:        5700,
		ReservedHrs:             0,
		Tenancy:                 "Shared",
		PreInstalled:            "NA",
		OperatingSystem:         "Linux",
		Operation:               "RunInstances",
		OfferingClass:           "standard",
		NormalizationSizeFactor: "32",
	}

	min, err := repo.FindMinimumInstanceType(r)
	if err != nil {
		t.Errorf("find minimum instance type: %v", err)
	}

	if min.InstanceType != "m4.large" {
		t.Errorf("invalid minimum instance type=%s", min.InstanceType)
	}
}

func TestFindByInstanceType(t *testing.T) {
	path := fmt.Sprintf(
		"%s/%s/%s",
		os.Getenv("GOPATH"),
		"src/github.com/itsubaki/hermes/internal/_serialized/awsprice",
		"ap-northeast-1.out",
	)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("file not found: %v", path)
	}

	repo, err := NewRepository(path)
	if err != nil {
		t.Errorf("%v", err)
	}

	rs := repo.FindByInstanceType("m4.large").
		OperatingSystem("Linux").
		Tenancy("Shared").
		PreInstalled("NA").
		OfferingClass("standard")

	for _, r := range rs {
		if r.InstanceType != "m4.large" {
			t.Error("invalid instance type")
		}
		if r.OperatingSystem != "Linux" {
			t.Error("invalid operationg system")
		}
		if r.Tenancy != "Shared" {
			t.Error("invalid tenancy")
		}
	}

	for _, r := range rs {
		if r.GetAnnualCost().Subtraction < 0 {
			t.Error("invalid subtraction")
		}
		if r.GetAnnualCost().DiscountRate < 0 {
			t.Error("invalid discount rate")
		}
	}
}
