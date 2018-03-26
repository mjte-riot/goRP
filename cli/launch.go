package cli

import (
	"errors"
	"fmt"
	"github.com/avarabyeu/goRP/gorp"
	"gopkg.in/urfave/cli.v1"
	"strings"
	"time"
)

var (
	launchCommand = cli.Command{
		Name:        "launch",
		Usage:       "Operations over launches",
		Subcommands: cli.Commands{listLaunchesCommand},
	}

	listLaunchesCommand = cli.Command{
		Name:  "list",
		Usage: "List launches",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "fn, filter-name",
				Usage:  "Filter Name",
				EnvVar: "FILTER_NAME",
			},
			cli.StringSliceFlag{
				Name:   "f, filter",
				Usage:  "Filter",
				EnvVar: "Filter",
			},
		},
		Action: listLaunches,
	}

	mergeCommand = cli.Command{
		Name:   "merge",
		Usage:  "Merge Launches",
		Action: mergeLaunches,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "f, filter",
				Usage:  "Launches Filter",
				EnvVar: "MERGE_LAUNCH_FILTER",
			},
			cli.StringSliceFlag{
				Name:   "ids",
				Usage:  "Launch IDS to Merge",
				EnvVar: "MERGE_LAUNCH_IDS",
			},

			cli.StringFlag{
				Name:   "n, name",
				Usage:  "New Launch Name",
				EnvVar: "MERGE_LAUNCH_NAME",
			},
			cli.StringFlag{
				Name:   "t, type",
				Usage:  "Merge Type",
				EnvVar: "MERGE_TYPE",
				Value:  "DEEP",
			},
		},
	}
)

func mergeLaunches(c *cli.Context) error {
	rpClient, err := buildClient(c)
	if nil != err {
		return err
	}

	ids, err := getMergeIDs(c, rpClient)
	if nil != err {
		return err
	}
	rq := &gorp.MergeLaunchesRQ{
		Name:      c.String("name"),
		MergeType: gorp.MergeType(c.String("type")),
		Launches:  ids,
		StartTime: gorp.Timestamp{Time: time.Now().Add(-10 * time.Hour)},
		EndTime:   gorp.Timestamp{Time: time.Now().Add(-1 * time.Minute)},
	}
	launchResource, err := rpClient.MergeLaunches(rq)
	if nil != err {
		return err
	}
	fmt.Println(launchResource.ID)
	return nil
}

func listLaunches(c *cli.Context) error {
	rpClient, err := buildClient(c)
	if nil != err {
		return err
	}

	var launches *gorp.LaunchPage

	if filters := c.StringSlice("filter"); nil != filters && len(filters) > 0 {
		filter := strings.Join(filters, "&")
		launches, err = rpClient.GetLaunchesByFilterString(filter)
	} else if filterName := c.String("filter-name"); "" != filterName {
		launches, err = rpClient.GetLaunchesByFilterName(filterName)
	} else {
		launches, err = rpClient.GetLaunches()
	}
	if nil != err {
		return err
	}

	for _, launch := range launches.Content {
		fmt.Printf("%s #%d \"%s\"\n", launch.ID, launch.Number, launch.Name)
	}
	return nil
}

func getMergeIDs(c *cli.Context, rpClient *gorp.Client) ([]string, error) {
	if ids := c.StringSlice("ids"); nil != ids && len(ids) > 0 {
		return ids, nil
	}

	filter := c.String("filter")
	if "" == filter {
		return nil, errors.New("no either IDs or filter provided")
	}
	launchesByFilterName, err := rpClient.GetLaunchesByFilterName(filter)
	if nil != err {
		return nil, err
	}
	ids := make([]string, len(launchesByFilterName.Content))
	for i, l := range launchesByFilterName.Content {
		ids[i] = l.ID
	}
	return ids, nil
}
