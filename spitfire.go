package main

import (
	"image"
	"image/color"
	_ "image/jpeg" // Import the JPEG decoder
	_ "image/png"  // Import the PNG decoder
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
)

var (
	gameOverFont    font.Face
	levelUpFont     font.Face     // New font for level up sign
	backgroundImage *ebiten.Image // Background image
	boldFont        font.Face     // Bold font for score, high score, and level
	smallBoldFont   font.Face     // Smaller bold font for health text
)

func init() {
	rand.Seed(time.Now().UnixNano())

	// Load custom font
	tt, err := opentype.Parse(gobold.TTF)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	boldFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	smallBoldFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    18,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	gameOverFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    48,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	levelUpFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    72,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Load background image
	bgFile, err := os.Open("assets/images/background.png")
	if err != nil {
		log.Fatalf("Failed to open background image file: %v", err)
	}
	defer bgFile.Close()

	bgImg, _, err := image.Decode(bgFile)
	if err != nil {
		log.Fatalf("Failed to decode background image: %v", err)
	}

	backgroundImage = ebiten.NewImageFromImage(bgImg)
}

type Bullet struct {
	Image *ebiten.Image
	X, Y  float64
}

func NewBullet(x, y float64, imagePath string) *Bullet {
	file, err := os.Open(imagePath)
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
	b.Y -= 5 // Adjust bullet speed if necessary
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

	// Add a visual effect for the bullet upgrade and health items
	if itemType == "upgrade" {
		// Create a new image with a green glow effect
		glow := ebiten.NewImage(ebitenImage.Bounds().Dx()+10, ebitenImage.Bounds().Dy()+10)
		glow.Fill(color.RGBA{0, 255, 0, 128}) // Green glow
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(5, 5)
		glow.DrawImage(ebitenImage, op)
		ebitenImage = glow
	} else if itemType == "health" {
		// Create a new image with a white glow effect
		glow := ebiten.NewImage(ebitenImage.Bounds().Dx()+10, ebitenImage.Bounds().Dy()+10)
		glow.Fill(color.RGBA{255, 255, 255, 128}) // White glow
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(5, 5)
		glow.DrawImage(ebitenImage, op)
		ebitenImage = glow
	}

	return &Item{
		Image: ebitenImage,
		X:     x,
		Y:     y,
		Scale: scale,
		Type:  itemType,
	}
}

func (i *Item) Update(level int) {
	i.Y += 2
	if level == 2 {
		i.Scale = 2.0 // Double the size at level 2
	} else if level == 3 {
		i.Scale = 3.0 // Triple the size at level 3
	} else if level == 4 {
		i.Scale = 4.0 // Quadruple the size at level 4
		i.Y += 2      // Further increase item speed at level 4
	}
}

func (i *Item) Draw(screen *ebiten.Image) {
	// Draw shadow
	shadowOp := &ebiten.DrawImageOptions{}
	shadowOp.GeoM.Scale(i.Scale, i.Scale)
	shadowOp.GeoM.Translate(i.X+3, i.Y+3) // Offset the shadow slightly
	screen.DrawImage(i.Image, shadowOp)

	// Draw item
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(i.Scale, i.Scale)
	op.GeoM.Translate(i.X, i.Y)
	screen.DrawImage(i.Image, op)
}

type Spitfire struct {
	Image         *ebiten.Image
	X, Y          float64
	Bullets       []*Bullet
	Items         []*Item
	Score         int
	HighScore     int
	Health        float64
	GameOver      bool
	GameCompleted bool // New field to track game completion
	BulletUpgrade bool // New field to track bullet upgrade
	Level         int  // New field to track the level
	ShakeDuration int  // New field to track screen shake duration
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
		X:         320 - float64(ebitenImage.Bounds().Dx())*0.15/2,    // Adjusted starting position
		Y:         480 - float64(ebitenImage.Bounds().Dy())*0.15 - 10, // Adjusted starting position
		Bullets:   []*Bullet{},
		Items:     []*Item{},
		Score:     0,
		HighScore: 0,
		Health:    1.0, // Full health
		GameOver:  false,
		Level:     1, // Start at level 1
	}
}

