package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// Timeout for scraping the favorite music
const TIMEOUT = 60 * time.Second

type ScraperState uint8

const (
	LoadingChannel        ScraperState = 0
	DenyingCookies        ScraperState = 1
	ScrapingFavoriteMusic ScraperState = 2
)

func main() {
	fmt.Println("Please make sure that \"Enable public stats\" is enabled in your youtube music channel settings.")
	fmt.Printf("Currently fetching the favorite music, this might take a bit long... (Timeout of %v)\n", TIMEOUT)
	name, link, author, err := GetFavoriteFromChannelId("")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(name, link, author)
}

// Get the first favorite music from that youtube music channel
// Expects that "Enable public stats" is enabled for the youtube channel, otherwise it won't work and will hit the timeout
func GetFavoriteFromChannelId(channel_id string) (name string, music_link string, author string, err error) {
	// Set language to english since we expect to get the english youtube music
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("lang", "en"),
		chromedp.Env("LANG=en"),
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0"),
	)

	// Create a context with a timeout of 10 seconds
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()
	bctx, bcancel := chromedp.NewContext(actx)
	defer bcancel()
	ctx, cancel := context.WithTimeout(bctx, TIMEOUT)
	defer cancel()

	// Set the ok value to true to prevent the href error from overwriting the real one
	href_ok := true
	scraper_state := LoadingChannel
	err = chromedp.Run(ctx,
		// Reject all cookies when going to the website
		chromedp.Navigate("https://music.youtube.com/channel/"+channel_id),
		chromedp.ActionFunc(func(ctx context.Context) error {
			scraper_state = DenyingCookies
			return nil
		}),
		chromedp.Click(`[aria-label="Reject all"]`, chromedp.ByQuery),

		// Get the name, link, and author of the first favorite music
		chromedp.ActionFunc(func(ctx context.Context) error {
			scraper_state = ScrapingFavoriteMusic
			return nil
		}),
		chromedp.Text(`div#contents.style-scope.ytmusic-shelf-renderer a`, &name, chromedp.NodeVisible),
		chromedp.AttributeValue(`div#contents.style-scope.ytmusic-shelf-renderer a`, "href", &music_link, &href_ok, chromedp.NodeVisible),
		chromedp.Text(`div#contents.style-scope.ytmusic-shelf-renderer .flex-column a`, &author, chromedp.NodeVisible),
	)

	// If the deadline has been reached, then print out a message explaining at what stage did it fail with a guideline
	if errors.Is(err, context.DeadlineExceeded) {
		switch scraper_state {
		case LoadingChannel:
			log.Print("Timeout triggered: Loading the channel took too long. Is youtube music even accessible? Or your internet speed too slow?")
		case DenyingCookies:
			log.Print("Timeout triggered: Coudln't click on \"Reject all\" in the cookie banner. Maybe youtube music updated their website and favorite_music_badge needs to be updated for the new website.")
		case ScrapingFavoriteMusic:
			log.Print("Timeout triggered: While getting the favorite music from the website, the website has finished loading. Did you enable \"Enable public stats\" in your youtube music channel settings?")
		}
		log.Print("Nonetheless, please retry running this script before reporting this as a bug if this is not a problem on your side.")
	}

	if href_ok == false {
		err = errors.New("Attribute 'href' not found in the \"a\" tag.")
	}

	music_link = "https://youtube.com/" + music_link
	return
}
