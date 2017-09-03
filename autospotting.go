package main

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/chrisurwin/autospotting/core"
	"github.com/chrisurwin/autospotting/healthcheck"
	"github.com/cristim/ec2-instances-info"
	"github.com/urfave/cli"
)

var VERSION = "v0.0.0-dev"

type cfgData struct {
	*autospotting.Config
}

var conf *cfgData

func main() {
	app := cli.NewApp()
	app.Name = "Autospotting"
	app.Version = VERSION
	app.Usage = "Replace AWS On Demand instances with Spot Instances"
	app.Action = start
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug,d",
			Usage:  "Debug logging",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:   "aws-key,k",
			Usage:  "AWS Access Key",
			EnvVar: "AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "aws-secret,s",
			Usage:  "AWS Secret Key",
			EnvVar: "AWS_SECRET_ACCESS_KEY",
		},
		cli.DurationFlag{
			Name:   "poll-interval,i",
			Value:  60 * time.Second,
			Usage:  "Polling interval for checks",
			EnvVar: "POLL_INTERVAL",
		},
		cli.StringFlag{
			Name:   "arn,a",
			Usage:  "AWS Role ARN",
			EnvVar: "ARN",
		},
		cli.StringFlag{
			Name: "regions",
			Usage: "Regions where it should be activated (comma or whitespace separated list, " +
				"also supports globs), by default it runs on all regions.\n\t" +
				"Example: ./autospotting -regions 'eu-*,us-east-1'",
			EnvVar: "regions",
			Value:  "",
		},
		cli.Int64Flag{
			Name: "min_on_demand_number",
			Usage: "On-demand capacity (as absolute number) ensured to be running in each of your groups.\n\t" +
				"Can be overridden on a per-group basis using the tag " +
				autospotting.OnDemandNumberLong,
			EnvVar: "min_on_demand_number",
			Value:  0,
		},
		cli.Float64Flag{
			Name: "min_on_demand_percentage",
			Usage: "On-demand capacity (percentage of the total number of instances in the group) " +
				"ensured to be running in each of your groups.\n\t" +
				"Can be overridden on a per-group basis using the tag " +
				autospotting.OnDemandPercentageLong +
				"\n\tIt is ignored if min_on_demand_number is also set.",
			EnvVar: "min_on_demand_percentage",
			Value:  0.0,
		},
		cli.StringFlag{
			Name: "allowed_instance_types",
			Usage: "If specified, the spot instances will have a specific instance type:\n" +
				"\tcurrent: the same as initial on-demand instances\n" +
				"\t<instance-type>: the actual instance type to use",
			EnvVar: "allowed_instance_types",
			Value:  "",
		},
		cli.StringFlag{
			Name:   "tag_name",
			Usage:  "If specified you can tag instances with a specific tag to process, default is spot-enabled",
			EnvVar: "tag_name",
			Value:  "spot-enabled",
		},
	}
	app.Run(os.Args)

}

func start(c *cli.Context) {
	var region string

	if r := os.Getenv("AWS_REGION"); r != "" {
		region = r
	} else {
		region = endpoints.UsEast1RegionID
	}

	conf = &cfgData{
		&autospotting.Config{
			LogFile:         os.Stdout,
			LogFlag:         log.Ldate | log.Ltime | log.Lshortfile,
			MainRegion:      region,
			SleepMultiplier: 1,
		},
	}
	go healthcheck.StartHealthcheck()
	conf.initialize(c)
	run()

	ticker := time.NewTicker(time.Minute * 1)
	for _ = range ticker.C {
		run()
	}
}

func run() {
	log.Println("Starting autospotting agent, build:", VERSION)

	log.Printf("Parsed command line flags: "+
		"regions='%s' "+
		"min_on_demand_number=%d "+
		"min_on_demand_percentage=%.1f "+
		"allowed_instance_types='%s' "+"tag_name='%s'",
		conf.Regions,
		conf.MinOnDemandNumber,
		conf.MinOnDemandPercentage,
		conf.AllowedInstanceTypes,
		conf.TagName)

	autospotting.Run(conf.Config)
	log.Println("Execution completed, nothing left to do")
}

// Configuration handling
func (c *cfgData) initialize(cli *cli.Context) {

	c.Regions = cli.String("regions")
	c.MinOnDemandNumber = cli.Int64("min_on_demand_number")
	c.MinOnDemandPercentage = cli.Float64("min_on_demand_percentage")
	c.AllowedInstanceTypes = cli.String("allowed_instance_types")
	c.TagName = cli.String("tag_name")

	data, err := ec2instancesinfo.Data()
	if err != nil {
		log.Fatal(err.Error())
	}
	c.InstanceData = data
}
