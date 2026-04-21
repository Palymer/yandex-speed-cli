// yandex-speed-cli — замер скорости через API yandex.ru/internet.
package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	apiIP     = "https://yandex.ru/internet/api/v0/ip"
	apiProbes = "https://yandex.ru/internet/api/v0/get-probes"
)

type probe struct {
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

type probeSection struct {
	Probes []probe `json:"probes"`
}

type config struct {
	Latency  probeSection `json:"latency"`
	Download probeSection `json:"download"`
	Upload   probeSection `json:"upload"`
}

type speedResult struct {
	TotalBytes  int64
	AvgMbps     float64
	PeakMbps    float64
	Elapsed     time.Duration
	InstSamples []float64
	P95InstMbps float64
}

func newHTTPClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 128
	t.MaxIdleConnsPerHost = 128
	t.IdleConnTimeout = 90 * time.Second
	return &http.Client{Transport: t, Timeout: 0}
}

func clientHeaders() http.Header {
	h := make(http.Header)
	h.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	h.Set("Referer", "https://yandex.ru/internet/")
	h.Set("Origin", "https://yandex.ru")
	h.Set("Cache-Control", "no-cache")
	h.Set("Accept", "*/*")
	h.Set("Accept-Encoding", "gzip, deflate, br")
	return h
}

func genRID() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		for i := range b {
			b[i] = alphabet[time.Now().UnixNano()%int64(len(alphabet))]
		}
		return string(b)
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

func withRID(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		if strings.Contains(raw, "?") {
			return raw + "&rid=" + genRID()
		}
		return raw + "?rid=" + genRID()
	}
	q := u.Query()
	q.Set("rid", genRID())
	u.RawQuery = q.Encode()
	return u.String()
}

func decodeBodyReader(res *http.Response) (io.Reader, error) {
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		return gzip.NewReader(res.Body)
	default:
		return res.Body, nil
	}
}

func readIP(ctx context.Context, c *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiIP, nil)
	if err != nil {
		return "", err
	}
	req.Header = clientHeaders()
	res, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ip: status %d", res.StatusCode)
	}
	r, err := decodeBodyReader(res)
	if err != nil {
		return "", err
	}
	if rc, ok := r.(io.ReadCloser); ok && r != res.Body {
		defer rc.Close()
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	var out string
	if err := json.Unmarshal(body, &out); err == nil {
		return strings.Trim(out, `"`), nil
	}
	return strings.TrimSpace(strings.Trim(string(body), `"`)), nil
}

func loadConfig(ctx context.Context, c *http.Client) (*config, error) {
	u, _ := url.Parse(apiProbes)
	q := u.Query()
	q.Set("t", fmt.Sprintf("%d", time.Now().UnixMilli()))
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = clientHeaders()
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get-probes: status %d", res.StatusCode)
	}
	r, err := decodeBodyReader(res)
	if err != nil {
		return nil, err
	}
	if rc, ok := r.(io.ReadCloser); ok && r != res.Body {
		defer rc.Close()
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var cfg config
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return u.Host
}

func pickProbe(probes []probe, hostHint, marker string) *probe {
	for i := range probes {
		p := &probes[i]
		if strings.Contains(p.URL, hostHint) && strings.Contains(p.URL, marker) {
			return p
		}
	}
	for i := range probes {
		p := &probes[i]
		if strings.Contains(p.URL, hostHint) {
			return p
		}
	}
	return nil
}

func pingOnce(ctx context.Context, c *http.Client, target string, timeout time.Duration) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, withRID(target), nil)
	if err != nil {
		return 0, err
	}
	req.Header = clientHeaders()
	t0 := time.Now()
	res, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	_, _ = io.Copy(io.Discard, res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %d", res.StatusCode)
	}
	return time.Since(t0), nil
}

