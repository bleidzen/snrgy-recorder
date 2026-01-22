package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
)

// generateAppIcon creates a simple SNRGY-branded icon
func generateAppIcon() fyne.Resource {
	size := 128
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Background - dark
	bgColor := color.RGBA{R: 18, G: 18, B: 18, A: 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Draw a green circle in the center (recording indicator style)
	centerX, centerY := size/2, size/2
	radius := size / 3
	accentColor := color.RGBA{R: 0, G: 220, B: 160, A: 255}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - centerX
			dy := y - centerY
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, accentColor)
			}
		}
	}

	// Draw inner dark circle (donut effect)
	innerRadius := radius / 2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - centerX
			dy := y - centerY
			if dx*dx+dy*dy <= innerRadius*innerRadius {
				img.Set(x, y, bgColor)
			}
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	png.Encode(&buf, img)

	return fyne.NewStaticResource("icon.png", buf.Bytes())
}

// generateTrayIcon creates an ICO format icon for Windows system tray
func generateTrayIcon() []byte {
	size := 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Solid bright green circle
	green := color.RGBA{R: 0, G: 220, B: 160, A: 255}
	centerX, centerY := size/2, size/2
	radius := size/2 - 1

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - centerX
			dy := y - centerY
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, green)
			}
		}
	}

	// Encode PNG first
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	pngData := pngBuf.Bytes()

	// Create ICO format (Windows Vista+ supports PNG in ICO)
	ico := new(bytes.Buffer)

	// ICONDIR header
	ico.Write([]byte{0, 0}) // Reserved
	ico.Write([]byte{1, 0}) // Type: 1 = ICO
	ico.Write([]byte{1, 0}) // Count: 1 image

	// ICONDIRENTRY
	ico.WriteByte(byte(size))           // Width
	ico.WriteByte(byte(size))           // Height
	ico.WriteByte(0)                    // Color palette
	ico.WriteByte(0)                    // Reserved
	ico.Write([]byte{1, 0})             // Color planes
	ico.Write([]byte{32, 0})            // Bits per pixel

	// Size of image data (4 bytes, little endian)
	imgSize := uint32(len(pngData))
	ico.WriteByte(byte(imgSize))
	ico.WriteByte(byte(imgSize >> 8))
	ico.WriteByte(byte(imgSize >> 16))
	ico.WriteByte(byte(imgSize >> 24))

	// Offset to image data (4 bytes, little endian) - starts after header (22 bytes)
	ico.Write([]byte{22, 0, 0, 0})

	// PNG image data
	ico.Write(pngData)

	return ico.Bytes()
}
