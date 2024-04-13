//go:build !(linux && amd64)

package journal

import "github.com/b1naryth1ef/yamon/common"

func Run(config *common.DaemonJournalConfig, sink common.Sink) error {
	return nil
}