func measureLatency(ctx context.Context, c *http.Client, cfg *config, scanWorkers, pingN int, scanTO, measureTO time.Duration, jsonOut bool, th theme) (host string, samplesMs []float64, failed int, err error) {
	probes := cfg.Latency.Probes
	if len(probes) == 0 {
		return "", nil, 0, fmt.Errorf("нет latency-зондов")
	}

	if !jsonOut {
		spinCtx, cancelSpin := context.WithCancel(ctx)
		spinDone := make(chan struct{})
		go func() {
			defer close(spinDone)
			tick := time.NewTicker(85 * time.Millisecond)
			defer tick.Stop()
			f := 0
			for {
				select {
				case <-spinCtx.Done():
					return
				case <-tick.C:
					f++
					clearLine()
					fmt.Printf("%s%s%s %sсканирование зондов задержки…%s", th.dim, spinner(f), th.rst, th.dim, th.rst)
				}
			}
		}()

		type scanRes struct {
			p  probe
			d  time.Duration
			ok bool
		}
		ch := make(chan scanRes, len(probes))
		var wg sync.WaitGroup
		sem := make(chan struct{}, scanWorkers)
		for _, pr := range probes {
			pr := pr
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				d, e := pingOnce(ctx, c, pr.URL, scanTO)
				ch <- scanRes{p: pr, d: d, ok: e == nil}
			}()
		}
		go func() {
			wg.Wait()
			close(ch)
		}()

		var best probe
		bestD := time.Hour
		for r := range ch {
			if r.ok && r.d < bestD {
				bestD = r.d
				best = r.p
			}
		}
		cancelSpin()
		<-spinDone
		clearLine()

		if best.URL == "" {
			return "", nil, 0, fmt.Errorf("ни один зонд не ответил")
		}
		h := hostOf(best.URL)

		for i := 0; i < pingN; i++ {
			clearLine()
			fmt.Printf("%s%s%s %sзадержка %d/%d к %s%s", th.dim, spinner(i*2), th.rst, th.dim, i+1, pingN, h, th.rst)
			d, e := pingOnce(ctx, c, best.URL, measureTO)
			if e != nil {
				failed++
			} else {
				samplesMs = append(samplesMs, float64(d.Microseconds())/1000.0)
			}
			time.Sleep(30 * time.Millisecond)
		}
		clearLine()
		if len(samplesMs) == 0 {
			return h, nil, failed, fmt.Errorf("нет успешных пингов")
		}
		return h, samplesMs, failed, nil
	}

	// JSON / без TTY — без анимации
	type scanRes struct {
		p  probe
		d  time.Duration
		ok bool
	}
	ch := make(chan scanRes, len(probes))
	var wg sync.WaitGroup
	sem := make(chan struct{}, scanWorkers)
	for _, pr := range probes {
		pr := pr
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			d, e := pingOnce(ctx, c, pr.URL, scanTO)
			ch <- scanRes{p: pr, d: d, ok: e == nil}
		}()
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	var best probe
	bestD := time.Hour
	for r := range ch {
		if r.ok && r.d < bestD {
			bestD = r.d
			best = r.p
		}
	}
	if best.URL == "" {
		return "", nil, 0, fmt.Errorf("ни один зонд не ответил")
	}
	h := hostOf(best.URL)

	for i := 0; i < pingN; i++ {
		d, e := pingOnce(ctx, c, best.URL, measureTO)
		if e != nil {
			failed++
		} else {
			samplesMs = append(samplesMs, float64(d.Microseconds())/1000.0)
		}
		time.Sleep(30 * time.Millisecond)
	}
	if len(samplesMs) == 0 {
		return h, nil, failed, fmt.Errorf("нет успешных пингов")
	}
	return h, samplesMs, failed, nil
}

