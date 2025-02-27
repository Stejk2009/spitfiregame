package main

import (
	"image"
	"image/color"
	_ "image/png" // Import the PNG decoder
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

var (
	gameOverFont font.Face
)

func init() {
	rand.Seed(time.Now().UnixNano())

	// Load custom font
	tt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	gameOverFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    48,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

type Bullet struct {
	Image *ebiten.Image
	X, Y  float64
}

func NewBullet(x, y float64) *Bullet {
	file, err := os.Open("assets/images/Tiles/tile_0012.png")
	if err != nil {
		log.Fatalf("Failed to open bullet image file: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Failed to decode bullet image: %v", err)
	}

	ebitenImage := ebiten.NewImageFromImage(img)
	return &Bullet{
		Image: ebitenImage,
		X:     x,
		Y:     y,
	}
}

func (b *Bullet) Update() {
	b.Y -= 5
}

func (b *Bullet) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.X, b.Y)
	screen.DrawImage(b.Image, op)
}

type Item struct {
	Image *ebiten.Image
	X, Y  float64
	Scale float64
	Type  string
}

func NewItem(imagePath string, x, y float64, itemType string) *Item {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("Failed to open image file: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Failed to decode image: %v", err)
	}

	ebitenImage := ebiten.NewImageFromImage(img)
	scale := 0.5 + rand.Float64() // Scale between 0.5 and 1.5
	return &Item{
		Image: ebitenImage,
		X:     x,
		Y:     y,
		Scale: scale,
		Type:  itemType,
	}
}

func (i *Item) Update() {
	i.Y += 2
}

func (i *Item) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(i.Scale, i.Scale)
	op.GeoM.Translate(i.X, i.Y)
	screen.DrawImage(i.Image, op)
}

type Spitfire struct {
	Image     *ebiten.Image
	X, Y      float64
	Bullets   []*Bullet
	Items     []*Item
	Score     int
	HighScore int
	Health    float64
	GameOver  bool
}

func NewSpitfire() *Spitfire {
	file, err := os.Open("assets/images/plane.png")
	if err != nil {
		log.Fatalf("Failed to open image file: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Failed to decode image: %v", err)
	}

	ebitenImage := ebiten.NewImageFromImage(img)
	return &Spitfire{
		Image:     ebitenImage,
		X:         320 - float64(ebitenImage.Bounds().Dx())/2,
		Y:         480 - float64(ebitenImage.Bounds().Dy())*2 - 10, // Adjusted starting position
		Bullets:   []*Bullet{},
		Items:     []*Item{},
		Score:     0,
		HighScore: 0,
		Health:    1.0, // Full health
		GameOver:  false,
	}
}

func (s *Spitfire) Update() {
	if s.GameOver {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			s.Restart()
		}
		return
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		s.Y -= 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		s.Y += 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		s.X -= 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		s.X += 2
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		s.Fire()
	}

	for _, bullet := range s.Bullets {
		bullet.Update()
	}

	for _, item := range s.Items {
		item.Update()
	}

	s.CheckCollisions()
	s.RemoveOffScreenBullets()
	s.RemoveOffScreenItems()
	s.SpawnItems()
	s.CheckGameOver()
}

func (s *Spitfire) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1.8, 1.8)
	op.GeoM.Translate(s.X, s.Y)
	screen.DrawImage(s.Image, op)

	for _, bullet := range s.Bullets {
		bullet.Draw(screen)
	}

	for _, item := range s.Items {
		item.Draw(screen)
	}

	scoreText := "Score: " + strconv.Itoa(s.Score)
	highScoreText := "High Score: " + strconv.Itoa(s.HighScore)
	scoreTextWidth := text.BoundString(basicfont.Face7x13, scoreText).Dx()
	highScoreTextWidth := text.BoundString(basicfont.Face7x13, highScoreText).Dx()
	text.Draw(screen, scoreText, basicfont.Face7x13, (640-scoreTextWidth)/2, 20, color.White)
	text.Draw(screen, highScoreText, basicfont.Face7x13, (640-highScoreTextWidth)/2, 40, color.White)

	// Draw health bar
	healthBarHeight := int(s.Health * 100)
	ebitenutil.DrawRect(screen, 620, 480-float64(healthBarHeight), 10, float64(healthBarHeight), color.RGBA{255, 0, 0, 255})
	healthText := strconv.Itoa(int(s.Health*100)) + "%"
	text.Draw(screen, healthText, basicfont.Face7x13, 610, 480-healthBarHeight-10, color.White)

	if s.GameOver {
		gameOverText := "Game Over"
		gameOverTextWidth := text.BoundString(gameOverFont, gameOverText).Dx()
		text.Draw(screen, gameOverText, gameOverFont, (640-gameOverTextWidth)/2, 240, color.White)
		ebitenutil.DebugPrint(screen, "Press R to Restart")
	}
}