func (s *Spitfire) Update() {
	if s.GameOver || s.GameCompleted {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			s.Restart()
		}
		return
	}

	if s.Score >= 3000 {
		s.GameCompleted = true
		return
	}

	if s.Score >= 1000 && s.Level == 3 {
		s.LevelUp(4)
	} else if s.Score >= 500 && s.Level == 2 {
		s.LevelUp(3)
	} else if s.Score >= 200 && s.Level == 1 {
		s.LevelUp(2)
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

	// Check if the space key is pressed to fire bullets
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		s.Fire()
	}

	// Prevent the plane from going off-screen
	planeWidth := float64(s.Image.Bounds().Dx()) * 0.15
	planeHeight := float64(s.Image.Bounds().Dy()) * 0.15
	if s.X < 0 {
		s.X = 0
	}
	if s.X > 640-planeWidth {
		s.X = 640 - planeWidth
	}
	if s.Y < 0 {
		s.Y = 0
	}
	if s.Y > 480-planeHeight {
		s.Y = 480 - planeHeight
	}

	for _, bullet := range s.Bullets {
		bullet.Update()
	}

	for _, item := range s.Items {
		item.Update(s.Level)
	}

	s.CheckCollisions()
	s.RemoveOffScreenBullets()
	s.RemoveOffScreenItems()
	s.SpawnItems()
	s.CheckGameOver()

	if s.ShakeDuration > 0 {
		s.ShakeDuration--
	}
}

func (s *Spitfire) LevelUp(newLevel int) {
	s.Level = newLevel
	s.ShakeDuration = 60 // Shake screen for 60 frames (1 second at 60 FPS)
	// Increase difficulty for new level
	for _, item := range s.Items {
		item.Y += float64(newLevel) // Increase item speed based on level
	}
	for _, bullet := range s.Bullets {
		bullet.Y -= float64(newLevel) // Increase bullet speed based on level
	}
	log.Printf("Level Up! Now at Level %d\n", newLevel)
}

