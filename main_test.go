package main

import (
	"fmt"
	"testing"

	"github.com/itsubaki/hermes/pkg/hermes"
	"github.com/itsubaki/hermes/pkg/pricing"
	"github.com/itsubaki/hermes/pkg/usage"
)

func TestPackage2(t *testing.T) {
	// usage forecast by aws account
	// normalize usage forecast by date
	// merge normalized usage forecast by date
	// break-even point with purchase option
}

func TestPackage(t *testing.T) {
	// forecast quantity
	forecast, err := usage.Deserialize("/var/tmp/hermes", usage.Last12Months())
	if err != nil {
		t.Errorf("usage deserialize: %v", err)
	}
	monthly := hermes.MonthlyUsage(forecast)

	// purchase RI
	query := []struct {
		OSEngine string
		Price    pricing.Price
	}{
		{
			OSEngine: "Linux/UNIX",
			Price: pricing.Price{
				Region:                  "ap-northeast-1",
				UsageType:               "APN1-BoxUsage:c4.large",
				Tenancy:                 "Shared",
				PreInstalled:            "NA",
				OperatingSystem:         "Linux",
				OfferingClass:           "standard",
				LeaseContractLength:     "1yr",
				PurchaseOption:          "All Upfront",
				OnDemand:                0.126,
				ReservedQuantity:        738,
				ReservedHrs:             0,
				NormalizationSizeFactor: "4",
			},
		},
		{
			OSEngine: "Linux/UNIX",
			Price: pricing.Price{
				Region:                  "ap-northeast-1",
				UsageType:               "APN1-BoxUsage:c4.8xlarge",
				Tenancy:                 "Shared",
				PreInstalled:            "NA",
				OperatingSystem:         "Linux",
				OfferingClass:           "standard",
				LeaseContractLength:     "1yr",
				PurchaseOption:          "All Upfront",
				OnDemand:                0.126 * 16,
				ReservedQuantity:        738 * 16,
				ReservedHrs:             0 * 16,
				NormalizationSizeFactor: "64",
			},
		},
		{
			OSEngine: "Linux/UNIX",
			Price: pricing.Price{
				Region:                  "ap-northeast-1",
				UsageType:               "APN1-BoxUsage:c4.2xlarge",
				Tenancy:                 "Shared",
				PreInstalled:            "NA",
				OperatingSystem:         "Linux",
				OfferingClass:           "standard",
				LeaseContractLength:     "1yr",
				PurchaseOption:          "Partial Upfront",
				OnDemand:                0.504,
				ReservedQuantity:        1511,
				ReservedHrs:             0.172,
				NormalizationSizeFactor: "16",
			},
		},
		{
			OSEngine: "Linux/UNIX",
			Price: pricing.Price{
				Region:                  "ap-northeast-1",
				UsageType:               "APN1-BoxUsage:c4.4xlarge",
				Tenancy:                 "Shared",
				PreInstalled:            "NA",
				OperatingSystem:         "Linux",
				OfferingClass:           "standard",
				LeaseContractLength:     "1yr",
				PurchaseOption:          "No Upfront",
				OnDemand:                1.008,
				ReservedQuantity:        0,
				ReservedHrs:             0.722,
				NormalizationSizeFactor: "32",
			},
		},
	}

	plist, err := pricing.Deserialize("/var/tmp/hermes", []string{"ap-northeast-1"})
	if err != nil {
		t.Errorf("desirialize: %v", err)
	}

	nn := make([]hermes.Tuple, 0)
	for _, qq := range query {
		hash := hermes.Hash(fmt.Sprintf("%s%s", qq.Price.UsageType, qq.OSEngine))
		q, p, err := hermes.BreakEvenPoint(monthly[hash], qq.Price)
		if err != nil {
			t.Errorf("%v", err)
		}
		//fmt.Printf("%s %s\n", q, p)

		n, p, err := hermes.Normalize(q, p, plist)
		if err != nil {
			t.Errorf("normalize: %v", err)
		}
		//fmt.Printf("%s %s\n", n, p)

		nn = append(nn, hermes.Tuple{Quantity: n, Price: p})
	}

	for _, m := range hermes.Merge(nn) {
		fmt.Println(m)
	}
}
