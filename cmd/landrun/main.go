package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/armin/landrun/internal/sandbox"
	"github.com/armin/landrun/internal/exec"
)

func main() {
	app := &cli.App{
		Name:  "landrun",
		Usage: "Run a command in a Landlock sandbox",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "ro",
				Usage: "Allow read-only access to this path",
			},
			&cli.StringSliceFlag{
				Name:  "rw",
				Usage: "Allow read-write access to this path",
			},
			&cli.BoolFlag{
				Name:  "exec",
				Usage: "Allow executing files in allowed paths",
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()
			if len(args) == 0 {
				log.Fatal("Missing command to run")
			}

			cfg := sandbox.Config{
				ReadOnlyPaths:  c.StringSlice("ro"),
				ReadWritePaths: c.StringSlice("rw"),
				AllowExec:      c.Bool("exec"),
			}

			if err := sandbox.Apply(cfg); err != nil {
				log.Fatalf("Failed to apply sandbox: %v", err)
			}

			return exec.Run(args)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