func (s *Spitfire) Draw(screen *ebiten.Image) {
	// Draw background image
	op := &ebiten.DrawImageOptions{}
	bgWidth, bgHeight := backgroundImage.Size()
	screenWidth, screenHeight := screen.Size()
	scaleX := float64(screenWidth) / float64(bgWidth)
	scaleY := float64(screenHeight) / float64(bgHeight)
	op.GeoM.Scale(scaleX, scaleY)
	screen.DrawImage(backgroundImage, op)

	// Draw plane
	op = &ebiten.DrawImageOptions{}
	if s.ShakeDuration > 0 {
		op.GeoM.Translate(float64(rand.Intn(10)-5), float64(rand.Intn(10)-5)) // Shake effect
	}
	scale := 0.15 // Adjust the scale to make the plane smaller
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(s.X, s.Y)
	screen.DrawImage(s.Image, op)

	// Draw bullets
	for _, bullet := range s.Bullets {
		bullet.Draw(screen)
	}

	// Draw items
	for _, item := range s.Items {
		item.Draw(screen)
	}

	// Draw score, high score, and level
	scoreText := "Score: " + strconv.Itoa(s.Score)
	highScoreText := "High Score: " + strconv.Itoa(s.HighScore)
	levelText := "Level: " + strconv.Itoa(s.Level)
	scoreTextWidth := text.BoundString(boldFont, scoreText).Dx()
	highScoreTextWidth := text.BoundString(boldFont, highScoreText).Dx()
	levelTextWidth := text.BoundString(boldFont, levelText).Dx()

	// Draw text
	text.Draw(screen, scoreText, boldFont, (540-scoreTextWidth)/2, 20, color.White)
	text.Draw(screen, highScoreText, boldFont, (540-highScoreTextWidth)/2, 40, color.White)
	text.Draw(screen, levelText, boldFont, (540-levelTextWidth)/2, 60, color.White)

	// Draw health bar
	healthBarHeight := int(s.Health * 120) // Increase the height of the health bar
	ebitenutil.DrawRect(screen, 622, 480-float64(healthBarHeight), 10, float64(healthBarHeight), color.RGBA{255, 0, 0, 255})
	healthText := strconv.Itoa(int(s.Health*100)) + "%"
	text.Draw(screen, healthText, basicfont.Face7x13, 610, 495-healthBarHeight-10, color.White)

	// Draw vertical "HEALTH" text inside the health bar
	healthLabel := "HEALTH"
	for i, c := range healthLabel {
		text.Draw(screen, string(c), smallBoldFont, 621, 500-healthBarHeight+10+(i*18), color.Black)
	}

	// Draw game over and game completed messages with shadow
	if s.GameOver {
		gameOverText := "Game Over"
		gameOverTextWidth := text.BoundString(gameOverFont, gameOverText).Dx()

		// Draw shadow
		text.Draw(screen, gameOverText, gameOverFont, (640-gameOverTextWidth)/2+2, 242, color.Black)

		// Draw text
		text.Draw(screen, gameOverText, gameOverFont, (640-gameOverTextWidth)/2, 240, color.White)

		if s.Score > s.HighScore {
			s.HighScore = s.Score
			newHighScoreText := "New High Score: " + strconv.Itoa(s.HighScore)
			newHighScoreTextWidth := text.BoundString(gameOverFont, newHighScoreText).Dx()

			// Draw text
			text.Draw(screen, newHighScoreText, gameOverFont, (640-newHighScoreTextWidth)/2, 300, color.White)
		} else {
			highScoreText := "High Score: " + strconv.Itoa(s.HighScore)
			highScoreTextWidth := text.BoundString(gameOverFont, highScoreText).Dx()

			// Draw text
			text.Draw(screen, highScoreText, gameOverFont, (640-highScoreTextWidth)/2, 300, color.White)
		}

		// Draw "Press R to Restart" text
		text.Draw(screen, "Press R to Restart", boldFont, 270, 340, color.White)
	}

	if s.GameCompleted {
		gameCompletedText := "Game Completed!"
		gameCompletedTextWidth := text.BoundString(gameOverFont, gameCompletedText).Dx()

		// Draw text
		text.Draw(screen, gameCompletedText, gameOverFont, (640-gameCompletedTextWidth)/2, 240, color.White)

		if s.Score > s.HighScore {
			s.HighScore = s.Score
			newHighScoreText := "New High Score: " + strconv.Itoa(s.HighScore)
			newHighScoreTextWidth := text.BoundString(gameOverFont, newHighScoreText).Dx()

			// Draw text
			text.Draw(screen, newHighScoreText, gameOverFont, (640-newHighScoreTextWidth)/2, 300, color.White)
		} else {
			highScoreText := "High Score: " + strconv.Itoa(s.HighScore)
			highScoreTextWidth := text.BoundString(gameOverFont, highScoreText).Dx()

			// Draw text
			text.Draw(screen, highScoreText, gameOverFont, (640-highScoreTextWidth)/2, 300, color.White)
		}

		// Draw "Press R to Restart" text
		text.Draw(screen, "Press R to Restart", boldFont, 270, 340, color.White)
	}

	if s.ShakeDuration > 0 {
		levelUpText := "Level " + strconv.Itoa(s.Level)
		levelUpTextWidth := text.BoundString(levelUpFont, levelUpText).Dx()

		// Draw text
		text.Draw(screen, levelUpText, levelUpFont, (640-levelUpTextWidth)/2, 240, color.White)
	}
}

