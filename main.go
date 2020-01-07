package main

import (
	"flag"
	"fmt"
	"github.com/gedex/go-instagram/instagram"
	"goinsta/downloader"
	"log"
	"os"
	"sync"
	"time"
)

var (
	clientId string

	help         bool
	version      bool
	instaName       string
	minX     int
	minY      int
	destination  string
	workers int
)

func init() {
	clientId = os.Getenv("InstagramID")
	if clientId == "" {
		fmt.Println("Please set InstagramID env with `export InstagramID=xxxxxx` before using this tool")
		os.Exit(1)
	}

	flag.BoolVar(&help, "h", false, "help info")
	flag.BoolVar(&version, "v", false, "version info")
	flag.StringVar(&instaName, "n", "", "Instagram username")
	flag.IntVar(&minX, "x", 300, "minimum photo size x (default 300)")
	flag.IntVar(&minY, "y", 300, "minimum photo size y (default 300)")
	flag.StringVar(&destination, "d", "", "destination path of downloaded files")
	flag.IntVar(&workers, "w", 2, "number of  concurrent workers (default 2)")
	flag.Usage = usage
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `Instagram downloader tool in golang to download photos
Version: 0.0.1
Usage: goinsta [-hvnxyd] [-h help] [-v version] [-n instagram username] [-x minimum photo size x (default 300)] [-y minimum photo size y (default 300)] [-d destination path of downloaded files] [-w number of  concurrent workers (default 2)]
Options
`)
	flag.PrintDefaults()
}

func run(clientId, instaName string) {
	client := instagram.NewClient(nil)
	client.ClientID = clientId

	var userId string
	users, _, err := client.Users.Search(instaName, nil)
	if err != nil {
		log.Fatalln("Can NOT find input instagram username: ", instaName, err)
		return
	}
	for _, user := range users {
		if user.Username == instaName {
			userId = user.ID
		}
	}

	if userId == "" {
		log.Fatalln("Can NOT find input instagram username: ", instaName)
		return
	}

	download(userId, client)
}

func download(userId string, client *instagram.Client) {
	fmt.Printf("Starting download photos for user (id: %s)\n", userId)
	wg := &sync.WaitGroup{}
	link := make(chan string, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(link chan string) {
			defer wg.Done()
			for l := range link {
				downloader.Downloader(destination, l, minX, minY)
			}
		}(link)
	}
	downloader.GetPhotos(userId, client, link)
	wg.Wait()
}

func main() {
	flag.Parse()

	if help {
		flag.Usage()
	} else if version {
		fmt.Println("version: 0.0.1")
	} else if instaName == "" || destination == "" {
		fmt.Println("too less arguments, use '-h' to see help info")
		os.Exit(1)
	} else {
		start := time.Now()
		// create dest folder if not exists
		if _, err := os.Stat(destination); os.IsNotExist(err) {
			err = os.MkdirAll(destination, os.ModePerm)
			if err != nil {
				fmt.Println("Fail to create destination folder")
				os.Exit(1)
			}
		}

		run(clientId, instaName)

		fmt.Println("Successfully executed")
		fmt.Println("time consumption: ", time.Now().Sub(start).Seconds())
	}
}
