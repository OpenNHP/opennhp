package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// renderCSV writes one flat table: section,key,count,share,extra.
func renderCSV(w io.Writer, r *Report) error {
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"section", "key", "pkts_exact", "sampled_events", "share", "extra"})
	pct := func(f float64) string { return fmt.Sprintf("%.4f", f) }
	for _, x := range r.Reasons {
		_ = cw.Write([]string{"reason", x.Reason, fmt.Sprint(x.Pkts), fmt.Sprint(x.Events), pct(x.PktShare), x.Category})
	}
	for _, x := range r.Categories {
		_ = cw.Write([]string{"category", x.Category, fmt.Sprint(x.Pkts), fmt.Sprint(x.Events), pct(x.Share), ""})
	}
	for _, x := range r.TopIPs {
		extra := fmt.Sprintf("%s %s; categories=%d; ports=%v", x.Country, x.City, x.Categories, x.TopPorts)
		_ = cw.Write([]string{"top_ip", x.IP, "", fmt.Sprint(x.Events), pct(x.Share), strings.TrimSpace(extra)})
	}
	for _, x := range r.TopGeo {
		_ = cw.Write([]string{"top_geo", x.Country, "", fmt.Sprint(x.Events), pct(x.Share), fmt.Sprintf("unique_ips=%d", x.IPs)})
	}
	for _, x := range r.TopPorts {
		_ = cw.Write([]string{"top_port", x.Key, "", fmt.Sprint(x.Count), pct(x.Share), ""})
	}
	for _, x := range r.Protos {
		_ = cw.Write([]string{"proto", x.Key, "", fmt.Sprint(x.Count), pct(x.Share), ""})
	}
	for _, x := range r.Hourly {
		_ = cw.Write([]string{"hourly", x.Key, fmt.Sprint(x.Count), "", pct(x.Share), ""})
	}
	for _, x := range r.Daily {
		_ = cw.Write([]string{"daily", x.Key, fmt.Sprint(x.Count), "", pct(x.Share), ""})
	}
	cw.Flush()
	return cw.Error()
}
