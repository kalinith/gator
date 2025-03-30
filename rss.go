package main
import (
	"net/http"
	"io"
	"fmt"
	"context"
	"encoding/xml"
	"html"
	)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	
	var rss *RSSFeed

	if feedURL == "" {
		return rss, fmt.Errorf("Empty url Error\n")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return rss, fmt.Errorf("Error generating request:%v\n",err)
	}

	req.Header.Set("User-Agent","gator")

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return rss, fmt.Errorf("failed to fetch:%v\n",err)		
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return rss, fmt.Errorf("Error in body of response:%v\n",err)		
	}


	if err := xml.Unmarshal(body, &rss); err != nil {
        return rss, fmt.Errorf("Error unmarshalling XML: %v", err)
    }

	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for i := range rss.Channel.Item {
	    rss.Channel.Item[i].Title = html.UnescapeString(rss.Channel.Item[i].Title)
	    rss.Channel.Item[i].Description = html.UnescapeString(rss.Channel.Item[i].Description)
	}
	return rss, nil
}

