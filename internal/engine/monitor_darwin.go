//go:build darwin

package engine

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func RunMonitor(ctx context.Context, p *Progress, sampleInterval time.Duration) {
	p.ServerName = discoverAdaptersDarwin()
	sendPhase(p, PhaseConnected)

	prev := map[string]counters{}

	ticker := time.NewTicker(sampleInterval)
	defer ticker.Stop()

	var last time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			if last.IsZero() {
				_, _ = readCountersDarwin(prev)
				last = t
				continue
			}
			elapsed := t.Sub(last).Seconds()
			last = t
			if elapsed <= 0 {
				continue
			}

			total, names := readCountersDarwin(nil)
			if len(names) > 0 {
				p.ServerName = strings.Join(names, ", ")
			}

			var rx, tx uint64
			for iface, c := range total {
				base, ok := prev[iface]
				if !ok {
					prev[iface] = c
					continue
				}
				rx += safeDelta(base.rx, c.rx)
				tx += safeDelta(base.tx, c.tx)
				prev[iface] = c
			}

			_ = sendSample(p, Sample{Phase: PhaseDownload, Rate: float64(rx) / elapsed, At: t})
			_ = sendSample(p, Sample{Phase: PhaseUpload, Rate: float64(tx) / elapsed, At: t})
		}
	}
}

func readCountersDarwin(seen map[string]counters) (map[string]counters, []string) {
	total := map[string]counters{}
	var names []string

	out, err := exec.Command("netstat", "-ib", "-n").Output()
	if err != nil {
		return total, names
	}

	for _, line := range bytes.Split(out, []byte("\n")) {
		s := string(bytes.TrimSpace(line))
		if s == "" || !strings.Contains(s, "<Link#") {
			continue
		}
		fields := strings.Fields(s)
		if len(fields) < 10 {
			continue
		}
		iface := fields[0]
		if iface == "lo0" || strings.HasSuffix(iface, "*") {
			continue
		}

		ibytes, _ := strconv.ParseUint(fields[5], 10, 64)
		obytes, _ := strconv.ParseUint(fields[8], 10, 64)

		c := counters{rx: ibytes, tx: obytes}
		total[iface] = c
		if seen != nil {
			seen[iface] = c
		}
		names = append(names, iface)
	}
	return total, names
}

func discoverAdaptersDarwin() string {
	_, names := readCountersDarwin(nil)
	if len(names) == 0 {
		return "no active interfaces"
	}
	return strings.Join(names, ", ")
}
