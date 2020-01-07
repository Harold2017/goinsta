package downloader

import (
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gedex/go-instagram/instagram"
)

func Downloader(destDir, link string, imgMinX, imgMinY int) {
	var ext string
	if strings.Contains(link, ".png") {
		ext = ".png"
	} else {
		ext = ".jpg"
	}

	resp, err := http.Get(link)
	if err != nil {
		log.Printf("Downloader: http.Get error: %s, with link: %s\n", err.Error(), link)
		return
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Printf("Downloader: image.Decode error: %s, with link: %s\n", err.Error(), link)
		return
	}

	bBox := img.Bounds()
	if bBox.Size().X < imgMinX && bBox.Size().Y < imgMinY {
		// size too small
		return
	} else {
		file, err := os.Create(destDir + "/" + strings.Split(link, "/")[4] + ext)
		if err != nil {
			log.Printf("Downloader: os.Create error: %s, with link: %s\n", err.Error(), link)
			return
		}
		defer file.Close()
		if ext == ".png" {
			err = png.Encode(file, img)
		} else {
			err = jpeg.Encode(file, img, nil)
		}
		if err != nil {
			log.Printf("Downloader: image Encode error: %s, with link: %s\n", err.Error(), link)
			return
		}
	}
}

func GetPhotos(userId string, client *instagram.Client, linkChan chan string) {
	param := &instagram.Parameters{
		Count:        20,  // number of medias to return
		MinID:        "",
		MaxID:        "",
		MinTimestamp: 0,
		MaxTimestamp: 0,
		Lat:          0,
		Lng:          0,
		Distance:     0,
	}

	for {
		mediaList, next, err := client.Users.RecentMedia(userId, param)
		if err != nil {
			log.Printf("GetPhotos error: %s, with userId: %s\n", err.Error(), userId)
			break
		}
		if next != nil {
			param.MaxID = next.NextMaxID
		}

		for _, media := range mediaList {
			linkChan <- media.Images.StandardResolution.URL
		}

		if len(mediaList) == 0 || next.NextMaxID == "" {
			break
		}
	}
}
