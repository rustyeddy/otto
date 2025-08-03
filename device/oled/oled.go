package oled

import (
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"math"
	"os"
	"time"

	"github.com/nfnt/resize"
	"github.com/rustyeddy/otto/device"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/devices/v3/ssd1306/image1bit"
	"periph.io/x/host/v3"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Bit is used to turn on or off a Bit on the ssd1306 OLED display
type Bit bool

const (
	On  Bit = true
	Off Bit = false
)

type OLED struct {
	*device.Device
	Dev        *ssd1306.Dev
	Height     int
	Width      int
	Font       *basicfont.Face
	Background *image1bit.VerticalLSB

	bus  string
	addr int
}

func New(name string, width, height int) (*OLED, error) {
	d := &OLED{
		Height: height,
		Width:  width,
		bus:    "/dev/i2c-1",
		addr:   0x3c,
	}
	d.Device = device.NewDevice(name)
	if device.IsMock() {
		return d, nil
	}

	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		return nil, err
	}

	// Open a handle to the first available I²C bus:
	bus, err := i2creg.Open("/dev/i2c-1")
	if err != nil {
		return nil, err
	}

	// Open a handle to a ssd1306 connected on the I²C bus:
	opts := &ssd1306.DefaultOpts
	opts.H = height
	opts.W = width

	d.Dev, err = ssd1306.NewI2C(bus, opts)
	if err != nil {
		return nil, err
	}

	d.Background = image1bit.NewVerticalLSB(image.Rect(0, 0, opts.W, opts.H))
	return d, nil
}

func (d *OLED) Clear() {
	// got to be a better way
	d.Rectangle(0, 0, d.Width, d.Height, Off)
}

func (d *OLED) Draw() error {
	err := d.Dev.Draw(d.Background.Bounds(), d.Background, image.Point{})
	if err != nil {
		fmt.Println("ERROR - OLED Draw: ", err)
		return err
	}
	return nil
}
func (d *OLED) Rectangle(x0, y0, x1, y1 int, value Bit) {
	d.Clip(&x0, &y0, &x1, &y1)

	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}

	for x := x0; x < x1; x++ {
		for y := y0; y < y1; y++ {
			d.SetBit(x, y, value)
		}
	}
}

func (d *OLED) Line(x0, y0, len, width int, value Bit) {
	x1 := x0 + len
	y1 := y0 + width
	d.Rectangle(x0, y0, x1, y1, value)
}

func (d *OLED) Diagonal(x0, y0, x1, y1 int, value Bit) {
	d.Clip(&x0, &y0, &x1, &y1)

	xf0 := float64(x0)
	xf1 := float64(x1)
	yf0 := float64(y0)
	yf1 := float64(y1)

	l := (xf1 - xf0)
	h := (yf1 - yf0)

	var slope float64
	if l > h {
		slope = h / l
	} else {
		slope = l / h
	}

	if l >= h {
		for x := xf0; x < xf1; x++ {
			y := slope*(x-xf0) + yf0
			d.SetBit(int(math.Round(x)), int(math.Round(y)), value)
		}
	} else {
		for y := yf0; y < yf1; y++ {
			x := slope*(y-yf0) + xf0
			d.SetBit(int(math.Round(x)), int(math.Round(y)), value)
		}
	}
}

func (d *OLED) Scroll(o ssd1306.Orientation, rate ssd1306.FrameRate, startLine, endLine int) error {
	return d.Dev.Scroll(o, rate, startLine, endLine)
}

func (d *OLED) StopScroll() error {
	return d.Dev.StopScroll()
}

func (d *OLED) SetBit(x, y int, value Bit) {
	d.Background.SetBit(x, y, image1bit.Bit(value))
}

func (d *OLED) Clip(x0, y0, x1, y1 *int) {
	if x0 != nil {
		if *x0 < 0 {
			*x0 = 0
		}
		if *x0 > d.Width {
			*x0 = d.Width
		}
	}
	if x1 != nil {
		if *x1 < 0 {
			*x1 = 0
		}
		if *x1 > d.Width {
			*x1 = d.Width
		}
	}

	if y0 != nil {
		if *y0 < 0 {
			*y0 = 0
		}
		if *y0 > d.Height {
			*y0 = d.Height
		}
	}

	if y1 != nil {
		if *y1 < 0 {
			*y1 = 0
		}
		if *y1 > d.Height {
			*y1 = d.Height
		}
	}
}

func (d *OLED) DrawString(x, y int, str string) {
	d.Font = basicfont.Face7x13
	drawer := &font.Drawer{
		Dst:  d.Background,
		Src:  image.White,
		Face: d.Font,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	drawer.DrawString(str)
}

func (d *OLED) AnimatedGIF(fname string, done <-chan time.Time) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}

	g, err := gif.DecodeAll(f)
	f.Close()
	if err != nil {
		return err
	}

	// Converts every frame to image.Gray and resize them:
	imgs := make([]*image.Gray, len(g.Image))
	for i := range g.Image {
		imgs[i] = d.convertAndResizeAndCenter(g.Image[i])
	}

	// OLED the frames in a loop:
	for i := 0; ; i++ {
		select {
		case <-done:
			return nil

		default:
			index := i % len(imgs)
			c := time.After(time.Duration(10*g.Delay[index]) * time.Millisecond)
			img := imgs[index]
			d.Dev.Draw(img.Bounds(), img, image.Point{})
			<-c
		}
	}
	return nil
}

// convertAndResizeAndCenter takes an image, resizes and centers it on a
// image.Gray of size w*h.
func (d *OLED) convertAndResizeAndCenter(src image.Image) *image.Gray {
	w := d.Width
	h := d.Height
	src = resize.Thumbnail(uint(w), uint(h), src, resize.Bicubic)
	img := image.NewGray(image.Rect(0, 0, w, h))
	r := src.Bounds()
	r = r.Add(image.Point{(w - r.Max.X) / 2, (h - r.Max.Y) / 2})
	draw.Draw(img, r, src, image.Point{}, draw.Src)
	return img
}
