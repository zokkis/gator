package command

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zokkis/gator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Items       []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var rss RSSFeed
	err = xml.NewDecoder(res.Body).Decode(&rss)
	if err != nil {
		return nil, err
	}

	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for i, item := range rss.Channel.Items {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		rss.Channel.Items[i] = item
	}

	return &rss, nil
}

func FetchFeed(state *State, cmd Command) error {
	rss, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("couldn't get feed: %w", err)
	}

	fmt.Println(rss)

	return nil
}

func AddFeed(state *State, cmd Command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	currentTime := time.Now()
	feed, err := state.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      name,
		Url:       url,
		UserID:    state.User.ID,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return fmt.Errorf("couldn't add feed: %w", err)
	}

	_, err = state.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    state.User.ID,
		FeedID:    feed.ID,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return fmt.Errorf("couldn't follow feed: %w", err)
	}

	return nil
}

func ListFeeds(state *State, cmd Command) error {
	feeds, err := state.DB.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't get feeds: %w", err)
	}

	for i, feed := range feeds {
		user, err := state.DB.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("couldn't get user: %w", err)
		}

		fmt.Printf("-------- Feed: %d ------->\n", i+1)
		fmt.Printf("Name: %s\n", feed.Name)
		fmt.Printf("Url: %s\n", feed.Url)
		fmt.Printf("Creator: %s\n", user.Name)
		fmt.Printf("<------- Feed: %d --------\n", i+1)
	}

	return nil
}

func FollowFeed(state *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}
	url := cmd.Args[0]

	feed, err := state.DB.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't get feed with url(%s): %w", url, err)
	}

	currentTime := time.Now()
	_, err = state.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    state.User.ID,
		FeedID:    feed.ID,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return fmt.Errorf("couldn't follow feed: %w", err)
	}

	fmt.Println("-------- Feed ------->")
	fmt.Printf("Name: %s\n", feed.Name)
	fmt.Printf("Url: %s\n", feed.Url)
	fmt.Printf("Creator: %s\n", state.User.Name)
	fmt.Println("<------- Feed --------")

	return nil
}

func ListFollowing(state *State, cmd Command) error {
	feeds, err := state.DB.GetFeedFollowsForUser(context.Background(), state.User.ID)
	if err != nil {
		return fmt.Errorf("couldn't get following feeds: %w", err)
	}

	fmt.Println("-------- Feeds following ------->")
	for _, feed := range feeds {
		user, err := state.DB.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("couldn't get user: %w", err)
		}

		fmt.Printf(" * %s - %s\n", user.Name, feed.Name)
	}
	fmt.Println("<------- Feeds following --------")

	return nil
}

func Unfollow(state *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}
	url := cmd.Args[0]

	feed, err := state.DB.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't get feed by url: %w", err)
	}

	return state.DB.DeleteFeedFollows(context.Background(), database.DeleteFeedFollowsParams{
		UserID: state.User.ID,
		FeedID: feed.ID,
	})
}