func (s *Spitfire) Fire() {
	normalBulletImagePath := "assets/images/Tiles/tile_0012.png"
	upgradedBulletImagePath := "assets/images/Tiles/tile_0015.png"
	scale := 0.13 // Size of the bullets
	planeWidth := float64(s.Image.Bounds().Dx()) * scale

	if s.BulletUpgrade {
		// Fire two lines of upgraded bullets
		bullet1 := NewBullet(s.X+planeWidth/2-20, s.Y, upgradedBulletImagePath)
		bullet2 := NewBullet(s.X+planeWidth/2+15, s.Y, upgradedBulletImagePath)
		s.Bullets = append(s.Bullets, bullet1, bullet2)
	} else {
		// Fire a single normal bullet
		bullet := NewBullet(s.X+planeWidth/2-20, s.Y, normalBulletImagePath)
		s.Bullets = append(s.Bullets, bullet)
	}
}

func (s *Spitfire) CheckCollisions() {
	var itemsToRemove []int

	planeWidth := float64(s.Image.Bounds().Dx()) * 0.15
	planeHeight := float64(s.Image.Bounds().Dy()) * 0.15

	for i := len(s.Items) - 1; i >= 0; i-- {
		item := s.Items[i]
		for j := len(s.Bullets) - 1; j >= 0; j-- {
			bullet := s.Bullets[j]
			if bullet.X < item.X+float64(item.Image.Bounds().Dx())*item.Scale &&
				bullet.X+float64(bullet.Image.Bounds().Dx()) > item.X &&
				bullet.Y < item.Y+float64(item.Image.Bounds().Dy())*item.Scale &&
				bullet.Y+float64(bullet.Image.Bounds().Dy()) > item.Y {
				// Remove all bullets and the item immediately
				s.Bullets = []*Bullet{}
				s.Items = append(s.Items[:i], s.Items[i+1:]...)
				if item.Type == "health" || item.Type == "upgrade" {
					s.Score -= 100
					if s.Score < 0 {
						s.Score = 0
					}
				} else {
					s.Score += 10
					if s.Score > s.HighScore {
						s.HighScore = s.Score
					}
				}
				break
			}
		}
		if s.X < item.X+float64(item.Image.Bounds().Dx())*item.Scale &&
			s.X+planeWidth > item.X &&
			s.Y < item.Y+float64(item.Image.Bounds().Dy())*item.Scale &&
			s.Y+planeHeight > item.Y {
			if item.Type == "health" {
				s.Health = 1.0 // Restore to full health
				log.Println("Health item collected")
			} else if item.Type == "upgrade" {
				s.BulletUpgrade = true // Activate bullet upgrade
				log.Println("Bullet upgrade collected")
			} else {
				s.Health -= 0.5
				if s.Health <= 0 {
					s.GameOver = true
				}
				log.Println("Normal item collected")
			}
			itemsToRemove = append(itemsToRemove, i)
		}
	}

	// Sort indices in descending order
	sort.Sort(sort.Reverse(sort.IntSlice(itemsToRemove)))

	// Remove items after iteration
	for _, i := range itemsToRemove {
		s.Items = append(s.Items[:i], s.Items[i+1:]...)
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
	if rand.Float64() < 0.001 { // Less frequent upgrade items
		x := rand.Float64() * 640
		imagePath := "assets/images/Tiles/tile_0007.png" // Ensure the correct path
		item := NewItem(imagePath, x, 0, "upgrade")
		s.Items = append(s.Items, item)
	}
}

func (s *Spitfire) CheckGameOver() {
	if s.Health <= 0 {
		s.GameOver = true
	}
}

func (s *Spitfire) Restart() {
	scale := 0.15 // Adjust the scale as needed
	s.X = 320 - float64(s.Image.Bounds().Dx())*scale/2
	s.Y = 480 - float64(s.Image.Bounds().Dy())*scale - 10 // Adjusted starting position
	s.Bullets = []*Bullet{}
	s.Items = []*Item{}
	s.Score = 0
	s.Health = 1.0
	s.GameOver = false
	s.GameCompleted = false // Reset game completion
	s.BulletUpgrade = false // Reset bullet upgrade
	s.Level = 1             // Reset level
	s.ShakeDuration = 0     // Reset shake duration
}
