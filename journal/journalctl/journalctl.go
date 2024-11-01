package journalctl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type Entry map[string]string

func (e Entry) RealtimeTimestamp() time.Time {
	realtimeTimestamp, _ := strconv.Atoi(e["__REALTIME_TIMESTAMP"])
	return time.Unix(int64(realtimeTimestamp/1000000), 0)
}

type Instance struct {
	cmd     *exec.Cmd
	entries chan Entry
}

type Opts struct {
	Output string
	Follow bool
	Lines  *int

	OnInvalidJSON func(data []byte, err error)
}

func New(opts *Opts) (*Instance, error) {
	cmd := exec.Command("journalctl")

	if opts.Output != "" {
		cmd.Args = append(cmd.Args, "--output", opts.Output)
	}

	if opts.Follow {
		cmd.Args = append(cmd.Args, "--follow")
	}

	if opts.Lines != nil {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-n%d", *opts.Lines))
	}

	output, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	entries := make(chan Entry, 0)

	go func() {
		scanner := bufio.NewScanner(output)

		for scanner.Scan() {
			line := scanner.Bytes()

			entry := map[string]string{}
			err := json.Unmarshal(line, &entry)
			if err != nil {
				if opts.OnInvalidJSON != nil {
					opts.OnInvalidJSON(line, err)
				}
				continue
			}

			entries <- entry
		}
	}()

	return &Instance{
		cmd:     cmd,
		entries: entries,
	}, nil
}

func (i *Instance) Entries() <-chan Entry {
	return i.entries
}

func (i *Instance) Close() error {
	return i.cmd.Process.Kill()
}
