package usage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/itsubaki/hermes/pkg/hermes"
	"github.com/itsubaki/hermes/pkg/pricing"

	"github.com/itsubaki/hermes/pkg/usage"
	"github.com/urfave/cli"
)

func Action(c *cli.Context) {
	region := c.StringSlice("region")
	dir := c.GlobalString("dir")
	format := c.String("format")
	normalize := c.Bool("normalize")
	merge := c.Bool("merge")
	overall := c.Bool("overall")
	monthly := c.Bool("monthly")

	date := usage.Last12Months()
	quantity, err := usage.Deserialize(dir, date)
	if err != nil {
		fmt.Errorf("deserialize usage: %v", err)
		os.Exit(1)
	}

	if normalize {
		plist, err := pricing.Deserialize("/var/tmp/hermes", region)
		if err != nil {
			fmt.Errorf("desirialize pricing: %v", err)
		}

		family := pricing.Family(plist)
		mini := pricing.Minimum(family, plist)

		quantity = hermes.Normalize(quantity, mini)
	}

	if merge && overall {
		quantity = usage.MergeOverall(quantity)
	}

	if merge && !overall {
		quantity = usage.Merge(quantity)
	}

	if format == "json" && !monthly {
		usage.Sort(quantity)
		for _, q := range quantity {
			bytes, err := json.Marshal(q)
			if err != nil {
				fmt.Printf("marshal: %v", err)
				os.Exit(1)
			}

			fmt.Println(string(bytes))
		}
		return
	}

	if format == "json" && monthly {
		mq := usage.Monthly(quantity)
		for _, q := range mq {
			bytes, err := json.Marshal(q)
			if err != nil {
				fmt.Printf("marshal: %v", err)
				os.Exit(1)
			}

			fmt.Println(string(bytes))
		}
		return
	}

	if format == "csv" {
		fmt.Printf("accountID, description, region, usage_type, os/engine, ")
		for i := range date {
			fmt.Printf("%s, ", date[i].YYYYMM())
		}
		fmt.Println()

		mq := usage.Monthly(quantity)
		keys := usage.SortedKey(mq)
		for _, k := range keys {
			fmt.Printf("%s, %s, ", mq[k][0].AccountID, mq[k][0].Description)
			fmt.Printf("%s, %s, ", mq[k][0].Region, mq[k][0].UsageType)
			fmt.Printf("%s, ", fmt.Sprintf("%s%s%s", mq[k][0].Platform, mq[k][0].CacheEngine, mq[k][0].DatabaseEngine))

			for _, d := range date {
				found := false
				for _, q := range mq[k] {
					if d.YYYYMM() == q.Date {
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
	}
}
