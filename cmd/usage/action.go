package usage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/itsubaki/hermes/pkg/usage"
	"github.com/urfave/cli"
)

func Action(c *cli.Context) {
	dir := c.GlobalString("dir")
	format := c.String("format")
	date := usage.Last12Months()

	quantity, err := usage.Deserialize(dir, date)
	if err != nil {
		fmt.Errorf("deserialize: %v", err)
		os.Exit(1)
	}

	if format == "json" {
		bytes, err := json.Marshal(quantity)
		if err != nil {
			fmt.Printf("marshal: %v", err)
			os.Exit(1)
		}

		fmt.Println(string(bytes))
		return
	}

	if format == "csv" {
		tmp := make(map[string][]usage.Quantity)
		for _, q := range quantity {
			tmp[Hash(q)] = append(tmp[Hash(q)], q)
		}

		fmt.Printf("accountID, description, region, usage_type, os/engine, ")
		for i := range date {
			fmt.Printf("%s, ", date[i].YYYYMM())
		}
		fmt.Println()

		for _, v := range tmp {
			fmt.Printf("%s, %s, ", v[0].AccountID, v[0].Description)
			fmt.Printf("%s, %s, ", v[0].Region, v[0].UsageType)
			fmt.Printf("%s, ", fmt.Sprintf("%s%s%s", v[0].Platform, v[0].CacheEngine, v[0].DatabaseEngine))

			for i := range date {
				found := false
				for _, q := range v {
					if date[i].YYYYMM() == q.Date {
						fmt.Printf("%.3f, ", q.InstanceNum)
						found = true
						break
					}
				}

				if !found {
					fmt.Printf("0.0, ")
				}
			}
			fmt.Println()
		}

		return
	}
}
