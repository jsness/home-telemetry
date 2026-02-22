package collectors

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type lhmNode struct {
	Text     string    `json:"Text"`
	Value    string    `json:"Value"`
	Children []lhmNode `json:"Children"`
}

type lhmRoot struct {
	Children []lhmNode `json:"Children"`
}

func CollectCPUTempFromLHM(url string) (float64, error) {
	if url == "" {
		return 0, errors.New("lhm url empty")
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return 0, errors.New("lhm http status: " + resp.Status)
	}

	var root lhmRoot
	if err := json.NewDecoder(resp.Body).Decode(&root); err != nil {
		return 0, err
	}

	max := 0.0
	found := false
	walk := func(n lhmNode, path []string, visit func(n lhmNode, path []string)) {}
	walk = func(n lhmNode, path []string, visit func(n lhmNode, path []string)) {
		visit(n, path)
		for _, c := range n.Children {
			walk(c, append(path, n.Text), visit)
		}
	}

	visit := func(n lhmNode, path []string) {
		// Look for temperature sensors under CPU
		pathStr := strings.ToLower(strings.Join(append(path, n.Text), "/"))
		if !strings.Contains(pathStr, "cpu") {
			return
		}
		if !strings.Contains(pathStr, "temperatures") {
			return
		}
		if n.Value == "" {
			return
		}
		val := parseTempC(n.Value)
		if val > 0 {
			if !found || val > max {
				max = val
				found = true
			}
		}
	}

	for _, c := range root.Children {
		walk(c, nil, visit)
	}

	if !found {
		return 0, errors.New("cpu temp not found")
	}
	return max, nil
}

func parseTempC(s string) float64 {
	// Value format: "45.0 °C"
	clean := strings.Builder{}
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			clean.WriteRune(r)
		}
	}
	v, _ := strconv.ParseFloat(clean.String(), 64)
	return v
}