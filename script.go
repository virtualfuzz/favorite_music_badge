package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
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
	providers, user_agent, timeout, message_color, style, logo, logoColor, logoSize, labelColor, color, cacheSeconds, repository, filename, err := parseCommandLineArgs()
	if err != nil {
		log.Fatal(err)
	}

	// Fetch the favorite music
	name, song_link, author := get_favorite_from_provider(providers, user_agent, timeout)

	// Create a link of it as an image
	image_link := Generate_image_link(name, author, message_color, style, logo, logoColor, logoSize, labelColor, color, cacheSeconds)
	fmt.Printf("Favorite music: %v by %v (%v)\n", name, author, song_link)
	fmt.Println(image_link)

	if repository != "" {
		fmt.Println("The image link has been generated we are now downloading the repository and adding the favorite_music_badge to it!")
		err = AddImageToRepository(repository, filename, image_link, song_link)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Get the favorite music from a list of providers, we try the first provider,
// then the second, etc.
func get_favorite_from_provider(providers []Provider, user_agent string, timeout time.Duration) (name string, song_link string, author string) {
	var err error
	for i := range providers {
		switch providers[i].Type {
		case Youtube:
			fmt.Println("Fetching most listened song from youtube...")
			fmt.Println("Please make sure that \"Enable public stats\" is enabled in your youtube music channel settings.")
			fmt.Printf("Currently fetching the favorite music, this might take a bit long... (Timeout of %v)\n", timeout)
			name, song_link, author, err = GetFavoriteFromChannelId(providers[i].YoutubeChannelId, user_agent, timeout)
			if err != nil {
				log.Print(err)
				log.Print("Failed to fetch from YouTube")
			} else {
				return
			}
		case LastFm:
			fmt.Println("Fetching top song from last.fm...")
			name, song_link, author, err = GetTopSongFromLastFm(providers[i].LastFmUsername, providers[i].LastFmPeriod, providers[i].LastFmAPIKey)
			if err != nil {
				log.Print(err)
				log.Print("Failed to fetch from last.fm")
			} else {
				return
			}
		}
	}

	return
}

// Function to download a git repository and push the new image to it
func AddImageToRepository(repository string, filename string, image_link string, video_link string) (err error) {
	// Clone the repository
	output, err := run("git", "clone", repository, REPOSITORY_DIR)
	if err != nil {
		return
	}
	fmt.Println(string(output))

	// Search the file and add the youtube music badge
	file, err := os.Open(REPOSITORY_DIR + filename)
	if err != nil {
		return
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
		return errors.New("Tried to add a favorite music badge without a FAVORITE_MUSIC_BADGE_AFTER_THIS_LINE inside of the readme")
	}

	err = scanner.Err()
	if err != nil {
		return
	}

	// Overwrite the existing file
	outputFile, err := os.Create(REPOSITORY_DIR + filename)
	if err != nil {
		return
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, line := range lines {
		_, err = writer.WriteString(line + "\n")
		if err != nil {
			return
		}
	}
	writer.Flush()

	// Try to do a git add the modified file
	output, err = run("git", "--git-dir", REPOSITORY_DIR+".git", "--work-tree", REPOSITORY_DIR, "add", filename)
	if err != nil {
		return
	}
	fmt.Println(string(output))

	output, err = run("git", "--git-dir", REPOSITORY_DIR+".git", "--work-tree", REPOSITORY_DIR, "diff-index", "--quiet", "HEAD", "--")
	if err != nil {
		// Command failed; Files have been changed, do a git commit

		// Create a git commit
		output, err = run("git", "--git-dir", REPOSITORY_DIR+".git", "--work-tree", REPOSITORY_DIR, "commit", "-m", "feat: updated favorite_music_badge")
		if err != nil {
			return
		}

		// Git push the commit
		output, err = run("git", "--git-dir", REPOSITORY_DIR+".git", "--work-tree", REPOSITORY_DIR, "push")
		if err != nil {
			return
		}
	} else {
		fmt.Println("Nothing has changed, same favorite music. Not trying to update repository.")
	}

	output, err = run("rm", "-rf", "./repository_to_modify")
	if err != nil {
		return
	}
	fmt.Println("Removed repository_to_modify")

	return nil
}

// Helper function to run a command
func run(name string, arg ...string) (output []byte, err error) {
	command := exec.Command(name, arg...)
	command.Stderr = os.Stderr
	output, err = command.Output()
	fmt.Println(string(output))
	return
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

// Type of a provider
type ProviderType string

const (
	Youtube ProviderType = "youtube"
	LastFm  ProviderType = "lastfm"
)

// A provider with a type and specific fields
type Provider struct {
	Type             ProviderType
	YoutubeChannelId string
	LastFmUsername   string
	LastFmAPIKey     string
	LastFmPeriod     string
}

// Parse command line arguments and the flags
//
// # Exits out automatically if the help flag is given or if we have an invalid amount of arguments passed
//
// Required:
// - if filename THEN repository and vice versa
// - one "provider" needs to be given (youtube information/lastfm information)
func parseCommandLineArgs() (providers []Provider, userAgent string, timeout time.Duration, messageColor string, style string, logo string, logoColor string, logoSize string, labelColor string, color string, cacheSeconds string, repository string, filename string, err error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Version: %s\n", VERSION)
		fmt.Fprintf(os.Stderr, "This creates a badge that shows your favorite music in youtube music or lastfm.\n")
		flag.PrintDefaults()
	}

	// Set up the possible flags and arguments that can be passed
	flag.StringVar(&userAgent, "user-agent", "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0", "[DEPRECATED, use userAgent].")
	flag.StringVar(&userAgent, "userAgent", "Mozilla/5.youtubeChannelId string, lastFmUsername string, lastFmPeriod string, lastFmAPIKey string, 0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0", "User agent used while fetching the favorite music. Do not modify this if it already works.")
	timeoutFlag := flag.String("timeout", "60s", "Timeout before we stop trying to fetch the favorite music.")
	flag.StringVar(&messageColor, "message-color", "mistyrose", "[DEPRECATED, use messageColor]")
	flag.StringVar(&messageColor, "messageColor", "mistyrose", "messageColor passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	flag.StringVar(&style, "style", "for-the-badge", "style passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	flag.StringVar(&logo, "logo", "youtube-music", "This is not a filename. logo passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	flag.StringVar(&logoColor, "logo-color", "", "[DEPRECATED, use logoColor]")
	flag.StringVar(&logoColor, "logoColor", "", "logoColor passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges). Empty means we don't pass it.")
	flag.StringVar(&logoSize, "logo-size", "", "[DEPRECATED, use logoSize]")
	flag.StringVar(&logoSize, "logoSize", "", "logoSize passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	flag.StringVar(&labelColor, "label-color", "darkred", "[DEPRECATED, use labelColor]")
	flag.StringVar(&labelColor, "labelColor", "darkred", "labelColor passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	flag.StringVar(&color, "color", "", "color passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	flag.StringVar(&cacheSeconds, "cacheSeconds", "", "cacheSeconds passed to shields.io while generating the markdown badge (documentation at https://shields.io/badges)")
	flag.StringVar(&repository, "repository", "", "repository to clone and update with the new favorite music badge. -file must also be added")
	flag.StringVar(&filename, "filename", "", "file where we add the new favorite music badge. -repository must also be added.")
	var youtubeChannelId string
	flag.StringVar(&youtubeChannelId, "youtubeChannelId", "", "Youtube channel ID if you want to get the most listened music from a channel. They must have \"Enable public stats\" turned on.")
	var lastFmUsername string
	flag.StringVar(&lastFmUsername, "lastFmUsername", "", "Last.fm username where we get the top song from.")
	var lastFmPeriod string
	flag.StringVar(&lastFmPeriod, "lastFmPeriod", "7day", "Last.fm period over which to retrieve top tracks for.")
	var lastFmAPIKey string
	lastFmAPIKey = os.Getenv("LAST_FM_API_KEY")
	var fallback string
	flag.StringVar(&fallback, "fallback", "", "Required if multiple providers are used (youtube and last.fm for example), each provider are separated by ','. The first one has higher priority over the lower one, if we can't find the favorite song from the first one, we take it from the other ones.")

	help := flag.Bool("help", false, "Display help information")
	helpShort := flag.Bool("h", false, "Display help information")
	flag.Parse()

	if (filename != "" && repository == "") || (filename == "" && repository != "") {
		log.Print("If the file flag is given, the repository flag must also be added, and vice-versa.")
		flag.Usage()
		os.Exit(64)
	}

	// Convert the timeout to an actual timeout and return an error on failure
	timeout, err = time.ParseDuration(*timeoutFlag)
	if err != nil {
		log.Print("While parsing the timeout flag (did you write the durationc correctly?)")
		return
	}

	if *help || *helpShort {
		flag.Usage()
		os.Exit(0)
	}

	// Check if the providers are correct
	if lastFmAPIKey == "" && lastFmUsername != "" {
		log.Print("If the lastFmUsername flag is given, the LAST_FM_API_KEY environment variable must be given.")
		flag.Usage()
		os.Exit(64)
	}

	if len(flag.Args()) == 1 {
		if youtubeChannelId != "" {
			log.Print("[ERROR] A youtube channel id was supplied by both --youtubeChannelId and the first argument, please use the --youtubeChannelId argument.")
			flag.Usage()
			os.Exit(64)
		} else {
			log.Print("[WARNING] A youtube channel id was provided using a normal argument, this has been deprecated, but will continue to function normally. Please use --youtubeChannelId from now on.")
			youtubeChannelId = flag.Args()[0]
		}
	}

	if youtubeChannelId != "" {
		providers = append(providers, Provider{Type: Youtube, YoutubeChannelId: youtubeChannelId})
	}

	if lastFmUsername == "" && youtubeChannelId == "" {
		log.Print("[ERROR] A last.fm username (--lastFmUsername and --lastFmAPIKey) or a youtube channel id (--youtubeChannelId) must be given, we have no idea where to take the favorite music from!")
		flag.Usage()
		os.Exit(64)
	}

	if lastFmUsername != "" {
		providers = append(providers, Provider{Type: LastFm, LastFmUsername: lastFmUsername, LastFmAPIKey: lastFmAPIKey, LastFmPeriod: lastFmPeriod})
	}

	fallback_order := strings.Split(fallback, ",")
	if len(providers) != 1 {
		if len(fallback_order) == len(providers) {
			for i := range fallback_order {
				switch strings.ToLower(fallback_order[i]) {
				case string(Youtube):
					moveProviderToIndex(providers, Youtube, i)
				case string(LastFm):
					moveProviderToIndex(providers, LastFm, i)
				default:
					log.Printf("[ERROR] Unknown provider passed, \"%v\" is an unknown provider. \"youtube\" and \"lastfm\" are all valid providers.\n", fallback_order[i])
					flag.Usage()
					os.Exit(64)
				}
			}
		} else {
			log.Print("[ERROR] A fallback order must be given if there are multiple providers used (lastfm and youtube for example). For example, to have last.fm have a higher priority over youtube, use (--fallback \"lastfm,youtube\"")
			flag.Usage()
			os.Exit(64)
		}
	}

	return
}

func moveProviderToIndex(providers []Provider, provider_type ProviderType, wanted_index int) (err error) {
	for i := range providers {
		if providers[i].Type == provider_type {
			if i == wanted_index {
				return nil
			} else {
				providers[wanted_index], providers[i] = providers[i], providers[wanted_index]
				return nil
			}
		}
	}

	return fmt.Errorf("Couldn't find the provider \"%v\" inside of the passed --fallback (%v)", provider_type, providers)
}

// Last.fm artist when we are parsing
type Artist struct {
	Name string `json:"name"`
}

// Last.fm track information parsed as json
type Track struct {
	Name   string `json:"name"`
	Url    string `json:"url"`
	Artist Artist `json:"artist"`
}

// Last fm toptrack information
type TopTracks struct {
	Track []Track `json:"track"`
}

type LastFMTopTracks struct {
	TopTracks TopTracks `json:"toptracks"`
}

// Get the top song from the lastfm API, would work inside of cicd
//
// API documentation: https://www.last.fm/api/show/user.getTopTracks
func GetTopSongFromLastFm(user string, period string, api_key string) (name string, music_link string, author string, err error) {
	request := fmt.Sprintf("http://ws.audioscrobbler.com/2.0/?method=user.gettoptracks&user=%v&period=%v&api_key=%v&limit=1&format=json", user, period, api_key)

	resp, err := http.Get(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var content []byte
		content, err = io.ReadAll(resp.Body)
		if err == nil {
			log.Println(string(content))
		} else {
			log.Println("Couldn't get the error message.")
			log.Println(err)
		}

		log.Println("Get the corresponding error number from https://www.last.fm/api/show/user.getTopTracks")
		err = errors.New(fmt.Sprint("Request failed with status:", resp.Status))
		return
	}

	var lastFMTopTracks LastFMTopTracks
	if err = json.NewDecoder(resp.Body).Decode(&lastFMTopTracks); err != nil {
		return
	}

	name = lastFMTopTracks.TopTracks.Track[0].Name
	music_link = lastFMTopTracks.TopTracks.Track[0].Url
	author = lastFMTopTracks.TopTracks.Track[0].Artist.Name
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
