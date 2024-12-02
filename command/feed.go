package command

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
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

func fetchFeed(state *State, feed *database.Feed) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", feed.Url, nil)
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

	currentTime := time.Now()
	return &rss, state.DB.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		ID:            feed.ID,
		LastFetchedAt: sql.NullTime{Time: currentTime, Valid: true},
		UpdatedAt:     currentTime,
	})
}

func FetchFeed(state *State, cmd Command) error {
	if len(cmd.Args) < 1 || len(cmd.Args) > 2 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.Name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	fmt.Printf("Collecting feeds every %s...\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		feed, err := state.DB.GetNextFeedToFetch(context.Background())
		if err != nil {
			fmt.Println(fmt.Errorf("couldn't get next feed: %w", err))
			continue
		}

		rss, err := fetchFeed(state, &feed)
		if err != nil {
			fmt.Println(fmt.Errorf("couldn't fetch rss for: %s - %w", feed.Url, err))
			continue
		}

		fmt.Printf("-------- Feed: '%s' ------->\n", feed.Name)
		for _, item := range rss.Channel.Items {
			publishedAt := sql.NullTime{}
			if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
				publishedAt = sql.NullTime{
					Time:  t,
					Valid: true,
				}
			}

			currentTime := time.Now()
			_, err = state.DB.CreatePost(context.Background(), database.CreatePostParams{
				ID:        uuid.New(),
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
				FeedID:    feed.ID,
				Title:     item.Title,
				Description: sql.NullString{
					String: item.Description,
					Valid:  true,
				},
				Url:         item.Link,
				PublishedAt: publishedAt,
			})
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
					continue
				}
				fmt.Printf("Couldn't create post: %v", err)
				continue
			}
		}
		fmt.Printf("<------- Feed: '%s' --------\n", feed.Name)
	}
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

func Browse(state *State, cmd Command) error {
	limit := 2
	if len(cmd.Args) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.Args[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}

	posts, err := state.DB.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: state.User.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), state.User.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon 2. Jan"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}
