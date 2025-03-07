# Picrocess - A Go Image Processing Library

## Overview

Picrocess is a Go package designed for handling and manipulating images in various formats, including PNG, JPEG, and GIF. It supports a variety of operations such as resizing, cropping, overlaying, and rendering text on images. Additionally, it offers functionality for creating and manipulating GIFs.

## Features

- **Image Manipulation**: Resize, crop, and overlay images.
- **Text Rendering**: Add custom text to images with configurable font size and color.
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
	"os"

	"github.com/fluffy-melli/picrocess"
)

func main() {
	// Load an image
	img, err := picrocess.LoadImage("input.jpg")
	if err != nil {
		log.Fatal(err)
	}

	// Load a font
	font, err := picrocess.LoadFont("path/to/font.ttf")
	if err != nil {
		log.Fatal(err)
	}

	// Add text to the image
	err = img.Text(font, picrocess.NewRGBA(255, 0, 0, 255), picrocess.NewOffset(100, 100), 24, "Hello, World!")
	if err != nil {
		log.Fatal(err)
	}

	// Save the modified image as a PNG
	err = img.SaveAsPNG("output.png")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Image saved as output.png")
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

## Supported Formats

- **PNG**: Using `png.Encode` and `png.Decode` for encoding and decoding.
- **JPEG**: Using `jpeg.Encode` and `jpeg.Decode` for encoding and decoding.
- **GIF**: Using `gif.EncodeAll` and `gif.DecodeAll` for encoding and decoding.

## Notes

- The library requires the `golang.org/x/image/webp` package for WebP support.
- The `github.com/golang/freetype` package is used for rendering text with TrueType fonts.

## License

This library is open source and available under the MIT License.