func meanStdMinMax(xs []float64) (avg, std, minV, maxV float64) {
	if len(xs) == 0 {
		return 0, 0, 0, 0
	}
	minV, maxV = xs[0], xs[0]
	var sum float64
	for _, x := range xs {
		sum += x
		if x < minV {
			minV = x
		}
		if x > maxV {
			maxV = x
		}
	}
	avg = sum / float64(len(xs))
	if len(xs) < 2 {
		return avg, 0, minV, maxV
	}
	var sq float64
	for _, x := range xs {
		d := x - avg
		sq += d * d
	}
	std = sqrt(sq / float64(len(xs)-1))
	return avg, std, minV, maxV
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func percentile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]float64(nil), xs...)
	sort.Float64s(cp)
	idx := int(float64(len(cp)-1) * p / 100.0)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}

func runDownload(parent context.Context, c *http.Client, rawURL string, workers int, duration time.Duration,
	jsonOut bool, th theme, label string) speedResult {
	ctx, cancel := context.WithTimeout(parent, duration)
	defer cancel()

	var bytes atomic.Int64
	var wg sync.WaitGroup
	buf := make([]byte, 128*1024)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, withRID(rawURL), nil)
				if err != nil {
					continue
				}
				req.Header = clientHeaders()
				res, err := c.Do(req)
				if err != nil {
					time.Sleep(50 * time.Millisecond)
					continue
				}
				for ctx.Err() == nil {
					n, err := res.Body.Read(buf)
					if n > 0 {
						bytes.Add(int64(n))
					}
					if err != nil {
						res.Body.Close()
						break
					}
				}
			}
		}()
	}

	t0 := time.Now()
	var peakWindow float64
	var lastB int64
	lastT := t0
	var samples []float64
	tick := time.NewTicker(90 * time.Millisecond)
	defer tick.Stop()
	frame := 0
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-tick.C:
			frame++
			now := time.Now()
			elapsed := now.Sub(t0)
			b := bytes.Load()
			dt := now.Sub(lastT).Seconds()
			var inst float64
			if dt > 0.08 {
				inst = float64(b-lastB) * 8 / 1e6 / dt
				samples = append(samples, inst)
				if inst > peakWindow {
					peakWindow = inst
				}
			}
			lastB, lastT = b, now
			sec := elapsed.Seconds()
			var avg float64
			if sec > 0.001 {
				avg = float64(b) * 8 / 1e6 / sec
			}
			if !jsonOut {
				ratio := elapsed.Seconds() / duration.Seconds()
				bar := progressBar(ratio, 12)
				printSpeedFrame(spinner(frame), label, bar, elapsed, duration, inst, avg, peakWindow, b, workers)
			}
		}
	}
	cancel()
	wg.Wait()
	clearLine()
	elapsed := time.Since(t0)
	total := bytes.Load()
	sec := elapsed.Seconds()
	var avgMbps float64
	if sec > 0 {
		avgMbps = float64(total) * 8 / 1e6 / sec
	}
	p95 := percentile(samples, 95)
	return speedResult{
		TotalBytes: total, AvgMbps: avgMbps, PeakMbps: peakWindow,
		Elapsed: elapsed, InstSamples: samples, P95InstMbps: p95,
	}
}

type countingReader struct {
	buf []byte
	off int
	n   *atomic.Int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	written := 0
	for written < len(p) {
		avail := len(r.buf) - r.off
		if avail == 0 {
			r.off = 0
			avail = len(r.buf)
		}
		need := len(p) - written
		take := avail
		if take > need {
			take = need
		}
		copy(p[written:], r.buf[r.off:r.off+take])
		r.off += take
		written += take
		r.n.Add(int64(take))
	}
	return written, nil
}

