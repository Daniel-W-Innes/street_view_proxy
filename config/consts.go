package config

const (
	Port             = 8080
	MinX             = 7
	MinY             = 3
	MaxX             = 9
	MaxY             = 5
	Zoom             = 4
	InputSizeX       = 512
	InputSizeY       = InputSizeX
	OutputSizeX      = InputSizeX * (MaxX - MinX)
	OutputSizeY      = InputSizeY * (MaxY - MinY)
	WorkerMultiplier = 10
	SaveImages       = false
	NumRetries       = 10
)
