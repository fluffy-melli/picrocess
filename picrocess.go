package picrocess

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"net/http"
	"os"

	_ "golang.org/x/image/webp"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/skip2/go-qrcode"
	"golang.org/x/image/math/fixed"
)

var ASCII_CHARS = []string{
	" ", ".", "'", "`", "^", "\"", ",", ";", ":", "I", "l", "!", "i", ">", "<", "~", "+", "_", "-",
	"?", "]", "[", "}", "{", "1", ")", "(", "|", "\\", "/", "t", "f", "j", "r", "x", "n", "u", "v",
	"c", "z", "X", "Y", "U", "J", "C", "L", "Q", "0", "O", "Z", "m", "w", "q", "p", "d", "b", "k", "h",
	"a", "o", "*", "#", "M", "W", "&", "8", "%", "B", "@", "$",
}

type RGBA struct {
	R, G, B, A uint8
}

// NewRGBA creates an RGBA color struct using the provided red (r), green (g), blue (b) values, and an optional alpha (a) value.
// If the alpha value is not provided, it defaults to 255 (fully opaque).
//
// r: Red value (0-255)
// g: Green value (0-255)
// b: Blue value (0-255)
// a: (Optional) Alpha value (0-255), defaults to 255 if not provided
//
// Returns: An RGBA struct initialized with the given color values and alpha.
func NewRGBA(r, g, b uint8, a ...uint8) RGBA {
	alpha := uint8(255)
	if len(a) > 0 {
		alpha = a[0]
	}
	return RGBA{r, g, b, alpha}
}

