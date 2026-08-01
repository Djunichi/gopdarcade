package gopdarcade

import (
	"reflect"
	"testing"

	"github.com/Djunichi/gopdsdk/playdate"
)

type logBitmap struct {
	log  *[]string
	name string
}

func (*logBitmap) Width() (int, error)       { return 1, nil }
func (*logBitmap) Height() (int, error)      { return 1, nil }
func (*logBitmap) Clear() error              { return nil }
func (*logBitmap) Fill(playdate.Color) error { return nil }
func (b *logBitmap) Close() error            { *b.log = append(*b.log, b.name); return nil }

type logSprite struct {
	log  *[]string
	name string
}

func (*logSprite) SetBitmap(playdate.Bitmap) error    { return nil }
func (*logSprite) SetPosition(float32, float32) error { return nil }
func (*logSprite) MoveBy(float32, float32) error      { return nil }
func (*logSprite) SetVisible(bool) error              { return nil }
func (*logSprite) SetZIndex(int) error                { return nil }
func (*logSprite) SetCollideRect(playdate.Rect) error { return nil }
func (*logSprite) ClearCollideRect() error            { return nil }
func (*logSprite) SetTag(uint8) error                 { return nil }
func (*logSprite) MoveWithCollisions(float32, float32) (playdate.MoveResult, error) {
	return playdate.MoveResult{}, nil
}
func (*logSprite) Add() error     { return nil }
func (*logSprite) Remove() error  { return nil }
func (s *logSprite) Close() error { *s.log = append(*s.log, s.name); return nil }

type logSound struct {
	log  *[]string
	name string
}

func (*logSound) Play() error                            { return nil }
func (*logSound) Stop() error                            { return nil }
func (*logSound) SetVolume(float32, float32) error       { return nil }
func (*logSound) Volume() (float32, float32, error)      { return 0, 0, nil }
func (*logSound) State() (playdate.PlaybackState, error) { return playdate.PlaybackStopped, nil }
func (*logSound) Pause() error                           { return nil }
func (*logSound) Resume() error                          { return nil }
func (s *logSound) Close() error                         { *s.log = append(*s.log, s.name); return nil }

type logFont struct{ log *[]string }

func (*logFont) TextWidth(string) (int, error) { return 0, nil }
func (*logFont) Height() (int, error)          { return 0, nil }
func (f *logFont) Close() error                { *f.log = append(*f.log, "font"); return nil }

func TestCleanupIsReverseOrderAndIdempotent(t *testing.T) {
	var got []string
	a := &arcade{}
	for i := range a.bitmaps {
		a.bitmaps[i] = &logBitmap{&got, "bitmap"}
	}
	for i := range a.sprites {
		a.sprites[i] = &logSprite{&got, "sprite"}
	}
	for i := range a.sounds {
		a.sounds[i] = &logSound{&got, "sound"}
	}
	a.font = &logFont{&got}
	if err := a.close(); err != nil {
		t.Fatal(err)
	}
	first := append([]string(nil), got...)
	if err := a.close(); err != nil {
		t.Fatal(err)
	}
	want := []string{"font", "sound", "sound", "sprite", "sprite", "sprite", "sprite", "sprite", "bitmap", "bitmap", "bitmap", "bitmap", "bitmap"}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("cleanup=%v", first)
	}
	if !reflect.DeepEqual(got, first) {
		t.Fatal("second cleanup repeated releases")
	}
}
