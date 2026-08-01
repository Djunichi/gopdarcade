// Package gopdarcade is the Playdate adapter for the pure-Go Crank Courier game.
package gopdarcade

import (
	"github.com/Djunichi/gopdarcade/game"
	"github.com/Djunichi/gopdsdk/playdate"
)

const (
	playerImage  = "images/player"
	targetImage  = "images/target"
	hazardImage  = "images/hazard"
	grassImage   = "images/grass"
	collectAudio = "audio/collect"
	crashAudio   = "audio/crash"
	fontAsset    = "fonts/arcade"
)

type arcade struct {
	state     game.State
	bitmaps   [5]playdate.Bitmap
	sprites   [5]playdate.Sprite
	sounds    [2]playdate.SoundEffect
	font      playdate.Font
	speedX    int
	pauseX    int
	gameOverX int
	retryX    int
	closed    bool
}

type gameError string

func (e gameError) Error() string { return string(e) }

func firstError(first, second error) error {
	if first == nil {
		return second
	}
	return first
}

// New returns the target-independent application entry point.
func New() playdate.Game { return &arcade{state: game.New()} }

func (a *arcade) Init(ctx playdate.Context) (err error) {
	paths := [...]string{playerImage, targetImage, hazardImage, grassImage}
	for i, path := range paths {
		a.bitmaps[i], err = ctx.LoadBitmap(path)
		if err != nil {
			return a.rollback(err)
		}
	}
	a.bitmaps[4], err = ctx.NewBitmap(game.ScreenWidth, 28)
	if err != nil {
		return a.rollback(err)
	}
	if err = a.bitmaps[4].Fill(playdate.ColorWhite); err != nil {
		return a.rollback(err)
	}
	for i := range a.sprites {
		a.sprites[i], err = ctx.NewSprite()
		if err != nil {
			return a.rollback(err)
		}
		kind := 0
		if i == 1 {
			kind = 1
		}
		if i > 1 {
			kind = 2
		}
		if err = a.sprites[i].SetBitmap(a.bitmaps[kind]); err != nil {
			return a.rollback(err)
		}
		if err = a.sprites[i].SetZIndex(i); err != nil {
			return a.rollback(err)
		}
		if err = a.sprites[i].SetCollideRect(playdate.Rect{X: -10, Y: -10, Width: 20, Height: 20}); err != nil {
			return a.rollback(err)
		}
		if err = a.sprites[i].Add(); err != nil {
			return a.rollback(err)
		}
	}
	a.sounds[0], err = ctx.LoadSoundEffect(collectAudio)
	if err != nil {
		return a.rollback(err)
	}
	a.sounds[1], err = ctx.LoadSoundEffect(crashAudio)
	if err != nil {
		return a.rollback(err)
	}
	fonts, ok := ctx.(playdate.FontGraphics)
	if !ok {
		return a.rollback(gameError("custom fonts unavailable"))
	}
	a.font, err = fonts.LoadFont(fontAsset)
	if err != nil {
		return a.rollback(err)
	}
	speedWidth, err := a.font.TextWidth("SPEED 160%")
	if err != nil {
		return a.rollback(err)
	}
	a.speedX = game.ScreenWidth - 12 - speedWidth
	if a.pauseX, err = centeredTextX(a.font, "PAUSED"); err != nil {
		return a.rollback(err)
	}
	if a.gameOverX, err = centeredTextX(a.font, "GAME OVER"); err != nil {
		return a.rollback(err)
	}
	if a.retryX, err = centeredTextX(a.font, "A: RETRY"); err != nil {
		return a.rollback(err)
	}
	return nil
}

func centeredTextX(font playdate.Font, text string) (int, error) {
	width, err := font.TextWidth(text)
	return (game.ScreenWidth - width) / 2, err
}

func (a *arcade) rollback(cause error) error { _ = a.close(); return cause }

