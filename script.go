package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// Timeout for scraping the favorite music
const VERSION = "v1.0.0"
const REPOSITORY_DIR = "./repository_to_modify/"

type ScraperState uint8

const (
	LoadingChannel        ScraperState = 0
	DenyingCookies        ScraperState = 1
	ScrapingFavoriteMusic ScraperState = 2
)

func main() {
	channel_id, user_agent, timeout, message_color, style, logo, logoColor, logoSize, labelColor, color, cacheSeconds, repository, filename, err := parseCommandLineArgs()
	if err != nil {
		log.Fatal(err)
	}

	// Fetch the favorite music
	fmt.Println("Please make sure that \"Enable public stats\" is enabled in your youtube music channel settings.")
	fmt.Printf("Currently fetching the favorite music, this might take a bit long... (Timeout of %v)\n", timeout)
	name, video_link, author, err := GetFavoriteFromChannelId(channel_id, *user_agent, timeout)
	if err != nil {
		log.Fatal(err)
	}

	// Create a link of it as an image
	image_link := Generate_image_link(name, author, *message_color, *style, *logo, *logoColor, *logoSize, *labelColor, *color, *cacheSeconds)
	fmt.Printf("Favorite music: %v by %v (%v)\n", name, author, video_link)
	fmt.Println(image_link)

	if *repository != "" {
		fmt.Println("The image link has been generated we are now downloading the repository and adding the favorite_music_badge to it!")

		// Clone the repository
		clone := exec.Command("git", "clone", *repository, REPOSITORY_DIR)
		clone.Stderr = os.Stderr

		output, err := clone.Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(output))

		// Search the file and add the youtube music badge
		file, err := os.Open(REPOSITORY_DIR + *filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		var lines []string

		// Did we add a music badge at least once?
		added_youtube_music_badge := false

		// Should we add a youtube music badge now
		add_youtube_music_badge := false
		for scanner.Scan() {
			line := scanner.Text()

			if add_youtube_music_badge {
				add_youtube_music_badge = false
				added_youtube_music_badge = true
				line = fmt.Sprintf("[<img src=\"%v\"/>](%v)", image_link, video_link)
			} else if strings.Contains(line, "FAVORITE_MUSIC_BADGE_AFTER_THIS_LINE") {
				add_youtube_music_badge = true
			}


			lines = append(lines, line)
		}

		if add_youtube_music_badge {
			add_youtube_music_badge = false
			added_youtube_music_badge = true
			lines = append(lines, fmt.Sprintf("[<img src=\"%v\"/>](%v)", image_link, video_link))
		}

		if added_youtube_music_badge == false {
			log.Fatal("Tried to add a favorite music badge without a FAVORITE_MUSIC_BADGE_AFTER_THIS_LINE inside of the readme")
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		// Overwrite the existing file
		outputFile, err := os.Create(REPOSITORY_DIR + *filename)
		if err != nil {
			log.Fatal(err)
		}
		defer outputFile.Close()

		writer := bufio.NewWriter(outputFile)
		for _, line := range lines {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				log.Fatal(err)
			}
		}
		writer.Flush()

		// Try to do a git add the modified file
		add := exec.Command("git", "--git-dir", REPOSITORY_DIR + ".git", "--work-tree", REPOSITORY_DIR, "add", *filename)
		add.Stderr = os.Stderr
		output, err = add.Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(output))

		git_diff := exec.Command("git",  "--git-dir", REPOSITORY_DIR + ".git", "--work-tree", REPOSITORY_DIR, "diff-index", "--quiet", "HEAD", "--")
		git_diff.Stderr = os.Stderr
		output, err = git_diff.Output()
		if err != nil {
			// Command failed; Files have been changed, do a git commit

			// Create a git commit
			commit := exec.Command("git", "--git-dir", REPOSITORY_DIR + ".git", "--work-tree", REPOSITORY_DIR, "commit", "-m", "feat: updated favorite_music_badge")
			commit.Stderr = os.Stderr
			output, err = commit.Output()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(output))

			// Git push the commit
			push := exec.Command("git", "--git-dir", REPOSITORY_DIR + ".git", "--work-tree", REPOSITORY_DIR, "push")
			push.Stderr = os.Stderr
			output, err = push.Output()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(output))
		} else {
			fmt.Println("Nothing has changed, same favorite music.")
		}

		remove_repository_to_modify := exec.Command("rm", "-rf", "./repository_to_modify")
		remove_repository_to_modify.Stderr= os.Stderr
		output, err = remove_repository_to_modify.Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Removed repository_to_modify")
	}
}

// Transform an string to make it safe within URL's
func safeUrl(str string) string {
	str = strings.ReplaceAll(str, "?", "%3F")
	str = strings.ReplaceAll(str, "\"", "%22")
	str = strings.ReplaceAll(str, " ", "%20")
	str = strings.ReplaceAll(str, "&", "%26")
	str = strings.ReplaceAll(str, "=", "%3D")
	str = strings.ReplaceAll(str, "\\", "%5C")
	return str
}