// Brightness calculates the perceived brightness of the color.
// It averages the RGB values and then adjusts the result based on the alpha value.
// The alpha value influences the final brightness, making transparent pixels darker.
//
// Returns: An integer representing the brightness of the color (0 to 255).
func (c RGBA) Brightness() int {
	r := int(c.R)
	g := int(c.G)
	b := int(c.B)
	a := int(c.A)
	if a == 0 {
		return 0
	}
	brightness := int(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
	brightness = brightness * a / 255
	return brightness
}

type Rect struct {
	W1, H1, W2, H2 uint
}

// NewRect creates a new Rect struct using the provided width (w1, w2) and height (h1, h2) values.
// It returns a pointer to the newly created Rect.
//
// w1: The width of the first point
// h1: The height of the first point
// w2: The width of the second point
// h2: The height of the second point
//
// Returns: A pointer to a Rect struct initialized with the given dimensions.
func NewRect(w1, h1, w2, h2 uint) Rect {
	return Rect{
		W1: w1,
		H1: h1,
		W2: w2,
		H2: h2,
	}
}

// Dx returns the horizontal distance (width) between the two points of the Rect.
// It calculates the difference between the second width (W2) and the first width (W1).
//
// Returns: The horizontal distance between W2 and W1.
func (r *Rect) Dx() uint {
	return r.W2 - r.W1
}

// Dy returns the vertical distance (height) between the two points of the Rect.
// It calculates the difference between the second height (H2) and the first height (H1).
//
// Returns: The vertical distance between H2 and H1.
func (r *Rect) Dy() uint {
	return r.H2 - r.H1
}

type Offset struct {
	W uint
	H uint
}

// NewOffset creates a new Offset struct using the provided width (w) and height (h) values.
// It returns the newly created Offset struct.
//
// w: The width value
// h: The height value
//
// Returns: An Offset struct initialized with the given width and height.
func NewOffset(w, h uint) Offset {
	return Offset{
		W: w,
		H: h,
	}
}

type Font struct {
	face *truetype.Font
}

// LoadFont loads a font from the specified file and returns a pointer to a Font struct.
// If there is an error reading the file or parsing the font, it returns an error.
//
// filename: The path to the font file to load.
//
// Returns: A pointer to a Font struct containing the parsed font, or an error if any issue occurs.
func LoadFont(filename string) (*Font, error) {
	fontBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	fontFace, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return &Font{face: fontFace}, nil
}

// TextSize calculates the width and height of the given text when rendered with the specified font size.
// It returns the width and height of the text in pixels.
//
// size: The font size to use for rendering the text.
// text: The text to measure.
//
// Returns: The width and height of the text in pixels.
func (f *Font) TextSize(size float64, text string) (uint, uint) {
	var width uint
	var height uint
	fontFace := truetype.NewFace(f.face, &truetype.Options{Size: size})
	for _, c := range text {
		bounds, advance, _ := fontFace.GlyphBounds(c)
		width += uint(advance.Ceil())
		if bounds.Max.Y > fixed.Int26_6(height) {
			height = uint(bounds.Max.Y.Ceil())
		}
	}
	return width, height
}

type Image struct {
	Width, Height uint
	Pixel         [][]RGBA // X / Y
}

// NewImage creates a new Image struct with the specified width (w), height (h), and initial color (color).
// It initializes the pixel data as a 2D slice and sets each pixel to the specified color.
//
// w: The width of the image.
// h: The height of the image.
// color: The color to fill each pixel in the image.
//
// Returns: A pointer to a new Image struct initialized with the given dimensions and color.
func NewImage(w, h uint, color RGBA) *Image {
	var respond = Image{
		Width:  w,
		Height: h,
		Pixel:  make([][]RGBA, w),
	}
	for x := uint(0); x < w; x++ {
		respond.Pixel[x] = make([]RGBA, h)
		for y := uint(0); y < h; y++ {
			respond.Pixel[x][y] = color
		}
	}
	return &respond
}

// LoadImage loads an image from a file, decodes it, and returns an Image struct.
// It returns an error if the file cannot be opened or the image cannot be decoded.
//
// filename: The path to the image file to load.
//
// Returns: A pointer to an Image struct containing the decoded image, or an error if any issue occurs.
func LoadImage(filename string) (*Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return Render(rgba), nil
}

// ImageURL loads an image from a URL, decodes it, and returns an Image struct.
// It returns an error if the HTTP request fails or the image cannot be decoded.
//
// url: The URL of the image to load.
//
// Returns: A pointer to an Image struct containing the decoded image, or an error if any issue occurs.
func ImageURL(url string) (*Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return Render(rgba), nil
}

// At returns the color of the pixel at the specified coordinates (x, y) in the image.
// If the coordinates are out of bounds, it returns a transparent black color (RGBA{0, 0, 0, 0}).
//
// x: The x-coordinate of the pixel.
// y: The y-coordinate of the pixel.
//
// Returns: The color of the pixel at the given coordinates.
func (i *Image) At(x, y uint) RGBA {
	if x >= i.Width || y >= i.Height {
		return RGBA{0, 0, 0, 0}
	}
	return i.Pixel[x][y]
}

// Set sets the color of the pixel at the specified coordinates (x, y) in the image.
// If the coordinates are out of bounds, it does nothing.
//
// x: The x-coordinate of the pixel.
// y: The y-coordinate of the pixel.
// c: The color to set at the given coordinates.
func (i *Image) Set(x, y uint, c RGBA) {
	if x >= i.Width || y >= i.Height {
		return
	}
	i.Pixel[x][y] = c
}

// The To function searches for a specific color (b) in the image and replaces it with a new color (a).
//
// Parameters:
//   - b: The color to be replaced (target color)
//   - a: The color to replace with (new color)
//
// This function iterates through all the pixels of the image and if a pixel matches color b,
// it replaces that pixel with color a.
func (i *Image) To(b, a RGBA) {
	for x := range i.Pixel {
		for y := range i.Pixel[x] {
			if i.Pixel[x][y] == b {
				i.Pixel[x][y] = a
			}
		}
	}
}

// Overlay overlays the second image (i2) onto the first image (i) at the specified offset (o).
// The pixel blending is done based on the alpha channel (transparency) of the pixels.
//
// i2: The image to overlay on top of the current image (i).
// o: The offset to position the second image on top of the first image (i).
//
// The function blends the pixels based on the alpha values. It uses the formula for alpha blending
// when both pixels are partially transparent, while fully opaque pixels are copied directly.
func (i *Image) Overlay(i2 *Image, o Offset) {
	for x := range i2.Pixel {
		for y := range i2.Pixel[x] {
			pixel := i2.At(uint(x), uint(y))
			destPixel := i.At(o.W+uint(x), o.H+uint(y))
			if pixel.A == 255 && destPixel.A == 255 {
				i.Set(o.W+uint(x), o.H+uint(y), pixel)
				continue
			}
			if pixel.A == 255 && destPixel.A == 0 {
				i.Set(o.W+uint(x), o.H+uint(y), pixel)
				continue
			}
			if pixel.A == 0 && destPixel.A == 255 {
				i.Set(o.W+uint(x), o.H+uint(y), destPixel)
				continue
			}
			alpha := float64(pixel.A) / 255.0
			blendR := (1-alpha)*float64(destPixel.R) + alpha*float64(pixel.R)
			blendG := (1-alpha)*float64(destPixel.G) + alpha*float64(pixel.G)
			blendB := (1-alpha)*float64(destPixel.B) + alpha*float64(pixel.B)
			i.Set(o.W+uint(x), o.H+uint(y), RGBA{
				R: uint8(blendR),
				G: uint8(blendG),
				B: uint8(blendB),
				A: pixel.A,
			})
		}
	}
}

// Resize resizes the image to the specified width (w) and height (h) using nearest-neighbor scaling.
// It creates a new pixel array with the new size and maps the pixels from the original image to the resized one.
//
// w: The new width of the image.
// h: The new height of the image.
func (i *Image) Resize(w, h uint) {
	newPixel := make([][]RGBA, w)
	for x := range newPixel {
		newPixel[x] = make([]RGBA, h)
		for y := range newPixel[x] {
			srcX := uint(x) * i.Width / w
			srcY := uint(y) * i.Height / h
			pixel := i.At(srcX, srcY)
			newPixel[x][y] = pixel
		}
	}
	i.Pixel = newPixel
	i.Width = w
	i.Height = h
}

// Crop crops a section of the image based on the given rectangle (r).
// It returns a new image that represents the cropped region.
//
// r: The rectangle defining the region to crop.
//
// Returns: A new Image struct containing the cropped region.
func (i *Image) Crop(r Rect) *Image {
	cropped := &Image{
		Width:  r.Dx(),
		Height: r.Dy(),
		Pixel:  make([][]RGBA, r.Dx()),
	}
	for x := range cropped.Pixel {
		cropped.Pixel[x] = make([]RGBA, r.Dy())
		for y := range cropped.Pixel[x] {
			srcX := r.W1 + uint(x)
			srcY := r.H1 + uint(y)
			cropped.Pixel[x][y] = i.At(srcX, srcY)
		}
	}
	return cropped
}

// Round applies a rounding effect to the corners of the image by setting pixels outside a circular region to transparent.
// The rounded corners are based on the specified pixel radius (px).
//
// px: The radius of the rounded corner, in pixels.
//
// This function modifies the image by setting the pixels outside the rounded area to transparent.
func (i *Image) Round(px uint) {
	for x := range i.Pixel {
		for y := range i.Pixel[x] {
			if uint(x) >= px && uint(x) <= i.Width-px || uint(y) >= px && uint(y) <= i.Width-px {
				continue
			}
			var dx float64
			var dy float64
			if uint(x) <= px && uint(y) <= px {
				dx = float64(px)
				dy = float64(px)
			} else if uint(x) <= px && uint(y) > i.Width-px {
				dx = float64(px)
				dy = float64(i.Height - px)
			} else if uint(x) >= i.Width-px && uint(y) <= px {
				dx = float64(i.Width - px)
				dy = float64(px)
			} else {
				dx = float64(i.Width - px)
				dy = float64(i.Height - px)
			}
			distance := math.Sqrt(math.Pow(float64(x)-dx, 2) + math.Pow(float64(y)-dy, 2))
			if distance > float64(px) {
				i.Set(uint(x), uint(y), RGBA{0, 0, 0, 0})
			}
		}
	}
}

// Rotate90 rotates the image 90 degrees clockwise.
// It creates a new pixel array, rotates each pixel by 90 degrees,
// and then updates the original image with the new rotated pixel data.
func (i *Image) Rotate90() {
	newPixel := make([][]RGBA, i.Height)
	for x := range newPixel {
		newPixel[x] = make([]RGBA, i.Width)
	}
	for x := range i.Pixel {
		for y := range i.Pixel[x] {
			pixel := i.At(uint(x), uint(y))
			newPixel[y][i.Height-1-uint(x)] = pixel
		}
	}
	i.Pixel = newPixel
	i.Width, i.Height = i.Height, i.Width
}

// RotateMinus90 rotates the image 90 degrees counterclockwise (anti-clockwise).
// It creates a new pixel array, rotates each pixel by -90 degrees,
// and then updates the original image with the new rotated pixel data.
func (i *Image) RotateMinus90() {
	newPixel := make([][]RGBA, i.Height)
	for x := range newPixel {
		newPixel[x] = make([]RGBA, i.Width)
	}
	for x := 0; x < int(i.Width); x++ {
		for y := 0; y < int(i.Height); y++ {
			pixel := i.At(uint(x), uint(y))
			newPixel[i.Height-1-uint(y)][x] = pixel
		}
	}
	i.Pixel = newPixel
	i.Width, i.Height = i.Height, i.Width
}

// FlipHorizontal flips the image horizontally (left to right).
// It mirrors the pixels in each row.
func (i *Image) FlipHorizontal() {
	for y := 0; y < int(i.Height); y++ {
		for x := 0; x < int(i.Width)/2; x++ {
			leftPixel := i.At(uint(x), uint(y))
			rightPixel := i.At(uint(i.Width-1-uint(x)), uint(y))
			i.Pixel[x][y] = rightPixel
			i.Pixel[i.Width-1-uint(x)][y] = leftPixel
		}
	}
}

// FlipVertical flips the image vertically (top to bottom).
// It mirrors the pixels in each column.
func (i *Image) FlipVertical() {
	for x := 0; x < int(i.Width); x++ {
		for y := 0; y < int(i.Height)/2; y++ {
			topPixel := i.At(uint(x), uint(y))
			bottomPixel := i.At(uint(x), uint(i.Height-1-uint(y)))
			i.Pixel[x][y] = bottomPixel
			i.Pixel[x][i.Height-1-uint(y)] = topPixel
		}
	}
}

// Text draws the specified text on the image using the provided font, color, offset, and size.
// It renders the text at the specified offset (o) on the image (i), with the given font size and color.
//
// font: The Font object to use for rendering the text.
// c: The color (RGBA) to use for the text.
// o: The offset specifying where to draw the text on the image.
// size: The font size to use for rendering the text.
// text: The string of text to be drawn on the image.
//
// Returns: An error if there is an issue rendering the text.
func (i *Image) Text(font *Font, c RGBA, o Offset, size float64, text string) error {
	img := i.Render()
	pt := freetype.Pt(int(o.W), int(o.H)+int(size))
	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(font.face)
	ctx.SetFontSize(size)
	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)
	ctx.SetSrc(&image.Uniform{C: color.RGBA{c.R, c.G, c.B, c.A}})
	_, err := ctx.DrawString(text, pt)
	if err != nil {
		return err
	}
	for x := range i.Pixel {
		for y := range i.Pixel[x] {
			r, g, b, a := img.RGBAAt(int(x), int(y)).RGBA()
			i.Pixel[x][y] = RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
		}
	}
	return nil
}