func (a *arcade) Update(ctx playdate.Context) (bool, error) {
	in := ctx.Input()
	plan, audio := a.state.Step(game.Input{CrankDelta: in.CrankDelta, Left: in.Held.Has(playdate.ButtonLeft), Right: in.Held.Has(playdate.ButtonRight), Action: in.Pressed.Has(playdate.ButtonA), Reset: in.Pressed.Has(playdate.ButtonB), DeltaSeconds: in.DeltaSeconds})
	ctx.Clear()
	for i := 0; i < plan.Count; i++ {
		if err := a.sprites[i].SetPosition(plan.Sprites[i].X, plan.Sprites[i].Y); err != nil {
			return false, err
		}
	}
	ctx.UpdateAndDrawSprites()
	fonts := ctx.(playdate.FontGraphics)
	if err := ctx.DrawBitmap(a.bitmaps[3], 0, 220); err != nil {
		return false, err
	}
	if err := ctx.DrawBitmap(a.bitmaps[4], 0, 0); err != nil {
		return false, err
	}
	if err := fonts.DrawTextFont(a.font, scoreText(plan.Score), 12, 8); err != nil {
		return false, err
	}
	if err := fonts.DrawTextFont(a.font, "SPEED "+decimal(plan.SpeedPercent, 0)+"%", a.speedX, 8); err != nil {
		return false, err
	}
	if plan.Mode == game.Paused {
		if err := ctx.DrawScaledBitmap(a.bitmaps[4], 0, 92, 1, 1.8); err != nil {
			return false, err
		}
		if err := fonts.DrawTextFont(a.font, "PAUSED", a.pauseX, 108); err != nil {
			return false, err
		}
	}
	if plan.Mode == game.GameOver {
		if err := ctx.DrawScaledBitmap(a.bitmaps[4], 0, 92, 1, 1.8); err != nil {
			return false, err
		}
		if err := fonts.DrawTextFont(a.font, "GAME OVER", a.gameOverX, 101); err != nil {
			return false, err
		}
		if err := fonts.DrawTextFont(a.font, "A: RETRY", a.retryX, 119); err != nil {
			return false, err
		}
	}
	for i := 0; i < audio.Count; i++ {
		var sound playdate.SoundEffect
		if audio.Sounds[i] == game.SoundCrash {
			sound = a.sounds[1]
		} else {
			sound = a.sounds[0]
		}
		if err := sound.Play(); err != nil {
			return false, err
		}
	}
	return true, nil
}

func scoreText(score int) string {
	return "SCORE " + decimal(score, 4)
}

func decimal(value, minimum int) string {
	if value < 0 {
		value = 0
	}
	var digits [12]byte
	index := len(digits)
	for value > 0 || len(digits)-index < minimum {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	if index == len(digits) {
		return "0"
	}
	return string(digits[index:])
}

func (a *arcade) HandleLifecycle(_ playdate.Context, event playdate.LifecycleEvent) error {
	switch event {
	case playdate.LifecyclePause, playdate.LifecycleLock, playdate.LifecycleLowPower:
		a.state.SetPaused(true)
		return a.pauseAudio()
	case playdate.LifecycleResume, playdate.LifecycleUnlock:
		a.state.SetPaused(false)
		return a.resumeAudio()
	case playdate.LifecycleTerminate:
		return a.close()
	}
	return nil
}

func (a *arcade) pauseAudio() (err error) {
	for _, s := range a.sounds {
		if s != nil {
			err = firstError(err, s.Pause())
		}
	}
	return err
}
func (a *arcade) resumeAudio() (err error) {
	for _, s := range a.sounds {
		if s != nil {
			state, e := s.State()
			err = firstError(err, e)
			if e == nil && state == playdate.PlaybackPaused {
				err = firstError(err, s.Resume())
			}
		}
	}
	return err
}
func (a *arcade) close() (err error) {
	if a.closed {
		return nil
	}
	a.closed = true
	if a.font != nil {
		err = firstError(err, a.font.Close())
	}
	for i := len(a.sounds) - 1; i >= 0; i-- {
		if a.sounds[i] != nil {
			err = firstError(err, a.sounds[i].Stop())
			err = firstError(err, a.sounds[i].Close())
		}
	}
	for i := len(a.sprites) - 1; i >= 0; i-- {
		if a.sprites[i] != nil {
			err = firstError(err, a.sprites[i].Close())
		}
	}
	for i := len(a.bitmaps) - 1; i >= 0; i-- {
		if a.bitmaps[i] != nil {
			err = firstError(err, a.bitmaps[i].Close())
		}
	}
	return err
}