// Generate an image link from a name and a author
// style, logo, logoColor, logoSize, labelColor, color, and cacheSeconds are all optional
// and will be omitted if they are set to the empty string, they are added directly to the badge creator
// and are the same as in https://shields.io/badges
//
// message_color is special, if it is empty, it will be set to mistyrose
func Generate_image_link(name string, author string, message_color string, style string, logo string, logoColor string, logoSize string, labelColor string, color string, cacheSeconds string) (link string) {
	name = safeUrl(name)
	author = safeUrl(author)
	if message_color == "" {
		message_color = "mistyrose"
	}
	if style != "" {
		style = fmt.Sprintf("style=%v&", style)
	}
	if logo != "" {
		logo = fmt.Sprintf("logo=%v&", logo)
	}
	if logoColor != "" {
		logoColor = fmt.Sprintf("logoColor=%v&", logoColor)
	}
	if logoSize != "" {
		logoSize = fmt.Sprintf("logoSize=%v&", logoSize)
	}
	if labelColor != "" {
		labelColor = fmt.Sprintf("labelColor=%v&", labelColor)
	}
	if color != "" {
		color = fmt.Sprintf("color=%v&", color)
	}
	if cacheSeconds != "" {
		cacheSeconds = fmt.Sprintf("cacheSeconds=%v&", cacheSeconds)
	}
	return fmt.Sprintf("https://img.shields.io/badge/Favorite%%20music-%v%%20by%%20%v-%v?%v%v%v%v%v%v%v", name, author, message_color, style, logo, logoColor, logoSize, labelColor, color, cacheSeconds)
}

// Parse command line arguments and the flags
//
// Exits out automatically if the help flag is given or if we have an invalid amount of arguments passed
func parseCommandLineArgs() (channel_id string, user_agent *string, timeout time.Duration, message_color *string, style *string, logo *string, logoColor *string, logoSize *string, labelColor *string, color *string, cacheSeconds *string, repository *string, filename *string, err error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] CHANNEL_ID\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Version: %s\n", VERSION)
		fmt.Fprintf(os.Stderr, "This creates a badge that shows your favorite music in youtube music.\n")
		flag.PrintDefaults()
	}

	// Set up the possible flags and arguments that can be passed
	user_agent = flag.String("user-agent", "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0", "User agent used while fetching the favorite music. Do not modify this if it already works.")
	timeout_flag := flag.String("timeout", "60s", "Timeout before we stop trying to fetch the favorite music.")
	message_color = flag.String("message-color", "mistyrose", "messageColor passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	style = flag.String("style", "for-the-badge", "style passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	logo = flag.String("logo", "youtube-music", "This is not a filename. logo passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	logoColor = flag.String("logo-color", "", "logoColor passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges). Empty means we don't pass it.")
	logoSize = flag.String("logo-size", "", "logoSize passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	labelColor = flag.String("label-color", "darkred", "labelColor passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	color = flag.String("color", "", "color passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	cacheSeconds = flag.String("cacheSeconds", "", "cacheSeconds passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	repository = flag.String("repository", "", "repository to clone and update with the new favorite music badge. -file must also be added")
	filename = flag.String("filename", "", "file where we add the new favorite music badge. -repository must also be added.")

	help := flag.Bool("help", false, "Display help information")
	helpShort := flag.Bool("h", false, "Display help information")
	flag.Parse()

	if (*filename != "" && *repository == "") || (*filename == "" && *repository != "") {
		log.Print("If the file flag is given, the repository flag must also be added, and vice-versa.")
		flag.Usage()
		os.Exit(64)
	}

	// Convert the timeout to an actual timeout and return an error on failure
	timeout, err = time.ParseDuration(*timeout_flag)
	if err != nil {
		log.Print("While parsing the timeout flag (did you write the durationc correctly?)")
		return
	}

	if *help || *helpShort {
		flag.Usage()
		os.Exit(0)
	}

	// Get the channel id from the args
	if len(flag.Args()) != 1 {
		log.Print("Exactly one channel_id needs to be supplied.")
		log.Print("Please note that channel_id must be passed in last.")
		flag.Usage()
		os.Exit(64)
	}
	channel_id = flag.Args()[0]

	return
}

// Get the first favorite music from that youtube music channel
// Expects that "Enable public stats" is enabled for the youtube channel, otherwise it won't work and will hit the timeout
func GetFavoriteFromChannelId(channel_id string, user_agent string, timeout time.Duration) (name string, music_link string, author string, err error) {
	// Set language to english since we expect to get the english youtube music
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("lang", "en"),
		chromedp.Env("LANG=en"),
		chromedp.UserAgent(user_agent),
	)

	// Create a context with a timeout of 10 seconds
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer acancel()
	bctx, bcancel := chromedp.NewContext(actx)
	defer bcancel()
	ctx, cancel := context.WithTimeout(bctx, timeout)
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

	name = strings.ReplaceAll(name, "(Official Video)", "")
	name = strings.TrimSpace(name)
	music_link = "https://youtube.com/" + music_link
	return
}
