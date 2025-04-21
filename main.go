package main

import (
	"fmt"
	"time"
	router "GoShare/Project/pkg"
)

func main() {
	presentTime := time.Now()
	url := "http://" + router.GetIP() + ":8080/"
	fmt.Println("The current time is: ", presentTime)
	fmt.Println("Starting server...")
	fmt.Println("Listening on : ", url)
	r := router.Router()
	router.OpenWebsite(url)
	r.Run("0.0.0.0:8080")
}