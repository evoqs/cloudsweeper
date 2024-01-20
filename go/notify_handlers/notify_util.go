package notifications

import (
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"

	logging "cloudsweep/logging"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func downloadSvgFile(source, destination string) error {
	// Check if the source is a URL
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// TODO: Check if the file has changed, only then download
		// TODO: Try for ETag, else sha256sum or size in the order of preference
		response, err := http.Get(source)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		// Create or open the destination file, overwriting if it already exists
		file, err := os.Create(destination)
		if err != nil {
			return err
		}
		defer file.Close()

		// Copy the contents from the response body to the destination file
		_, err = io.Copy(file, response.Body)
		if err != nil {
			return err
		}
		return nil
	}

	// It's a local file
	sourceContent, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	// Write the content to the destination file, overwriting if it already exists
	if err := os.WriteFile(destination, sourceContent, 0644); err != nil {
		return err
	}
	return nil
}

func convertSVGtoPNG(source string, outPath string, width, height int) error {
	logging.NewDefaultLogger().Debugf("Converting source: %s to output: %s with width: %d height: %d", source, outPath, width, height)
	// Read SVG file content
	inputFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	// Create SVG icon
	icon, err := oksvg.ReadIconStream(inputFile)
	if err != nil {
		return err
	}

	// Set target dimensions
	icon.SetTarget(0, 0, float64(width), float64(height))

	// Create RGBA image
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))

	// Draw SVG to image
	icon.Draw(rasterx.NewDasher(width, height, rasterx.NewScannerGV(width, height, rgba, rgba.Bounds())), 1)

	// Create output file
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Encode as PNG and write to the output file
	err = png.Encode(out, rgba)
	if err != nil {
		logging.NewDefaultLogger().Warnf("Error : %v", err)
		return err
	}

	return nil
}

func getAndConvertLogo(remoteSvgPath string, localSvgPath string, localPngPath string, width int, height int) error {
	if err := downloadSvgFile(remoteSvgPath, localSvgPath); err != nil {
		return err
	}
	return convertSVGtoPNG(localSvgPath, localPngPath, width, height)
}
