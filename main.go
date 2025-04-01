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
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return  fmt.Errorf("failed to reset users:%s\n", err)
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"
	rss, err := FetchFeed(context.Background(), url)
	if err != nil {
		fmt.Printf("Fatal Error:%v\n", err)
		os.Exit(1)
	}
	fmt.Println(rss)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("not enough parameters. 2 expected but %n given.\n",len(cmd.args))
	}
	user, usrErr := s.db.GetUser(context.Background(), s.config.CurrentUser)
	if usrErr != nil{
		return fmt.Errorf("Error checking user: %s", usrErr)
	}

	dt := time.Now()
	feedParams := database.CreateFeedParams{uuid.New(),dt,dt,cmd.args[0],cmd.args[1],user.ID}
	feed, feedErr := s.db.CreateFeed(context.Background(), feedParams)
	if feedErr != nil {
		fmt.Errorf("Error saving feed: %v\n", feedErr)
	}
	fmt.Printf("ID: %v\nCreated At: %v\nUpdated At: %v\nName: %s\nURL:%v\nUser ID:%v\n",
		feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.Url, feed.UserID)
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

func main() {
	commandhandler := make(map[string]func(*state, command) error)
	commandhandler["login"] = handlerLogin
	commandhandler["register"] = handlerRegister
	commandhandler["reset"] = handlerReset
	commandhandler["users"] = handlerUsers
	commandhandler["agg"] = handlerAgg
	commandhandler["addfeed"] = handlerAddFeed
	commandhandler["feeds"] = handlerFeeds

	comm := commands{commandhandler}

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