// Package game implements the pure-Go Crank Courier gameplay state machine.
package game

import "math"

const (
	ScreenWidth      = 400
	ScreenHeight     = 240
	speedStepSeconds = 15
	speedStep        = float32(.06)
	maxSpeed         = float32(1.6)
)

type Mode uint8

const (
	Playing Mode = iota
	Paused
	GameOver
)

type Input struct {
	CrankDelta                 float32
	Left, Right, Action, Reset bool
	DeltaSeconds               float32
}

type Sound uint8

const (
	SoundCollect Sound = iota
	SoundCrash
	SoundStart
)

type SpritePlan struct {
	Kind  uint8
	X, Y  float32
	Frame int
}
type RenderPlan struct {
	Mode         Mode
	Score        int
	SpeedPercent int
	Sprites      [5]SpritePlan
	Count        int
}
type AudioPlan struct {
	Sounds [2]Sound
	Count  int
}

type State struct {
	Mode             Mode
	PlayerX          float32
	TargetX, TargetY float32
	HazardX          [3]float32
	HazardY          [3]float32
	Score            int
	Time             float32
	seed             uint32
}

func New() State { var s State; s.Reset(); return s }

func (s *State) Reset() {
	*s = State{PlayerX: 200, TargetX: 200, TargetY: 172, seed: 0xC0FFEE}
	s.HazardX = [3]float32{40, 180, 330}
	s.HazardY = [3]float32{90, 145, 200}
}

func (s *State) SetPaused(paused bool) {
	if s.Mode == GameOver {
		return
	}
	if paused {
		s.Mode = Paused
	} else {
		s.Mode = Playing
	}
}

func (s *State) Step(in Input) (RenderPlan, AudioPlan) {
	var audio AudioPlan
	restarted := in.Reset || (s.Mode == GameOver && in.Action)
	if restarted {
		s.Reset()
		audio.Sounds[0] = SoundStart
		audio.Count = 1
	}
	if in.Action && s.Mode != GameOver && !restarted {
		s.SetPaused(s.Mode != Paused)
	}
	if s.Mode == Playing {
		dt := in.DeltaSeconds
		if dt < 0 {
			dt = 0
		}
		if dt > .05 {
			dt = .05
		}
		move := in.CrankDelta * 1.8
		if in.Left {
			move -= 110 * dt
		}
		if in.Right {
			move += 110 * dt
		}
		s.PlayerX = wrap(s.PlayerX+move, 12, 388)
		s.Time += dt
		speed := s.speedMultiplier()
		s.TargetY += 34 * speed * dt
		if s.TargetY > 252 {
			s.respawnTarget()
		}
		for i := range s.HazardX {
			s.HazardY[i] += (42 + float32(i)*9) * speed * dt
			if s.HazardY[i] > 252 {
				s.HazardY[i] = -12
				s.HazardX[i] = float32(20 + s.next()%360)
			}
		}
		if near(s.PlayerX, 216, s.TargetX, s.TargetY, 22) {
			s.Score++
			s.respawnTarget()
			audio.Sounds[audio.Count] = SoundCollect
			audio.Count++
		}
		for i := range s.HazardX {
			if near(s.PlayerX, 216, s.HazardX[i], s.HazardY[i], 20) {
				s.Mode = GameOver
				audio.Sounds[audio.Count] = SoundCrash
				audio.Count++
				break
			}
		}
	}
	return s.renderPlan(), audio
}

func (s *State) renderPlan() RenderPlan {
	p := RenderPlan{Mode: s.Mode, Score: s.Score, SpeedPercent: int(s.speedMultiplier()*100 + .5)}
	p.Sprites[0] = SpritePlan{Kind: 0, X: s.PlayerX, Y: 216, Frame: int(s.Time*10) % 2}
	p.Count = 1
	p.Sprites[p.Count] = SpritePlan{Kind: 1, X: s.TargetX, Y: s.TargetY, Frame: int(s.Time*6) % 2}
	p.Count++
	for i := range s.HazardX {
		p.Sprites[p.Count] = SpritePlan{Kind: 2, X: s.HazardX[i], Y: s.HazardY[i], Frame: i % 2}
		p.Count++
	}
	return p
}

func (s *State) next() uint32 { s.seed = s.seed*1664525 + 1013904223; return s.seed }
func (s *State) speedMultiplier() float32 {
	multiplier := 1 + float32(int(s.Time/speedStepSeconds))*speedStep
	if multiplier > maxSpeed {
		return maxSpeed
	}
	return multiplier
}
func (s *State) respawnTarget() {
	s.TargetX = float32(24 + s.next()%352)
	s.TargetY = 40
}
func wrap(v, lo, hi float32) float32 {
	width := hi - lo
	for v < lo {
		v += width
	}
	for v > hi {
		v -= width
	}
	return v
}
func near(ax, ay, bx, by, r float32) bool {
	dx, dy := ax-bx, ay-by
	return float32(math.Abs(float64(dx))) < r && float32(math.Abs(float64(dy))) < r
}