func pointToLineDistance(x1, y1, x2, y2, px, py float64) float64 {
	vx, vy := x2-x1, y2-y1
	lenSq := vx*vx + vy*vy
	if lenSq == 0 {
		dx, dy := px-x1, py-y1
		return math.Sqrt(dx*dx + dy*dy)
	}
	wx, wy := px-x1, py-y1
	t := (wx*vx + wy*vy) / lenSq
	if t < 0 {
		return math.Sqrt(wx*wx + wy*wy)
	} else if t > 1 {
		dx, dy := px-x2, py-y2
		return math.Sqrt(dx*dx + dy*dy)
	}
	projX := x1 + t*vx
	projY := y1 + t*vy
	dx, dy := px-projX, py-projY
	return math.Sqrt(dx*dx + dy*dy)
}

// Line draws a line on the image from point (r.W1, r.H1) to point (r.W2, r.H2) with the specified color (c)
// and thickness. It iterates over the pixels of the image and sets the pixel color to the specified color
// if the pixel is within the thickness of the line.
//
// r: The rectangle defining the start and end points of the line (W1, H1) to (W2, H2).
// c: The color (RGBA) to use for the line.
// thickness: The thickness of the line.
func (i *Image) Line(r Rect, c RGBA, thickness float64) {
	for x := range i.Pixel {
		for y := range i.Pixel[x] {
			distance := pointToLineDistance(float64(r.W1), float64(r.H1), float64(r.W2), float64(r.H2), float64(x), float64(y))
			if distance <= thickness/2 {
				i.Set(uint(x), uint(y), c)
			}
		}
	}
}