func runUpload(parent context.Context, c *http.Client, rawURL string, workers int, duration time.Duration, chunk int,
	jsonOut bool, th theme, label string) speedResult {
	ctx, cancel := context.WithTimeout(parent, duration)
	defer cancel()

	buf := make([]byte, chunk)
	if _, err := rand.Read(buf); err != nil {
		for i := range buf {
			buf[i] = byte(i % 251)
		}
	}
	var sent atomic.Int64
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				body := &countingReader{buf: buf, n: &sent}
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, withRID(rawURL), body)
				if err != nil {
					continue
				}
				req.Header = clientHeaders()
				req.Header.Set("Content-Type", "application/octet-stream")
				res, err := c.Do(req)
				if err != nil {
					time.Sleep(50 * time.Millisecond)
					continue
				}
				_, _ = io.Copy(io.Discard, res.Body)
				res.Body.Close()
			}
		}()
	}

	t0 := time.Now()
	var peakWindow float64
	var lastB int64
	lastT := t0
	var samples []float64
	tick := time.NewTicker(90 * time.Millisecond)
	defer tick.Stop()
	frame := 0
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-tick.C:
			frame++
			now := time.Now()
			elapsed := now.Sub(t0)
			b := sent.Load()
			dt := now.Sub(lastT).Seconds()
			var inst float64
			if dt > 0.08 {
				inst = float64(b-lastB) * 8 / 1e6 / dt
				samples = append(samples, inst)
				if inst > peakWindow {
					peakWindow = inst
				}
			}
			lastB, lastT = b, now
			sec := elapsed.Seconds()
			var avg float64
			if sec > 0.001 {
				avg = float64(b) * 8 / 1e6 / sec
			}
			if !jsonOut {
				ratio := elapsed.Seconds() / duration.Seconds()
				bar := progressBar(ratio, 12)
				printSpeedFrame(spinner(frame), label, bar, elapsed, duration, inst, avg, peakWindow, b, workers)
			}
		}
	}
	cancel()
	wg.Wait()
	clearLine()
	elapsed := time.Since(t0)
	total := sent.Load()
	sec := elapsed.Seconds()
	var avgMbps float64
	if sec > 0 {
		avgMbps = float64(total) * 8 / 1e6 / sec
	}
	p95 := percentile(samples, 95)
	return speedResult{
		TotalBytes: total, AvgMbps: avgMbps, PeakMbps: peakWindow,
		Elapsed: elapsed, InstSamples: samples, P95InstMbps: p95,
	}
}

const summaryLabelColRunes = 28

func printSpeedSummary(th theme, title string, r speedResult) {
	p50 := percentile(r.InstSamples, 50)
	n := len(r.InstSamples)
	w := summaryLabelColRunes
	fmt.Printf("%s✓ %s%s\n", th.grn+th.bold, title, th.rst)
	fmt.Printf("  %s%s%s  %.2f Мбит/с\n", th.dim, labelCol("Средняя скорость:", w), th.rst, r.AvgMbps)
	fmt.Printf("  %s%s%s  %.2f Мбит/с\n", th.dim, labelCol("Пик за окно замера:", w), th.rst, r.PeakMbps)
	fmt.Printf("  %s%s%s  %.2f / %.2f Мбит/с\n", th.dim, labelCol("Мгновенная p95 / p50:", w), th.rst, r.P95InstMbps, p50)
	fmt.Printf("  %s%s%s  %.2f Мбит (%.2f МиБ) за %.2f с\n", th.dim, labelCol("Передано:", w), th.rst,
		float64(r.TotalBytes)*8/1e6, float64(r.TotalBytes)/(1024*1024), r.Elapsed.Seconds())
	fmt.Printf("  %s%s%s  %d\n", th.dim, labelCol("Точек измерения:", w), th.rst, n)
	fmt.Println()
}

