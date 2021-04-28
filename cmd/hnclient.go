package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// An HNItem is a Hacker News article
// API definition is here: https://github.com/HackerNews/API
type HNItem struct {
	ID          int    `json:"id"`
	By          string `json:"by"`
	Descendants int    `json:"descendants"`
	Kids        []int  `json:"kids"`
	Score       int    `json:"score"`
	Time        int    `json:"time"`
	Title       string `json:"title"`
	ItemType    string `json:"type"`
	URL         string `json:"url"`
}

type ItemFetcher struct {
	mu     sync.Mutex
	wg     sync.WaitGroup
	output string
}

func query(endpoint string, datatype interface{}) error {
	// log.Printf("opening request to %s endpoint", endpoint)
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://hacker-news.firebaseio.com/v0/%s", endpoint),
		nil,
	)
	if err != nil {
		log.Fatalf("error creating HTTP request: %v", err)
		return err
	}

	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("error sending HTTP request: %v", err)
		return err
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&datatype); err != nil {
		log.Fatalf("error deserializing data")
	}

	if err != nil {
		log.Println("got error:", err)
		return err
	}
	// log.Println("We got the response:", string(responseBytes))
	return nil
}

func getItem(itemId int, fetcher *ItemFetcher) {
	defer fetcher.wg.Done()
	// log.Printf("looping for item %d", itemId)
	endpoint := fmt.Sprintf("item/%d.json", itemId)
	// log.Printf("querying for item %d", top500[i])
	var item HNItem
	err := query(endpoint, &item)
	if err != nil {
		log.Fatalf("error while querying for item %s", err)
	}
	itemRepr := fmt.Sprintf(
		"📝 \033[1m%s\033[0m\n🗣  %s by %s | %d comments | ⭐️ %d\n🔗 %s\n💬 %s\n\n",
		item.Title,
		item.ItemType,
		item.By,
		item.Descendants,
		item.Score,
		item.URL,
		fmt.Sprintf("https://news.ycombinator.com/item?id=%d", item.ID),
	)
	// log.Printf("end querying for item %d", itemId)
	fetcher.mu.Lock()
	fetcher.output += itemRepr
	fetcher.mu.Unlock()
}

func GetMaxItems(args []string) int {
	maxItems := 10
	if len(args) == 1 {
		maxItems, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("argument should be a number got '%s' instead", args[0])
		}
		if maxItems > 500 {
			log.Fatalf("maximum number of stories is 500, got '%d'", maxItems)
		}
	}
	return maxItems
}

func GetStories(storyType string, maxItems int) {
	var top500 []int
	var fetcher ItemFetcher

	// log.Printf("querying for %s stories", stories)
	err := query(fmt.Sprintf("%sstories.json", storyType), &top500)
	if err != nil {
		log.Fatalf("error doing query to topstories: %v", err)
	}
	// log.Println("SUCCESS:", top500)
	// let's parallel stuff!!!
	fetcher.wg.Add(maxItems)
	for i := 0; i < maxItems; i++ {
		go getItem(top500[i], &fetcher)
	}
	fetcher.wg.Wait()
	fmt.Print(fetcher.output)
}
