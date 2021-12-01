package servers

import (
	"errors"
	"github.com/Daniel-W-Innes/street_view_proxy/config"
	"github.com/Daniel-W-Innes/street_view_proxy/proxy"
	"github.com/Daniel-W-Innes/street_view_proxy/view"
	"github.com/Daniel-W-Innes/street_view_proxy/workers"
	"io"
	"log"
)

type ImageDownloaderServer struct {
	ApiKey string
	view.UnimplementedImageDownloaderServer
}

func getImage(tileWorker *workers.TileWorker, encodeWorker *workers.EncodeWorker, location *view.Location, key string) error {
	log.Println("getting metadata")
	metadata, err := proxy.GetMetadata(location, key)
	log.Printf("got metadata: %v\n", metadata)
	if err != nil {
		return err
	}
	if metadata.Status != "OK" {
		return errors.New(metadata.Status)
	}
	log.Println("downloading mosaic")
	go tileWorker.DownloadMosaic(metadata)
	tile, err := tileWorker.GetMosaic()
	log.Println("downloaded mosaic")
	if err != nil {
		return err
	}
	log.Println("encoding response")
	encodeWorker.Input <- workers.EncodeRequest{
		Tile:     tile,
		Location: metadata.Location,
	}
	return nil
}

func (s *ImageDownloaderServer) GetImage(server view.ImageDownloader_GetImageServer) error {
	log.Println("opened request channel")
	tileWorker := workers.GetTileWorkers()
	defer tileWorker.Exit()
	encodeWorker := workers.GetEncodeWorkers(server, config.SaveImages)
	defer encodeWorker.Exit()
	log.Println("created workers")
	for {
		log.Println("ready for request")
		in, err := server.Recv()
		log.Println("got request")
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		location := in.Location
		if location != nil {
			err := getImage(tileWorker, encodeWorker, location, s.ApiKey)
			if err != nil {
				return server.Send(&view.Response{Error: &view.Error{Description: err.Error()}})
			}
		}
	}
}
