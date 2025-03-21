package main
import (
	"github.com/kalinith/Gator/internal/config"
	"fmt"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config:%s", err)
		return
	}

	err = conf.SetUser("Kalinith")
	if err != nil {
		fmt.Printf("Error writing config:%s", err)
		return
	}
	config2, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config:%s", err)
		return
	}
	fmt.Println(config2.DatabaseURL, config2.CurrentUser)
}