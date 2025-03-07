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

type RGBA struct {
	R, G, B, A uint8
}

func NewRGBA(r, g, b uint8, a ...uint8) RGBA {
	alpha := uint8(255)
	if len(a) > 0 {
		alpha = a[0]
	}
	return RGBA{r, g, b, alpha}
}

type Rect struct {
	W1, H1, W2, H2 uint
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

func NewOffset(w, h uint) Offset {
	return Offset{
		W: w,
		H: h,
	}
}

type Font struct {
	face *truetype.Font
}

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

func (i *Image) At(x, y uint) RGBA {
	if x >= i.Width || y >= i.Height {
		return RGBA{0, 0, 0, 0}
	}
	return i.Pixel[x][y]
}

func (i *Image) Set(x, y uint, c RGBA) {
	if x >= i.Width || y >= i.Height {
		return
	}
	i.Pixel[x][y] = c
}

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

func (i *Image) Crop(r *Rect) *Image {
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

func (i *Image) ToPNGByte() ([]byte, error) {
	buffer, err := i.ToPNGBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (i *Image) ToJPGByte(quality int) ([]byte, error) {
	buffer, err := i.ToJPGBuffer(quality)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (i *Image) ToPNGBuffer() (*bytes.Buffer, error) {
	img := i.Render()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return &buf, nil
}

func (i *Image) ToJPGBuffer(quality int) (*bytes.Buffer, error) {
	img := i.Render()
	var buf bytes.Buffer
	opt := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&buf, img, opt); err != nil {
		return nil, err
	}
	return &buf, nil
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

type GIF struct {
	Delay []int
	Image []*image.RGBA
}

func NewGIF() *GIF {
	return &GIF{
		Delay: make([]int, 0),
		Image: make([]*image.RGBA, 0),
	}
}

func (gf *GIF) Append(image *Image, delay int) {
	gf.Delay = append(gf.Delay, delay)
	gf.Image = append(gf.Image, image.Render())
}

func (i *GIF) ToGIFByte() ([]byte, error) {
	buffer, err := i.ToGIFBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (gf *GIF) ToGIFBuffer() (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gifImages := make([]*image.Paletted, len(gf.Image))
	disposal := make([]byte, len(gf.Image))
	for i, img := range gf.Image {
		gifImages[i] = image.NewPaletted(img.Bounds(), Palette(img))
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

func Palette(frame *image.RGBA) color.Palette {
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
	if len(colors) > 256 {
		colors = colors[:256]
	}
	return colors
}

func NewQRCode(content string, size int, fgColor RGBA, bgColor RGBA) (*Image, error) {
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