func (s *Spitfire) Fire() {
	bullet := NewBullet(s.X+float64(s.Image.Bounds().Dx())/2-4, s.Y)
	s.Bullets = append(s.Bullets, bullet)
}

func (s *Spitfire) CheckCollisions() {
	for i := len(s.Items) - 1; i >= 0; i-- {
		item := s.Items[i]
		for j := len(s.Bullets) - 1; j >= 0; j-- {
			bullet := s.Bullets[j]
			if bullet.X < item.X+float64(item.Image.Bounds().Dx())*item.Scale &&
				bullet.X+float64(bullet.Image.Bounds().Dx()) > item.X &&
				bullet.Y < item.Y+float64(item.Image.Bounds().Dy())*item.Scale &&
				bullet.Y+float64(bullet.Image.Bounds().Dy()) > item.Y {
				s.Items = append(s.Items[:i], s.Items[i+1:]...)
				s.Bullets = append(s.Bullets[:j], s.Bullets[j+1:]...)
				s.Score += 10
				if s.Score > s.HighScore {
					s.HighScore = s.Score
				}
				break
			}
		}
		if s.X < item.X+float64(item.Image.Bounds().Dx())*item.Scale &&
			s.X+float64(s.Image.Bounds().Dx()) > item.X &&
			s.Y < item.Y+float64(item.Image.Bounds().Dy())*item.Scale &&
			s.Y+float64(s.Image.Bounds().Dy()) > item.Y {
			if item.Type == "health" {
				s.Health = 1.0 // Restore to full health
			} else {
				s.Health -= 0.5
				if s.Health <= 0 {
					s.GameOver = true
				}
			}
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
		}
	}
}

func (s *Spitfire) RemoveOffScreenBullets() {
	for i := len(s.Bullets) - 1; i >= 0; i-- {
		bullet := s.Bullets[i]
		if bullet.Y < -float64(bullet.Image.Bounds().Dy()) {
			s.Bullets = append(s.Bullets[:i], s.Bullets[i+1:]...)
		}
	}
}

func (s *Spitfire) RemoveOffScreenItems() {
	for i := len(s.Items) - 1; i >= 0; i-- {
		item := s.Items[i]
		if item.Y > 480 {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
		}
	}
}

func (s *Spitfire) SpawnItems() {
	if rand.Float64() < 0.02 {
		x := rand.Float64() * 640
		itemImages := []string{
			"assets/images/Tiles/tile_0006.png",
			"assets/images/Tiles/tile_0005.png",
			"assets/images/Tiles/tile_0016.png",
		}
		imagePath := itemImages[rand.Intn(len(itemImages))]
		item := NewItem(imagePath, x, 0, "normal")
		s.Items = append(s.Items, item)
	}
	if rand.Float64() < 0.001 { // Less frequent health items
		x := rand.Float64() * 640
		imagePath := "assets/images/Tiles/tile_0024.png"
		item := NewItem(imagePath, x, 0, "health")
		s.Items = append(s.Items, item)
	}
}

func (s *Spitfire) CheckGameOver() {
	if s.Health <= 0 {
		s.GameOver = true
	}
}

func (s *Spitfire) Restart() {
	s.X = 320 - float64(s.Image.Bounds().Dx())/2
	s.Y = 480 - float64(s.Image.Bounds().Dy())*2 - 10 // Adjusted starting position
	s.Bullets = []*Bullet{}
	s.Items = []*Item{}
	s.Score = 0
	s.Health = 1.0
	s.GameOver = false
}
