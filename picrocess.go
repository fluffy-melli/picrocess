package picrocess

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type RGBA struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func NewRGBA(r, g, b uint8, a ...uint8) *RGBA {
	if len(a) == 0 {
		a = append(a, 255)
	}
	return &RGBA{
		R: r,
		G: g,
		B: b,
		A: a[0],
	}
}

type Rect struct {
	W1 uint
	H1 uint
	W2 uint
	H2 uint
}

func NewRect(w1, h1, w2, h2 uint) *Rect {
	return &Rect{
		W1: w1,
		H1: h1,
		W2: w2,
		H2: h2,
	}
}

func (r *Rect) Dx() uint {
	return r.W2 - r.W1
}

func (r *Rect) Dy() uint {
	return r.H2 - r.H1
}

type Offset struct {
	W uint
	H uint
}

func NewOffset(w, h uint) *Offset {
	return &Offset{
		W: w,
		H: h,
	}
}

func LoadFont(filename string) (*truetype.Font, error) {
	fontBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	fontFace, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return fontFace, nil
}

type Image struct {
	Width  uint
	Height uint
	Pixel  map[uint]map[uint]*RGBA // X / Y
}

func NewImage(w, h uint, color *RGBA) *Image {
	var respond = Image{
		Width:  w,
		Height: h,
		Pixel:  make(map[uint]map[uint]*RGBA),
	}
	for x := range respond.Pixel {
		respond.Pixel[x] = make(map[uint]*RGBA)
		for y := range x {
			respond.Pixel[x][y] = color
		}
	}
	return &respond
}

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

func (i *Image) At(x, y uint) *RGBA {
	if x > i.Width || y > i.Height {
		return &RGBA{0, 0, 0, 0}
	}
	pixel := i.Pixel[x][y]
	if pixel == nil {
		pixel = &RGBA{0, 0, 0, 0}
	}
	return pixel
}

func (i *Image) Set(x, y uint, c *RGBA) {
	if i.Pixel[x] == nil {
		i.Pixel[x] = map[uint]*RGBA{}
	}
	i.Pixel[x][y] = c
}

func (i *Image) Overlay(i2 *Image, o *Offset) {
	for x := range i2.Pixel {
		for y := range x {
			pixel := i2.At(x, y)
			if pixel == nil {
				pixel = &RGBA{0, 0, 0, 0}
			}
			destPixel := i.At(o.W+x, o.H+y)
			if destPixel == nil {
				destPixel = &RGBA{0, 0, 0, 0}
			}
			if pixel.A == 255 && destPixel.A == 255 {
				i.Set(o.W+x, o.H+y, pixel)
				continue
			}
			alpha := float64(pixel.A) / 255.0
			blendR := (1-alpha)*float64(destPixel.R) + alpha*float64(pixel.R)
			blendG := (1-alpha)*float64(destPixel.G) + alpha*float64(pixel.G)
			blendB := (1-alpha)*float64(destPixel.B) + alpha*float64(pixel.B)
			i.Set(o.W+x, o.H+y, &RGBA{
				R: uint8(blendR),
				G: uint8(blendG),
				B: uint8(blendB),
				A: pixel.A,
			})
		}
	}
}

func (i *Image) Resize(w, h uint) {
	newPixel := make(map[uint]map[uint]*RGBA)
	for x := range i.Pixel {
		newPixel[x] = make(map[uint]*RGBA)
		for y := range x {
			srcX := x * i.Width / w
			srcY := y * i.Height / h
			pixel := i.At(srcX, srcY)
			if pixel == nil {
				pixel = &RGBA{0, 0, 0, 0}
			}
			newPixel[x][y] = pixel
		}
	}
	i.Pixel = newPixel
	i.Width = w
	i.Height = h
}

func (i *Image) Crop(r *Rect) *Image {
	cropped := &Image{
		Width:  r.Dx(),
		Height: r.Dy(),
		Pixel:  make(map[uint]map[uint]*RGBA),
	}
	for x := range cropped.Pixel {
		cropped.Pixel[x] = make(map[uint]*RGBA)
		for y := range x {
			srcX := r.W1 + x
			srcY := r.H1 + y
			cropped.Pixel[x][y] = i.At(srcX, srcY)
		}
	}
	return cropped
}

func (i *Image) DrawString(font *truetype.Font, c *RGBA, o *Offset, size float64, text string) error {
	img := i.Render()
	pt := freetype.Pt(int(o.W), int(o.H)+int(size))
	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(font)
	ctx.SetFontSize(size)
	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)
	ctx.SetSrc(&image.Uniform{C: color.RGBA{c.R, c.G, c.B, c.A}})
	_, err := ctx.DrawString(text, pt)
	if err != nil {
		return err
	}
	for x := range i.Pixel {
		if i.Pixel[x] == nil {
			i.Pixel[x] = make(map[uint]*RGBA)
		}
		for y := range x {
			r, g, b, a := img.RGBAAt(int(x), int(y)).RGBA()
			i.Pixel[x][y] = &RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
		}
	}
	return nil
}

func (i *Image) Render() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, int(i.Width), int(i.Height)))
	for x := range i.Pixel {
		for y := range x {
			pixel := i.Pixel[x][y]
			if pixel == nil {
				pixel = &RGBA{0, 0, 0, 0}
			}
			img.Set(int(x), int(y), color.RGBA{pixel.R, pixel.G, pixel.B, pixel.A})
		}
	}
	return img
}

func Render(i *image.RGBA) *Image {
	width := uint(i.Bounds().Dx())
	height := uint(i.Bounds().Dy())
	img := &Image{
		Width:  width,
		Height: height,
		Pixel:  make(map[uint]map[uint]*RGBA),
	}
	for x := uint(0); x < width; x++ {
		img.Pixel[x] = make(map[uint]*RGBA)
	}
	for x := 0; x < int(width); x++ {
		for y := 0; y < int(height); y++ {
			c := i.At(x, y)
			rgba, ok := c.(color.RGBA)
			if !ok {
				rgba = color.RGBA{0, 0, 0, 0}
			}
			img.Pixel[uint(x)][uint(y)] = &RGBA{
				R: rgba.R,
				G: rgba.G,
				B: rgba.B,
				A: rgba.A,
			}
		}
	}
	return img
}

func (i *Image) ToPNG() ([]byte, error) {
	img := i.Render()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (i *Image) ToJPG(quality int) ([]byte, error) {
	img := i.Render()
	var buf bytes.Buffer
	opt := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&buf, img, opt); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

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
