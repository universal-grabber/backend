package main

import (
	pb "backend/gen/proto/service/api"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name: "ugctl",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "context",
				Value: "context",
				Usage: "Context",
			},
		},
		Usage: "Universal Grabber Cli Tool",
		Commands: []*cli.Command{
			{
				Name:    "stats",
				Aliases: []string{"c"},
				Usage:   "show stats for page reference tasks",
				//ArgsUsage:
				//	Action, : func (c *cli.Context) error{
				//	runStats(c)
				//	return nil
				//},
			},
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a task to the list",
				Action: func(c *cli.Context) error {
					log.Print("asd2")
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		return
	}
}

func runStats(c *cli.Context) {
	//apiServices := initApiServices(c)
}

func initApiServices(*cli.Context) ApiServices {
	// initialize grpc
	// Set up a connection to the server.
	conn, err := grpc.Dial("127.0.0.1:6565", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Errorf("did not connect: %v", err)
	}

	return ApiServices{
		pageRefService:      pb.NewPageRefServiceClient(conn),
		pageRefStatsService: pb.NewPageRefStatsServiceClient(conn),
		configService:       pb.NewConfigServiceClient(conn),
	}
}

type ApiServices struct {
	pageRefService      pb.PageRefServiceClient
	pageRefStatsService pb.PageRefStatsServiceClient
	configService       pb.ConfigServiceClient
}
