package collector

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"

	"github.com/b1naryth1ef/yamon/common"
)

var aptCollector = Simple("apt", func(ctx context.Context, sink common.Sink) error {
	cmd := exec.CommandContext(ctx, "apt", "list", "--upgradable")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return err
	}

	upgradable := 0
	security := 0

	for {
		line, err := stdout.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		parts := strings.Split(line, " ")
		nameRepo := strings.Split(parts[0], "/")
		if len(nameRepo) < 2 {
			continue
		}

		if strings.Contains(nameRepo[1], "-security") {
			security += 1
		} else {
			upgradable += 1
		}
	}

	total := 0
	cmd = exec.CommandContext(ctx, "apt", "list", "--installed")

	var listOut bytes.Buffer
	cmd.Stdout = &listOut
	err = cmd.Run()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(&listOut)
	for scanner.Scan() {
		total += 1
	}

	sink.WriteMetric(common.NewGauge("apt.packages", total-security-upgradable, map[string]string{
		"security":   "false",
		"upgradable": "false",
	}))
	sink.WriteMetric(common.NewGauge("apt.packages", upgradable, map[string]string{
		"security":   "false",
		"upgradable": "true",
	}))
	sink.WriteMetric(common.NewGauge("apt.packages", security, map[string]string{
		"security":   "true",
		"upgradable": "true",
	}))

	return nil
})
