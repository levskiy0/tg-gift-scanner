package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

type GiftInfo struct {
	Model  string
	Rarity string
}

type GiftStats struct {
	ModelCounts map[string]ModelInfo
	TotalGifts  int
	Available   int
	LastUpdate  time.Time
}

type ModelInfo struct {
	Count  int
	Rarity string
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <ModelName> <count> [threads]\n", os.Args[0])
		os.Exit(1)
	}

	modelName := os.Args[1]
	count, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Error: count must be a number\n")
		os.Exit(1)
	}

	threads := 10
	if len(os.Args) > 3 {
		threads, err = strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("Error: threads must be a number\n")
			os.Exit(1)
		}
	}

	baseURL := fmt.Sprintf("https://t.me/nft/%s-", modelName)
	startGift := 1
	endGift := count

	pterm.Println("\033[H\033[2J")

	progressBar, _ := pterm.DefaultProgressbar.WithTotal(endGift - startGift + 1).WithTitle("Scanning gifts").Start()

	stats := scanGifts(baseURL, startGift, endGift, threads, progressBar)

	pterm.Println("\033[H\033[2J")
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("Telegram Gift Scanner - Final Results")
	pterm.Println("")
	pterm.Info.Printf("Checked: %d gifts", count)
	pterm.Println("")
	pterm.Info.Printf("Available: %d gifts", stats.Available)
	pterm.Println("")

	type ModelCount struct {
		Model  string
		Count  int
		Rarity string
	}
	var sortedModels []ModelCount
	for model, info := range stats.ModelCounts {
		sortedModels = append(sortedModels, ModelCount{model, info.Count, info.Rarity})
	}
	sort.Slice(sortedModels, func(i, j int) bool {
		rarityI := parseRarity(sortedModels[i].Rarity)
		rarityJ := parseRarity(sortedModels[j].Rarity)

		if rarityI != rarityJ {
			return rarityI < rarityJ
		}

		return sortedModels[i].Count > sortedModels[j].Count
	})

	tableData := pterm.TableData{
		{"Model", "Count", "Percent", "Rarity"},
	}

	for _, mc := range sortedModels {
		percent := float64(mc.Count) / float64(stats.TotalGifts) * 100
		tableData = append(tableData, []string{
			mc.Model,
			strconv.Itoa(mc.Count),
			fmt.Sprintf("%.1f%%", percent),
			mc.Rarity,
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	pterm.Println()
	pterm.Info.Println("Press Enter to exit...")
	fmt.Scanln()
}

func scanGifts(baseURL string, start, end, threads int, progressBar *pterm.ProgressbarPrinter) GiftStats {
	stats := GiftStats{
		ModelCounts: make(map[string]ModelInfo),
		TotalGifts:  0,
		Available:   0,
		LastUpdate:  time.Now(),
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex
	semaphore := make(chan struct{}, threads)
	progress := make(chan int, 100)

	go func() {
		for range progress {
			progressBar.Increment()
		}
	}()

	for i := start; i <= end; i++ {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(giftNumber int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			url := baseURL + strconv.Itoa(giftNumber)
			info, err := getGiftModel(url)
			if err != nil {
				progress <- 1
				return
			}

			mutex.Lock()
			modelInfo, exists := stats.ModelCounts[info.Model]
			if !exists {
				modelInfo = ModelInfo{
					Count:  0,
					Rarity: info.Rarity,
				}
			}
			modelInfo.Count++
			stats.ModelCounts[info.Model] = modelInfo
			stats.TotalGifts++
			stats.Available++
			mutex.Unlock()
			progress <- 1
		}(i)
	}

	wg.Wait()
	close(progress)
	return stats
}

func getGiftModel(url string) (GiftInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return GiftInfo{}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return GiftInfo{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GiftInfo{}, err
	}

	re := regexp.MustCompile(`<tr><th>Model</th><td>([^<]+?)\s*<mark>([^<]+)</mark>`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 3 {
		return GiftInfo{}, fmt.Errorf("model or rarity not found in response")
	}

	model := strings.TrimSpace(matches[1])
	rarity := strings.TrimSpace(matches[2])

	return GiftInfo{Model: model, Rarity: rarity}, nil
}

func parseRarity(rarity string) float64 {
	rarity = strings.TrimSuffix(rarity, "%")
	value, err := strconv.ParseFloat(rarity, 64)
	if err != nil {
		return 100
	}
	return value
}
