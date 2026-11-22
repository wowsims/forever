package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/wowsims/forever/sim/selfupdate"
)

func newUpdateCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:          "update",
		Short:        "updates wowsimcli to the newest release",
		Long:         "updates wowsimcli to the newest release",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			executable, err := os.Executable()
			if err != nil {
				return err
			}
			source, err := selfupdate.GitHubSource()
			if err != nil {
				return err
			}
			updater, err := selfupdate.New(selfupdate.Config{Name: "wowsimcli", Version: version, Executable: executable, Source: source})
			if err != nil {
				return err
			}
			updater.RemoveOld()
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()
			rel, err := updater.Check(ctx)
			if errors.Is(err, selfupdate.ErrDevelopmentBuild) {
				return errors.New("this is a development build, which does not update itself; download a release instead")
			}
			if err != nil {
				return fmt.Errorf("checking for a newer release: %w", err)
			}
			if rel == nil {
				fmt.Printf("wowsimcli %s is the newest release.\n", version)
				return nil
			}
			fmt.Printf("Updating wowsimcli %s -> %s...\n", version, rel.Tag)
			if err := updater.Apply(ctx, rel); err != nil {
				if errors.Is(err, selfupdate.ErrNotWritable) {
					return fmt.Errorf("%w; download %s from %s instead", err, rel.Tag, rel.URL)
				}
				return err
			}
			fmt.Printf("Updated wowsimcli to %s.\n", rel.Tag)
			return nil
		},
	}
}
