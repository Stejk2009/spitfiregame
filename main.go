package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	Spitfire *Spitfire
}

func (g *Game) Update() error {
	if g.Spitfire.GameOver && ebiten.IsKeyPressed(ebiten.KeyR) {
		g.Spitfire.Restart()
	}
	g.Spitfire.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Spitfire.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 640, 480
}

func main() {
	spitfire := NewSpitfire()

	game := &Game{
		Spitfire: spitfire,
	}

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Spitfire Game")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
