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

type State struct {
	db *database.Queries
	config *config.Config
}

type Command struct {
	name string
	args []string
}

type commands struct {
	handler map[string]func(*State, Command) error
}

func (c *commands) register(name string, f func(*State, Command) error) {
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
} // This method registers a new handler function for a  name.

func (c *commands) run(s *State, cmd Command) error {
	f := c.handler[cmd.name]
	if  f == nil {
		return fmt.Errorf("unable to retrieve command: %s", cmd.name)
	}
	err := f(s, cmd)
	if err != nil {
		return fmt.Errorf("Failed to execute command: %s", err)
	}
	return nil
} // This method runs a given command with the provided State if it exists.

func handlerLogin(s *State, cmd Command) error {
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

func handlerRegister(s *State, cmd Command) error {
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

func handlerUsers(s *State, cmd Command) error {
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

func handlerReset(s *State, cmd Command) error {
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

func handlerAgg(s *State, cmd Command) error {
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
		err := ScrapeFeeds(s)
		if err != nil {
			return err
		}
		fmt.Printf("\nwaiting %v\n", time_between_reqs)
	}
	return nil
}

func handlerAddFeed(s *State, cmd Command, user database.User) error {
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

func handlerFeeds(s *State, cmd Command) error {
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

func handlerFollow(s *State, cmd Command, user database.User) error {
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

func handlerFollowing(s *State, cmd Command, user database.User) error {
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

func handlerUnFollow(s *State, cmd Command, user database.User) error {
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
	comm := commands{make(map[string]func(*State, Command) error)}
	
	comm.register("login", handlerLogin)
	comm.register("register", handlerRegister)
	comm.register("reset", handlerReset)
	comm.register("users", handlerUsers)
	comm.register("agg", handlerAgg)
	comm.register("addfeed", MiddlewareLoggedIn(handlerAddFeed))
	comm.register("feeds", handlerFeeds)
	comm.register("follow", MiddlewareLoggedIn(handlerFollow))
	comm.register("following", MiddlewareLoggedIn(handlerFollowing)) //lists feeds followed by current user
	comm.register("unfollow", MiddlewareLoggedIn(handlerUnFollow))

	if len(os.Args) < 2 {
		fmt.Println("no arguments provided")
		os.Exit(1)
	}

	args := []string{}

	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	cmd := Command{os.Args[1], args,}

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
	systemState := &State{dbQueries, conf,}

	err = comm.run(systemState, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}



}