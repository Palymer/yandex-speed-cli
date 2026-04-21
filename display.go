package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

const appName = "yandex-speed-cli"

// Версия подставляется при сборке: -ldflags="-X main.version=1.2.3"
var version = "dev"

func stdoutIsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

type theme struct {
	bold, dim, cya, grn, yel, mag, red, rst string
}

func newTheme(noColor bool) theme {
	if noColor || !stdoutIsTTY() {
		return theme{}
	}
	return theme{
		bold: "\033[1m",
		dim:  "\033[2m",
		cya:  "\033[36m",
		grn:  "\033[32m",
		yel:  "\033[33m",
		mag:  "\033[35m",
		red:  "\033[31m",
		rst:  "\033[0m",
	}
}

var spinnerRunes = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

func spinner(frame int) string {
	if len(spinnerRunes) == 0 {
		return "*"
	}
	return string(spinnerRunes[frame%len(spinnerRunes)])
}

// progressBar — заполнение по доле (одна строка при обновлении \r).
func progressBar(ratio float64, width int) string {
	if width < 4 {
		width = 4
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	var b strings.Builder
	b.Grow(width + 2)
	b.WriteByte('[')
	for i := 0; i < width; i++ {
		if i < filled {
			if filled > 0 && i == filled-1 {
				b.WriteRune('▶')
			} else {
				b.WriteRune('█')
			}
		} else {
			b.WriteRune('░')
		}
	}
	b.WriteByte(']')
	return b.String()
}

func fmtDur(d time.Duration) string {
	d = d.Round(100 * time.Millisecond)
	if d < 0 {
		d = 0
	}
	s := d.Seconds()
	return fmt.Sprintf("%.1fs", s)
}

func clearLine() {
	if stdoutIsTTY() {
		fmt.Print("\033[2K\r")
	} else {
		fmt.Print("\r")
	}
}

func flushStdout() {
	_ = os.Stdout.Sync()
}

// padDisplayWidth дополняет пробелами по числу **рун** (видимая ширина в консоли).
func padDisplayWidth(s string, runeWidth int) string {
	n := utf8.RuneCountInString(s)
	if n >= runeWidth {
		return s
	}
	return s + strings.Repeat(" ", runeWidth-n)
}

// truncateRunes обрезает строку до runeWidth рун (для ровной колонки подписей).
func truncateRunes(s string, runeWidth int) string {
	if runeWidth <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= runeWidth {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= runeWidth {
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

// labelCol подпись фиксированной ширины (дополняем или обрезаем).
func labelCol(s string, runeWidth int) string {
	n := utf8.RuneCountInString(s)
	if n > runeWidth {
		return truncateRunes(s, runeWidth)
	}
	return s + strings.Repeat(" ", runeWidth-n)
}

// printSpeedFrame — одна короткая строка без ANSI (чтобы не раздувать ширину и не ломать перенос в Windows).
// Обновление: очистка строки + \r + текст + паддинг + сброс буфера.
func printSpeedFrame(spin, label, bar string, elapsed, totalDur time.Duration,
	instMbps, avgMbps, peakMbps float64, totalBytes int64, workers int) {
	remain := totalDur - elapsed
	if remain < 0 {
		remain = 0
	}
	mb := float64(totalBytes) / (1024 * 1024)

	var b strings.Builder
	b.Grow(120)
	b.WriteString(spin)
	b.WriteByte(' ')
	b.WriteString(label)
	b.WriteByte(' ')
	b.WriteString(bar)
	fmt.Fprintf(&b, " %.1f/%.1fс ост%.1fс |", elapsed.Seconds(), totalDur.Seconds(), remain.Seconds())
	fmt.Fprintf(&b, " сейчас %.0f", instMbps)
	fmt.Fprintf(&b, " средн %.0f", avgMbps)
	fmt.Fprintf(&b, " пик %.0f Мбит/с", peakMbps)
	fmt.Fprintf(&b, " | %.0f МиБ", mb)
	fmt.Fprintf(&b, " | %d п.", workers)

	line := padDisplayWidth(b.String(), 96)
	clearLine()
	fmt.Print(line)
	flushStdout()
}

// formatUTCOffset — смещение локальной зоны относительно UTC, вид «UTC+03:00».
func formatUTCOffset(t time.Time) string {
	_, off := t.Zone()
	sign := '+'
	if off < 0 {
		sign = '-'
		off = -off
	}
	h := off / 3600
	m := (off % 3600) / 60
	return fmt.Sprintf("UTC%c%02d:%02d", sign, h, m)
}

const infoLabelColRunes = 26 // ширина колонки подписей

// printInfoPanel — дата/время/пояс и адреса, значения в одном столбце.
func printInfoPanel(th theme, now time.Time, ipv4, ipv6, regionByIP, region string) {
	w := infoLabelColRunes
	fmt.Printf("%s%s%s  %s%s%s\n", th.dim, labelCol("Дата:", w), th.rst, th.bold, now.Format("02.01.2006"), th.rst)
	fmt.Printf("%s%s%s  %s%s%s\n", th.dim, labelCol("Время:", w), th.rst, th.bold, now.Format("15:04"), th.rst)
	fmt.Printf("%s%s%s  %s%s%s\n", th.dim, labelCol("Часовой пояс:", w), th.rst, th.bold, formatUTCOffset(now), th.rst)
	fmt.Printf("%s%s%s  %s%s%s\n", th.dim, labelCol("IPv4-адрес:", w), th.rst, th.bold, ipv4, th.rst)
	fmt.Printf("%s%s%s  %s%s%s\n", th.dim, labelCol("IPv6-адрес:", w), th.rst, th.bold, ipv6, th.rst)
	fmt.Printf("%s%s%s  %s%s%s\n", th.dim, labelCol("Регион по IP-адресу:", w), th.rst, th.bold, regionByIP, th.rst)
	fmt.Printf("%s%s%s  %s%s%s\n", th.dim, labelCol("Регион:", w), th.rst, th.bold, region, th.rst)
	fmt.Println()
}

func printBanner(t theme) {
	title := appName
	sub := "Яндекс.Интернетометр · замер скорости"
	line := strings.Repeat("═", 56)
	fmt.Println()
	fmt.Printf("%s%s%s\n", t.cya, line, t.rst)
	fmt.Printf("  %s%s%s  %s%s%s\n", t.bold, title, t.rst, t.dim, sub, t.rst)
	if version != "" && version != "dev" {
		fmt.Printf("  %sверсия %s%s\n", t.dim, version, t.rst)
	}
	fmt.Printf("%s%s%s\n\n", t.cya, line, t.rst)
}
