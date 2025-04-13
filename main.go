package main
import _ "github.com/lib/pq"
import (
	"Gator/internal/config"
	"Gator/internal/database"
	"os"
	"fmt"
	"time"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type state struct {
	db *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handler map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	if name == "" {
		fmt.Println("Error: Empty command name provided")
		return
	}
	if f == nil {
		fmt.Println("function cannot be empty")
		return
	}
	c.handler[name] = f
	return
} // This method registers a new handler function for a command name.

func (c *commands) run(s *state, cmd command) error {
	f := c.handler[cmd.name]
	if  f == nil {
		return fmt.Errorf("unable to retrieve command: %s", cmd.name)
	}
	err := f(s, cmd)
	if err != nil {
		return fmt.Errorf("Failed to execute command: %s", err)
	}
	return nil
} // This method runs a given command with the provided state if it exists.

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(s *state, c command) error {
	return func(s *state, cmd command) error {
		user, usrErr := s.db.GetUser(context.Background(), s.config.CurrentUser)
		if usrErr != nil{
			return fmt.Errorf("Error checking user: %s", usrErr)
		}
		return handler(s, cmd, user)
	}
} //higher order function that takes a handler of the "logged in" type and returns a "normal" handler that we can register

func scrapeFeeds(s *state) error {
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

	fmt.Println(rss.Channel.Title)
	for _, item := range rss.Channel.Item {
		fmt.Println(item.Title)
	}
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("no username provided for login")
	}
	login := cmd.args[0]
	user, usrerr := s.db.GetUser(context.Background(), login)
	if usrerr != nil{
		return fmt.Errorf("Error checking user: %s", usrerr)
	}

	err := s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("unable to set Current User: %s", err)
	}
	fmt.Printf("user has been set to %s\n", user.Name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("no username provided for login")
	}
	dt := time.Now()
	userparam := database.CreateUserParams{uuid.New(),dt,dt,cmd.args[0]}
	user, err := s.db.CreateUser(context.Background(), userparam)
	if err != nil {
		return fmt.Errorf("user creation failed: %s\n", err)
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("unable to set Current User: %s", err)
	}
	fmt.Println("user has been created ")
	fmt.Printf("ID: %v\nCreated At: %v\nUpdated At: %v\nName: %s\n",
		user.ID,user.CreatedAt,user.UpdatedAt,user.Name)
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return  fmt.Errorf("user fetch failed:%s\n", err)
	}	
	for i := 0; i < len(users); i++ {
		if users[i] == s.config.CurrentUser {
			fmt.Printf("* %s (current)\n", users[i])
		} else {
			fmt.Printf("* %s\n", users[i]) 
		}
	}
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteFeedFollows(context.Background())
	if err != nil {
		return  fmt.Errorf("failed to reset feed_follows:%s\n", err)
	}

	err = s.db.DeleteUsers(context.Background())
	if err != nil {
		return  fmt.Errorf("failed to reset users:%s\n", err)
	}

	err = s.db.DeleteFeeds(context.Background())
	if err != nil {
		return  fmt.Errorf("failed to reset feeds:%s\n", err)
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("not enough parameters. 2 expected but %n given.\n",len(cmd.args))
	}
	time_between_reqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid time format: %v", err)
	}
	ticker := time.NewTicker(time_between_reqs)
	fmt.Println("Collecting feeds every %v", time_between_reqs)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			return err
		}
		fmt.Printf("\nwaiting %v\n", time_between_reqs)
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("not enough parameters. 2 expected but %n given.\n",len(cmd.args))
	}

	dt := time.Now()
	feedParams := database.CreateFeedParams{uuid.New(),dt,dt,cmd.args[0],cmd.args[1]}
	feed, feedErr := s.db.CreateFeed(context.Background(), feedParams)
	if feedErr != nil {
		return fmt.Errorf("Error saving feed: %v\n", feedErr)
	}

	fmt.Printf("ID: %v\nCreated At: %v\nUpdated At: %v\nName: %s\nURL:%v\nLast Fetched:%v\n",
		feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.Url, feed.LastFetchedAt)

	followParams := database.CreateFeedFollowParams{uuid.New(),dt,dt,user.ID, feed.ID}
	follow, followErr := s.db.CreateFeedFollow(context.Background(), followParams)
	if followErr != nil {
		return fmt.Errorf("Error following feed: %v\n", followErr)
	}
	fmt.Printf("%s is now following the \"%v\" feed\n",follow.Username, follow.Feedname)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.SelectFeeds(context.Background())
	if err != nil{
		return fmt.Errorf("Error fetching feeds: %s", err)
	}
	for i := 0; i < len(feeds); i++ {
		feed := feeds[i]
		fmt.Printf("Feed Name: %v\n	URL:%v\n	User Name:%v\n",
			feed.Feedname, feed.Url, feed.Username)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("No URL provided")
	}
	feed, uErr := s.db.SelectFeedURL(context.Background(), cmd.args[0])
	if uErr != nil{
		return fmt.Errorf("Error fetching feed: %s", uErr)
	}
	dt := time.Now()
	followParams := database.CreateFeedFollowParams{uuid.New(),dt,dt,user.ID, feed.ID}
	follow, followErr := s.db.CreateFeedFollow(context.Background(), followParams)
	if followErr != nil {
		return fmt.Errorf("Error following feed: %v\n", followErr)
	}
	fmt.Printf("%s is now following the \"%v\" feed\n",follow.Username, follow.Feedname)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	following, err := s.db.GetFeedFollowsForUser(context.Background(),  user.Name)
	if err != nil{
		return fmt.Errorf("Error fetching feeds: %s", err)
	}
	for i := 0; i < len(following); i++ {
		follow := following[i]
		fmt.Printf("Feed Name: %v\n		User Name:%v\n",
			follow.Feedname, follow.Username)
	}
	return nil
}

func handlerUnFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("not enough parameters. 2 expected but %n given.\n",len(cmd.args))
	}
	url, uErr := s.db.SelectFeedURL(context.Background(), cmd.args[0])
	if uErr != nil {
		return uErr
	}
	arg := database.DeleteFeedFollowsForUserParams{user.ID,url.ID}
	err := s.db.DeleteFeedFollowsForUser(context.Background(), arg)
	if err != nil {
		return fmt.Errorf("unable to unfollow:%v", err)
	}
	return nil
}

func main() {
	comm := commands{make(map[string]func(*state, command) error)}
	
	comm.register("login", handlerLogin)
	comm.register("register", handlerRegister)
	comm.register("reset", handlerReset)
	comm.register("users", handlerUsers)
	comm.register("agg", handlerAgg)
	comm.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	comm.register("feeds", handlerFeeds)
	comm.register("follow", middlewareLoggedIn(handlerFollow))
	comm.register("following", middlewareLoggedIn(handlerFollowing)) //lists feeds followed by current user
	comm.register("unfollow", middlewareLoggedIn(handlerUnFollow))

	if len(os.Args) < 2 {
		fmt.Println("no arguments provided")
		os.Exit(1)
	}

	args := []string{}

	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	cmd := command{os.Args[1], args,}

	conf := &config.Config{}
	var err error
	*conf, err = config.Read()
	if err != nil {
		fmt.Printf("Error reading config:%s", err)
		return
	}

	db, dberr := sql.Open("postgres", conf.DatabaseURL)

	if dberr != nil {
		fmt.Printf("Unable to connect to DB:%s", dberr)
		return
	}
	dbQueries := database.New(db)
	systemstate := &state{dbQueries, conf,}

	err = comm.run(systemstate, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}



}