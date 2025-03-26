package main
import (
	"Gator/internal/config"
	"os"
	"fmt"
)

type state struct {
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
	err := s.config.SetUser(login)
	if err != nil {
		return fmt.Errorf("unable to set Current User: %s", err)
	}
	fmt.Printf("user has been set to %s\n", login)
	return nil
}

func main() {

	commandhandler := make(map[string]func(*state, command) error)
	commandhandler["login"] = handlerLogin

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
	systemstate := &state{conf,}

	err = comm.run(systemstate, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}



}