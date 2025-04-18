package main
import "github.com/lib/pq"
import (
	"gator/internal/database"
	"os"
	"fmt"
	"time"
	"context"
	"database/sql"
	"github.com/google/uuid"
	"strconv"
	"math"
)

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(s *State, c Command) error {
	return func(s *State, cmd Command) error {
		user, usrErr := s.db.GetUser(context.Background(), s.config.CurrentUser)
		if usrErr != nil{
			return fmt.Errorf("Error checking user: %s", usrErr)
		}
		return handler(s, cmd, user)
	}
} //higher order function that takes a handler of the "logged in" type and returns a "normal" handler that we can register

func ScrapeFeeds(s *State) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("no feed to fetch:%v", err)
	}

	args := database.MarkFeedFetchedParams{LastFetchedAt: sql.NullTime{
        Time:  time.Now(),
        Valid: true,
    }, ID: feed.ID}

	err = s.db.MarkFeedFetched(context.Background(), args)
	if err != nil {
		return fmt.Errorf("issue saving fetch state:%v", err)
	}

	rss, err := FetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("Fatal Error:%v\n", err)
		os.Exit(1)
	}

	for _, item := range rss.Channel.Item {
		publishedAt, err := parseDate(item.PubDate)
		if err != nil {
			return err
		}
		
		postparam := database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			Title: item.Title,
			Url: item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid: true,
			},
			PublishedAt: publishedAt,
			FeedID: feed.ID,
		}

		post, err := s.db.CreatePost(context.Background(), postparam)
		if err != nil {
    		if pqErr, ok := err.(*pq.Error); ok {
        		// Check for duplicate key error code
        		if pqErr.Code == "23505" { // Unique violation in PostgreSQL
            		// Ignore duplicate URLs
            		fmt.Printf("Ignoring duplicate post: %s\n", item.Title)
            		continue
            	}
            }
    		// For any other error, return it
    		return fmt.Errorf("error creating post: %v", err)
    	}
		fmt.Println(post)
	}
	return nil
}

func parseDate(dateStr string) (time.Time, error) {
    formats := []string{
        time.RFC1123Z,
        time.RFC1123,
        time.RFC3339,
        // Add more formats as you encounter them
    }
    
    for _, format := range formats {
        t, err := time.Parse(format, dateStr)
        if err == nil {
            return t, nil
        }
    }
    
    return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func StrToInt(str string) (int32, error) {
	int, err := strconv.Atoi(str)
	if err != nil {
	    return 0, fmt.Errorf("invalid numeric value '%s': %v", str, err)
	}

	if int <= 0 {
	    return 0, fmt.Errorf("number must be a positive number")
	}

	if int > math.MaxInt32 {
	    return 0, fmt.Errorf("numeber is too large")
	}
	return int32(int), nil

}