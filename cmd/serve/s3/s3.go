package s3

import (
	"context"
	"strings"

	"github.com/rclone/rclone/cmd"
	"github.com/rclone/rclone/fs/config/flags"
	"github.com/rclone/rclone/fs/hash"
	httplib "github.com/rclone/rclone/lib/http"
	"github.com/rclone/rclone/vfs"
	"github.com/rclone/rclone/vfs/vfsflags"
	"github.com/spf13/cobra"
)

// DefaultOpt is the default values used for Options
var DefaultOpt = Options{
	pathBucketMode: true,
	hashName:       "MD5",
	hashType:       hash.MD5,
	noCleanup:      false,
	HTTP:           httplib.DefaultCfg(),
}

// Opt is options set by command line flags
var Opt = DefaultOpt

const flagPrefix = ""

func init() {
	flagSet := Command.Flags()
	httplib.AddHTTPFlagsPrefix(flagSet, flagPrefix, &Opt.HTTP)
	vfsflags.AddFlags(flagSet)
	flags.BoolVarP(flagSet, &Opt.pathBucketMode, "force-path-style", "", Opt.pathBucketMode, "If true use path style access if false use virtual hosted style (default true)")
	flags.StringVarP(flagSet, &Opt.hashName, "etag-hash", "", Opt.hashName, "Which hash to use for the ETag, or auto or blank for off")
	flags.StringArrayVarP(flagSet, &Opt.authPair, "s3-authkey", "", Opt.authPair, "Set key pair for v4 authorization, split by comma")
	flags.BoolVarP(flagSet, &Opt.noCleanup, "no-cleanup", "", Opt.noCleanup, "Not to cleanup empty folder after object is deleted")
}

// Command definition for cobra
var Command = &cobra.Command{
	Use:   "s3 remote:path",
	Short: `Serve remote:path over s3.`,
	Long:  strings.ReplaceAll(longHelp, "|", "`") + httplib.Help(flagPrefix) + vfs.Help,
	RunE: func(command *cobra.Command, args []string) error {
		cmd.CheckArgs(1, 1, command, args)
		f := cmd.NewFsSrc(args)

		if Opt.hashName == "auto" {
			Opt.hashType = f.Hashes().GetOne()
		} else if Opt.hashName != "" {
			err := Opt.hashType.Set(Opt.hashName)
			if err != nil {
				return err
			}
		}
		cmd.Run(false, false, command, func() error {
			s, err := newServer(context.Background(), f, &Opt)
			if err != nil {
				return err
			}
			router := s.Router()
			s.Bind(router)
			s.Serve()
			s.Wait()
			return nil
		})
		return nil
	},
}