func waitEnterPrompt() {
	fmt.Println()
	fmt.Print("Нажмите Enter для выхода… ")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func main() {
	var (
		showVersion = flag.Bool("version", false, "показать версию и выйти")
		quick       = flag.Bool("quick", false, "короткий тест")
		durationSec = flag.Float64("duration", 10, "длительность DL/UL (сек)")
		workers     = flag.Int("workers", 4, "параллельных потоков")
		pingN       = flag.Int("ping", 10, "число пингов после выбора узла")
		noDL        = flag.Bool("no-download", false, "пропустить download")
		noUL        = flag.Bool("no-upload", false, "пропустить upload")
		jsonOut     = flag.Bool("json", false, "вывод в JSON")
		noColor     = flag.Bool("no-color", false, "без ANSI-цветов")
		noWait      = flag.Bool("no-wait", false, "не ждать Enter после теста (скрипты, CI)")
		noGeo       = flag.Bool("no-geo", false, "не запрашивать геолокацию по IP (ipwho.is)")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	dur := time.Duration(*durationSec * float64(time.Second))
	w := *workers
	pn := *pingN
	scanWorkers := 16
	scanTO := 1200 * time.Millisecond
	measureTO := 2500 * time.Millisecond
	if *quick {
		dur = 4 * time.Second
		w = 3
		pn = 5
		scanTO = 800 * time.Millisecond
		measureTO = 2000 * time.Millisecond
	}
	if w < 1 {
		w = 1
	}

	th := newTheme(*noColor)
	ctx := context.Background()
	c := newHTTPClient()

	if !*jsonOut {
		tryEnableVirtualTerminal()
		printBanner(th)
	}

	ip, err := readIP(ctx, c)
	if err != nil && !*jsonOut {
		fmt.Fprintf(os.Stderr, "%sIP:%s ошибка: %v\n", th.red, th.rst, err)
	}
	sessionTime := time.Now()
	ipv4, ipv6 := classifyPublicIP(ip)
	var geo *geoLookup
	if !*noGeo && strings.TrimSpace(ip) != "" {
		geo = lookupGeoIP(ctx, c, ip)
	}
	regionByIP := regionByIPLabel(geo)
	regionName := regionLabel(geo)
	if !*jsonOut {
		printInfoPanel(th, sessionTime, ipv4, ipv6, regionByIP, regionName)
	}

	cfg, err := loadConfig(ctx, c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "конфиг: %v\n", err)
		os.Exit(1)
	}

	host, samples, fail, err := measureLatency(ctx, c, cfg, scanWorkers, pn, scanTO, measureTO, *jsonOut, th)
	if err != nil {
		fmt.Fprintf(os.Stderr, "latency: %v\n", err)
		os.Exit(1)
	}
	avg, jitter, minV, maxV := meanStdMinMax(samples)
	p95ms := percentile(samples, 95)

	if !*jsonOut {
		// «узел» с отступом 2 пробела: 2 + подпись(24) + 2 = до значения как у панели (26+2)
		const subLabelW = 24
		fmt.Printf("%s● Задержка%s  %.1f мс  (min %.0f · max %.0f · σ %.1f · p95 %.0f мс)\n",
			th.cya, th.rst, avg, minV, maxV, jitter, p95ms)
		fmt.Printf("  %s%s%s  %s%s%s\n", th.dim, labelCol("узел:", subLabelW), th.rst, th.dim, host, th.rst)
		if fail > 0 {
			fmt.Printf("  %s%s%s  %d запросов\n", th.yel, labelCol("потери:", subLabelW), th.rst, fail)
		}
		fmt.Println()
	}

	type jsonSpeed struct {
		AvgMbps     float64 `json:"avg_mbps"`
		PeakMbps    float64 `json:"peak_mbps"`
		P95InstMbps float64 `json:"p95_instant_mbps"`
		TotalByte   int64   `json:"total_bytes"`
		DurationS   float64 `json:"duration_s"`
		SampleCount int     `json:"instant_sample_count"`
		InstMbpsP50 float64 `json:"instant_mbps_p50,omitempty"`
	}

	type jsonSession struct {
		DateLocal      string `json:"date_local"`
		TimeLocal      string `json:"time_local"`
		TimezoneOffset string `json:"timezone_utc_offset"`
		IPv4           string `json:"ipv4"`
		IPv6           string `json:"ipv6"`
		RegionByIP     string `json:"region_by_ip"`
		Region         string `json:"region"`
		GeoSource      string `json:"geo_source,omitempty"`
		GeoError       string `json:"geo_error,omitempty"`
	}
	js := jsonSession{
		DateLocal:      sessionTime.Format("02.01.2006"),
		TimeLocal:      sessionTime.Format("15:04"),
		TimezoneOffset: formatUTCOffset(sessionTime),
		IPv4:           ipv4,
		IPv6:           ipv6,
		RegionByIP:     regionByIP,
		Region:         regionName,
	}
	if *noGeo {
		js.GeoError = "отключено (-no-geo)"
	} else if geo != nil && !geo.OK && geo.Err != "" {
		js.GeoError = geo.Err
	} else if !*noGeo && strings.TrimSpace(ip) != "" {
		js.GeoSource = "ipwho.is"
	}

	out := struct {
		App      string      `json:"app"`
		Version  string      `json:"version,omitempty"`
		IP       string      `json:"ip"`
		Session  jsonSession `json:"session"`
		Latency  any         `json:"latency,omitempty"`
		Download *jsonSpeed  `json:"download,omitempty"`
		Upload   *jsonSpeed  `json:"upload,omitempty"`
	}{
		App:     appName,
		Version: version,
		IP:      ip,
		Session: js,
		Latency: struct {
			Host      string    `json:"host"`
			AvgMs     float64   `json:"avg_ms"`
			JitterMs  float64   `json:"jitter_ms"`
			MinMs     float64   `json:"min_ms"`
			MaxMs     float64   `json:"max_ms"`
			P95Ms     float64   `json:"p95_ms"`
			Samples   []float64 `json:"samples_ms"`
			FailedReq int       `json:"failed_requests"`
		}{host, avg, jitter, minV, maxV, p95ms, samples, fail},
	}

	if !*noDL {
		p := pickProbe(cfg.Download.Probes, host, "50mb")
		if p != nil {
			if !*jsonOut {
				fmt.Printf("%s▼ Входящий канал%s (%s)\n", th.mag+th.bold, th.rst, fmtDur(dur))
			}
			r := runDownload(ctx, c, p.URL, w, dur, *jsonOut, th, "↓")
			if !*jsonOut {
				printSpeedSummary(th, "Входящий канал (download)", r)
			}
			p50 := percentile(r.InstSamples, 50)
			out.Download = &jsonSpeed{
				AvgMbps: r.AvgMbps, PeakMbps: r.PeakMbps, P95InstMbps: r.P95InstMbps,
				TotalByte: r.TotalBytes, DurationS: r.Elapsed.Seconds(),
				SampleCount: len(r.InstSamples), InstMbpsP50: p50,
			}
		}
	}

	if !*noUL {
		p := pickProbe(cfg.Upload.Probes, host, "52428800")
		if p == nil {
			p = pickProbe(cfg.Upload.Probes, host, "")
		}
		if p != nil {
			if !*jsonOut {
				fmt.Printf("%s▼ Исходящий канал%s (%s)\n", th.mag+th.bold, th.rst, fmtDur(dur))
			}
			r := runUpload(ctx, c, p.URL, w, dur, 1024*1024, *jsonOut, th, "↑")
			if !*jsonOut {
				printSpeedSummary(th, "Исходящий канал (upload)", r)
			}
			p50 := percentile(r.InstSamples, 50)
			out.Upload = &jsonSpeed{
				AvgMbps: r.AvgMbps, PeakMbps: r.PeakMbps, P95InstMbps: r.P95InstMbps,
				TotalByte: r.TotalBytes, DurationS: r.Elapsed.Seconds(),
				SampleCount: len(r.InstSamples), InstMbpsP50: p50,
			}
		}
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
	} else if !*noWait {
		waitEnterPrompt()
	}
}