// The Ascii function converts the image into an ASCII art representation and returns it as a string.
// The image is resized to the given width (w) and height (h), then rotated 90 degrees,
// and ASCII characters corresponding to the brightness of each pixel are selected for output.
// `length` specifies how many times each ASCII character should be repeated.
func (i *Image) Ascii(w, h, length uint) string {
	img := *i
	img.Resize(w, h)
	img.Rotate90()
	img.FlipVertical()
	respond := ""
	for x := range img.Pixel {
		for y := range img.Pixel[x] {
			pixel := img.Pixel[x][y]
			brightness := pixel.Brightness()
			index := brightness * (len(ASCII_CHARS) - 1) / 255
			for i := 0; i < int(length); i++ {
				respond += ASCII_CHARS[index]
			}
		}
		respond += "\n"
	}
	return respond
}

// Render converts the custom Image structure to an image.RGBA object,
// mapping each pixel in the custom Image to the corresponding color in the RGBA image.
//
// Returns: A pointer to an image.RGBA object representing the image.
func (i *Image) Render() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, int(i.Width), int(i.Height)))
	for x := range i.Pixel {
		if i.Pixel[x] == nil {
			continue
		}
		for y := range i.Pixel[x] {
			pixel := i.Pixel[x][y]
			img.Set(int(x), int(y), color.RGBA{pixel.R, pixel.G, pixel.B, pixel.A})
		}
	}
	return img
}

