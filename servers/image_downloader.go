package servers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Daniel-W-Innes/street_view_proxy/config"
	"github.com/Daniel-W-Innes/street_view_proxy/proxy"
	"github.com/Daniel-W-Innes/street_view_proxy/view"
	"github.com/Daniel-W-Innes/street_view_proxy/workers"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
)

func saveImage(img image.Image, path string) error {
	out, err := os.Create(path + ".jpeg")
	if err != nil {
		return err
	}
	var opt jpeg.Options
	opt.Quality = 100

	err = jpeg.Encode(out, img, &opt)
	if err != nil {
		return err
	}
	return nil
}

type ImageDownloaderServer struct {
	TileWorker *workers.TileWorker
	ApiKey     string
	view.UnimplementedImageDownloaderServer
}

func getImage(tileWorker *workers.TileWorker, location *view.Location, key string, saveImages bool) (*view.Image, error) {
	log.Println("getting metadata")
	metadata, err := proxy.GetMetadata(location, key)
	log.Printf("got metadata: %v\n", metadata)
	if err != nil {
		return nil, err
	}
	if metadata.Status != "OK" {
		return nil, errors.New(metadata.Status)
	}
	log.Println("downloading mosaic")
	go tileWorker.DownloadMosaic(metadata)
	tile, err := tileWorker.GetMosaic()
	log.Println("downloaded mosaic")
	if err != nil {
		return nil, err
	}
	log.Println("encoding response")
	buf := new(bytes.Buffer)
	enc := &png.Encoder{
		CompressionLevel: png.NoCompression,
	}
	err = enc.Encode(buf, tile)
	if err != nil {
		return nil, err
	}
	if saveImages {
		go func() {
			err := saveImage(tile, fmt.Sprintf("%f,%f_x:%d-%d_y:%d-%d_%d", metadata.Location.Latitude, metadata.Location.Longitude, config.MinX, config.MaxX, config.MinY, config.MaxY, config.Zoom))
			if err != nil {
				log.Println("failed to save image")
			}
		}()
	}
	outImage := view.Image{
		Width:     int32(tile.Bounds().Dx()),
		Height:    int32(tile.Bounds().Dy()),
		ImageData: buf.Bytes(),
	}
	log.Println("sending response")
	return &outImage, nil
}

func (s *ImageDownloaderServer) GetImage(_ context.Context, location *view.Location) (*view.Image, error) {
	return getImage(s.TileWorker, location, s.ApiKey, config.SaveImages)
}
