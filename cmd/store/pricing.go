package store

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/itsubaki/hermes/pkg/pricing"
	"github.com/urfave/cli"
)

func ActionStorePricing(c *cli.Context) {
	project := c.String("project")
	dir := c.GlobalString("dir")
	region := c.StringSlice("region")

	if err := StorePricing(project, dir, region); err != nil {
		fmt.Printf("store pricing: %v", err)
		os.Exit(1)
	}
}

func StorePricing(project, dir string, region []string) error {
	ctx := context.Background()
	ds, err := datastore.NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("new datastore client: %v", err)
	}

	kind := "pricing"
	for i := range region {
		repo, err := pricing.Read(fmt.Sprintf("%s/%s/%s.out", dir, kind, region[i]))
		if err != nil {
			return fmt.Errorf("read pricing (region=%s): %v", region[i], err)
		}

		var key []*datastore.Key
		var src []interface{}

		rs := repo.SelectAll()
		for j := range rs {
			src = append(src, rs[j])
			key = append(key, &datastore.Key{
				Kind:   kind,
				Name:   rs[j].ID(),
				Parent: nil,
			})
		}

		for j := range key {
			if _, err := ds.Put(ctx, key[j], src[j]); err != nil {
				return fmt.Errorf("put entity: %v", err)
			}
			fmt.Printf("put (region=%s): %v\n", region[i], key[j])
		}
	}

	return nil
}