// Render converts an image.RGBA object back into the custom Image structure,
// extracting pixel values from the image and storing them in the custom Image format.
//
// i: The image.RGBA object to convert into the custom Image format.
//
// Returns: A pointer to an Image object representing the custom image format.
func Render(i *image.RGBA) *Image {
	width := uint(i.Bounds().Dx())
	height := uint(i.Bounds().Dy())
	img := &Image{
		Width:  width,
		Height: height,
		Pixel:  make([][]RGBA, width),
	}
	for x := 0; x < int(width); x++ {
		img.Pixel[x] = make([]RGBA, height)
		for y := 0; y < int(height); y++ {
			c := i.At(x, y)
			rgba, ok := c.(color.RGBA)
			if !ok {
				rgba = color.RGBA{0, 0, 0, 0}
			}
			img.Pixel[uint(x)][uint(y)] = RGBA{
				R: rgba.R,
				G: rgba.G,
				B: rgba.B,
				A: rgba.A,
			}
		}
	}
	return img
}

// ToPNGByte converts the Image to a PNG byte slice.
func (i *Image) ToPNGByte() ([]byte, error) {
	buffer, err := i.ToPNGBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// ToJPGByte converts the Image to a JPG byte slice with the specified quality.
func (i *Image) ToJPGByte(quality int) ([]byte, error) {
	buffer, err := i.ToJPGBuffer(quality)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// ToPNGBuffer converts the Image to a PNG format and returns a bytes.Buffer.
func (i *Image) ToPNGBuffer() (*bytes.Buffer, error) {
	img := i.Render()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return &buf, nil
}

// ToJPGBuffer converts the Image to a JPG format with the specified quality and returns a bytes.Buffer.
func (i *Image) ToJPGBuffer(quality int) (*bytes.Buffer, error) {
	img := i.Render()
	var buf bytes.Buffer
	opt := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&buf, img, opt); err != nil {
		return nil, err
	}
	return &buf, nil
}

// SaveAsPNG saves the Image as a PNG file to the specified path.
func (i *Image) SaveAsPNG(filename string) error {
	img := i.Render()
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	err = png.Encode(file, img)
	if err != nil {
		return err
	}
	return nil
}

// SaveAsJPG saves the Image as a JPG file to the specified path with the specified quality.
func (i *Image) SaveAsJPG(filename string, quality int) error {
	img := i.Render()
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	opt := &jpeg.Options{Quality: quality}
	return jpeg.Encode(file, img, opt)
}

type GIF struct {
	Delay []int
	Image []*image.RGBA
}

// NewGIF creates and returns a new GIF object.
func NewGIF() *GIF {
	return &GIF{
		Delay: make([]int, 0),
		Image: make([]*image.RGBA, 0),
	}
}

// Append adds a new frame (image) to the GIF with a specified delay.
func (gf *GIF) Append(image *Image, delay int) {
	gf.Delay = append(gf.Delay, delay)
	gf.Image = append(gf.Image, image.Render())
}

// ToGIFByte converts the GIF object to a byte slice in GIF format.
func (i *GIF) ToGIFByte() ([]byte, error) {
	buffer, err := i.ToGIFBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// ToGIFBuffer converts the GIF object into a bytes buffer containing the GIF data.
func (gf *GIF) ToGIFBuffer() (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gifImages := make([]*image.Paletted, len(gf.Image))
	disposal := make([]byte, len(gf.Image))
	for i, img := range gf.Image {
		gifImages[i] = image.NewPaletted(img.Bounds(), Palette(img, 256*256*256))
		draw.Draw(gifImages[i], img.Bounds(), img, image.Point{}, draw.Src)
		disposal[i] = gif.DisposalBackground
	}
	err := gif.EncodeAll(&buf, &gif.GIF{
		Image:    gifImages,
		Delay:    gf.Delay,
		Disposal: disposal,
	})
	if err != nil {
		return nil, err
	}
	return &buf, nil
}

// SaveAsGIF saves the GIF data to a file.
func (i *GIF) SaveAsGIF(filename string) error {
	data, err := i.ToGIFByte()
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// Palette generates a color palette for the given RGBA frame, with a customizable limit on the number of colors.
// It extracts unique colors from the image and returns a color.Palette.
// If the number of colors exceeds the limit, the palette is truncated to the specified limit.
func Palette(frame *image.RGBA, limit int) color.Palette {
	colorSet := make(map[color.RGBA]struct{})
	for y := 0; y < frame.Bounds().Dy(); y++ {
		for x := 0; x < frame.Bounds().Dx(); x++ {
			colorSet[frame.RGBAAt(x, y)] = struct{}{}
		}
	}
	var colors []color.Color
	for c := range colorSet {
		colors = append(colors, c)
	}
	if len(colors) > limit {
		colors = colors[:limit]
	}
	return colors
}

// NewQRCode generates a new QR code image from the given content, with customizable foreground and background colors.
// It creates a QR code of the specified size and color options, and returns the generated image.
func NewQRCode(bgColor, fgColor RGBA, size int, content string) (*Image, error) {
	qr, err := qrcode.New(content, qrcode.High)
	if err != nil {
		return nil, err
	}
	qr.BackgroundColor = color.RGBA{
		R: bgColor.R,
		G: bgColor.G,
		B: bgColor.B,
		A: bgColor.A,
	}
	qr.ForegroundColor = color.RGBA{
		R: fgColor.R,
		G: fgColor.G,
		B: fgColor.B,
		A: fgColor.A,
	}
	binary, err := qr.PNG(size)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(binary)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Over)
	return Render(rgba), nil
}

type GrapeLayer struct {
	Value float64
}

type LineGrape struct {
	Value []float64
}

func NewLineGrape() *LineGrape {
	return &LineGrape{
		Value: make([]float64, 0),
	}
}

func (g *LineGrape) Append(v float64) {
	g.Value = append(g.Value, v)
}

func (g *LineGrape) Render() *Image {
	min := func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		minValue := values[0]
		for _, v := range values {
			if v < minValue {
				minValue = v
			}
		}
		return minValue
	}(g.Value)
	max := func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		maxValue := values[0]
		for _, v := range values {
			if v > maxValue {
				maxValue = v
			}
		}
		return maxValue
	}(g.Value)
	base := NewImage(700, 500, NewRGBA(255, 255, 255))
	base.Line(NewRect(30, 30, 30, 470), NewRGBA(120, 120, 120), 2)
	base.Line(NewRect(30, 30, 670, 30), NewRGBA(120, 120, 120), 2)
	base.Line(NewRect(30, 470, 670, 470), NewRGBA(120, 120, 120), 2)
	base.Line(NewRect(670, 30, 670, 470), NewRGBA(120, 120, 120), 2)
	lastX := uint(30)
	lastY := uint(0)
	for i := uint(0); i < 6; i++ {
		base.Line(NewRect(30, 440/6*(i+1)+30, 670, 440/6*(i+1)+30), NewRGBA(120, 120, 120), 1)
	}
	step := float64(640) / float64(len(g.Value))
	for i := range g.Value {
		x := uint(step*float64(i)) + 30
		y := 500 - (uint((g.Value[i]-min)/(max-min)*440) + 30)
		if i == 0 {
			lastY = y
		}
		base.Line(NewRect(lastX, lastY, x, y), NewRGBA(255, 0, 0), 2)
		if i != len(g.Value)-1 {
			base.Line(NewRect(x, 30, x, 470), NewRGBA(120, 120, 120), 1)
		}
		lastX = x
		lastY = y
	}
	return base
}
