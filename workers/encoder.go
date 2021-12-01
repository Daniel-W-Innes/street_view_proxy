package workers

import (
	"bytes"
	"fmt"
	"github.com/Daniel-W-Innes/street_view_proxy/config"
	"github.com/Daniel-W-Innes/street_view_proxy/proxy"
	"github.com/Daniel-W-Innes/street_view_proxy/view"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"runtime"
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

type EncodeRequest struct {
	Tile     image.Image
	Location proxy.Location
}

type EncodeWorker struct {
	Input      chan EncodeRequest
	stop       chan struct{}
	Server     view.ImageDownloader_GetImageServer
	SaveImages bool
}

func (e *EncodeWorker) run() {
	for {
		select {
		case encodeRequest := <-e.Input:
			buf := new(bytes.Buffer)
			enc := &png.Encoder{
				CompressionLevel: png.NoCompression,
			}
			err := enc.Encode(buf, encodeRequest.Tile)
			if err != nil {
				err := e.Server.Send(&view.Response{Error: &view.Error{Description: err.Error()}})
				if err != nil {
					log.Printf("failed to send response: %s\n", err)
				}
			}
			outImage := view.Image{
				Width:     int32(encodeRequest.Tile.Bounds().Dx()),
				Height:    int32(encodeRequest.Tile.Bounds().Dy()),
				ImageData: buf.Bytes(),
			}
			err = e.Server.Send(&view.Response{Image: &outImage})
			if err != nil {
				log.Printf("failed to send response: %s\n", err)
			}
			if e.SaveImages {
				err := saveImage(encodeRequest.Tile, fmt.Sprintf("%f,%f_x:%d-%d_y:%d-%d_%d", encodeRequest.Location.Latitude, encodeRequest.Location.Longitude, config.MinX, config.MaxX, config.MinY, config.MaxY, config.Zoom))
				if err != nil {
					log.Println("failed to save image")
				}
			}
		case <-e.stop:
			return
		}
	}
}

func GetEncodeWorkers(server view.ImageDownloader_GetImageServer, saveImages bool) *EncodeWorker {
	encodeWorker := EncodeWorker{
		Input:      make(chan EncodeRequest),
		stop:       make(chan struct{}),
		Server:     server,
		SaveImages: saveImages,
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		go encodeWorker.run()
	}
	return &encodeWorker
}

func (e *EncodeWorker) Exit() {
	for i := 0; i < runtime.NumCPU(); i++ {
		e.stop <- struct{}{}
	}
}
