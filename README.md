# Picrocess - A Go Image Processing Library

## Overview

Picrocess is a Go package designed for handling and manipulating images in various formats, including PNG, JPEG, and GIF. It supports a variety of operations such as resizing, cropping, overlaying, and rendering text on images. Additionally, it offers functionality for creating and manipulating GIFs and generating QR codes.

## Features

- **Image Manipulation**: Resize, crop, and overlay images.
- **Text Rendering**: Add custom text to images with configurable font size and color.
- **QR Code Generations**: Create customizable QR codes with configurable size and colors.
- **Image Formats**: Support for PNG, JPEG, and GIF encoding and decoding.
- **GIF Creation**: Generate GIFs with multiple frames and adjustable delays.
- **Font Handling**: Load and use TrueType fonts for text rendering.

## Installation

To install Picrocess, use the following Go command:

```bash
go get github.com/fluffy-melli/picrocess
```

## Example Usage

Below is a simple example that demonstrates how to use Picrocess to load an image, overlay text, and save it as a PNG:

```go
package main

import (
	"fmt"
	"log"

	"github.com/fluffy-melli/picrocess"
)

func main() {
	// Define the content for the QR code (URL in this case)
	content := "https://www.example.com"
	
	// Generate the QR code with the specified content, size, and colors
	qrCode, err := picrocess.NewQRCode(content, 256, picrocess.NewRGBA(0, 0, 0, 255), picrocess.NewRGBA(255, 255, 255, 255))
	if err != nil {
		log.Fatal(err) // Handle error if QR code generation fails
	}

	// Load the specified font for text rendering
	font, err := picrocess.LoadFont("HakgyoansimDunggeunmisoTTF-B.ttf")
	if err != nil {
		log.Fatal(err) // Handle error if font loading fails
	}

	// Calculate the width of the text "Scan me!" at font size 24
	ow, _ := font.TextSize(24, "Scan me!")

	// Add the text "Scan me!" to the center of the QR code
	err = qrCode.Text(font, picrocess.NewRGBA(255, 0, 0, 255), picrocess.NewOffset((256-ow)/2, 0), 24, "Scan me!")
	if err != nil {
		log.Fatal(err) // Handle error if adding text fails
	}

	// Save the modified QR code image with the text as a PNG file
	err = qrCode.SaveAsPNG("qrcode_output.png")
	if err != nil {
		log.Fatal(err) // Handle error if saving the PNG file fails
	}

	// Print a success message
	fmt.Println("QR code saved as qrcode_output.png")
}
```

## Types

### `RGBA`

The `RGBA` type represents a color with Red, Green, Blue, and Alpha (transparency) channels.

#### Constructor

```go
func NewRGBA(r, g, b uint8, a ...uint8) *RGBA
```

### `Rect`

The `Rect` type represents a rectangle with top-left and bottom-right coordinates.

#### Constructor

```go
func NewRect(w1, h1, w2, h2 uint) *Rect
```

### `Offset`

The `Offset` type represents an offset for positioning elements.

#### Constructor

```go
func NewOffset(w, h uint) *Offset
```

### `Font`

The `Font` type represents a TrueType font used for rendering text.

#### Constructor

```go
func LoadFont(filename string) (*Font, error)
```

### `Image`

The `Image` type represents an image with width, height, and a pixel map.

#### Constructor

```go
func NewImage(w, h uint, color *RGBA) *Image
```

#### Methods

- `At(x, y uint) *RGBA`: Get the color of a pixel at (x, y).
- `Set(x, y uint, c *RGBA)`: Set the color of a pixel at (x, y).
- `Overlay(i2 *Image, o *Offset)`: Overlay another image on top of the current image.
- `Resize(w, h uint)`: Resize the image to the given width and height.
- `Crop(r *Rect) *Image`: Crop the image to a rectangle.
- `Text(font *Font, c *RGBA, o *Offset, size float64, text string)`: Render text on the image.
- `Round(px uint)`: Apply rounded corners to the image with a specified radius in pixels.
- `Render() *image.RGBA`: Render the image as an `image.RGBA` type.
- `ToPNGByte() ([]byte, error)`: Convert the image to a PNG byte slice.
- `ToJPGByte(quality int) ([]byte, error)`: Convert the image to a JPG byte slice.
- `SaveAsPNG(filename string) error`: Save the image as a PNG file.
- `SaveAsJPG(filename string, quality int) error`: Save the image as a JPG file.

### `GIF`

The `GIF` type represents an animated GIF with multiple frames and delays.

#### Constructor

```go
func NewGIF() *GIF
```

#### Methods

- `Append(image *Image, delay int)`: Append a frame to the GIF with a specified delay.
- `ToGIFByte() ([]byte, error)`: Convert the GIF to a byte slice.
- `SaveAsGIF(filename string) error`: Save the GIF as a file.

### `QRCode`

The `QRCode` type represents a QR code image.

```go
func NewQRCode(content string, size int, fgColor RGBA, bgColor RGBA) (*Image, error)
```

## Supported Formats

- **PNG**: Using `png.Encode` and `png.Decode` for encoding and decoding.
- **JPEG**: Using `jpeg.Encode` and `jpeg.Decode` for encoding and decoding.
- **GIF**: Using `gif.EncodeAll` and `gif.DecodeAll` for encoding and decoding.

## Notes

- The library requires the `golang.org/x/image/webp` package for WebP support.
- The `github.com/golang/freetype` package is used for rendering text with TrueType fonts.

## License

This library is open source and available under the MIT License.