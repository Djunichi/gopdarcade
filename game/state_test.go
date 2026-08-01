package game

import (
	"reflect"
	"testing"
)

func TestResetAndPlansAreDeterministic(t *testing.T) {
	a, b := New(), New()
	inputs := []Input{{CrankDelta: 12, DeltaSeconds: .033}, {Right: true, DeltaSeconds: .02}, {DeltaSeconds: .05}}
	for _, in := range inputs {
		ar, aa := a.Step(in)
		br, ba := b.Step(in)
		if !reflect.DeepEqual(ar, br) || !reflect.DeepEqual(aa, ba) {
			t.Fatal("same inputs diverged")
		}
	}
	a.Step(Input{Reset: true})
	if !reflect.DeepEqual(a, New()) {
		t.Fatal("reset did not restore exact initial state")
	}
}

func TestPauseLockResumeFreezesSimulation(t *testing.T) {
	s := New()
	before := s
	s.SetPaused(true)
	s.Step(Input{CrankDelta: 90, DeltaSeconds: .05})
	if s.PlayerX != before.PlayerX || s.Time != before.Time {
		t.Fatal("paused simulation advanced")
	}
	s.SetPaused(false)
	s.Step(Input{CrankDelta: 10, DeltaSeconds: .01})
	if s.PlayerX == before.PlayerX {
		t.Fatal("resume did not advance")
	}
}

func TestGameOverAndActionReset(t *testing.T) {
	s := New()
	s.HazardX[0], s.HazardY[0] = s.PlayerX, 216
	_, audio := s.Step(Input{})
	if s.Mode != GameOver || audio.Count != 1 || audio.Sounds[0] != SoundCrash {
		t.Fatal("collision did not end game")
	}
	_, audio = s.Step(Input{Action: true})
	if s.Mode != Playing || s.Score != 0 || audio.Sounds[0] != SoundStart {
		t.Fatal("action did not restart")
	}
}

func TestCollectibleCrossesPlayerLaneAndIncrementsScore(t *testing.T) {
	s := New()
	s.TargetX, s.TargetY = s.PlayerX, 215
	_, audio := s.Step(Input{DeltaSeconds: 1.0 / 30})
	if s.Score != 1 || audio.Count != 1 || audio.Sounds[0] != SoundCollect {
		t.Fatalf("score=%d audio=%+v", s.Score, audio)
	}
	if s.TargetY != 40 {
		t.Fatalf("target did not respawn at top: %v", s.TargetY)
	}
}

func TestMissedCollectibleRespawnsDeterministically(t *testing.T) {
	a, b := New(), New()
	a.TargetY, b.TargetY = 251, 251
	a.Step(Input{DeltaSeconds: .05})
	b.Step(Input{DeltaSeconds: .05})
	if a.TargetY != 40 || a.TargetX != b.TargetX {
		t.Fatalf("respawn diverged: a=%+v b=%+v", a, b)
	}
}

func TestSpeedIncreasesLinearlyEveryFifteenSeconds(t *testing.T) {
	s := New()
	s.Time = 14.99
	if got := s.renderPlan().SpeedPercent; got != 100 {
		t.Fatalf("speed before boundary=%d", got)
	}
	s.Time = 15
	if got := s.renderPlan().SpeedPercent; got != 106 {
		t.Fatalf("speed at first boundary=%d", got)
	}
	s.Time = 45
	if got := s.renderPlan().SpeedPercent; got != 118 {
		t.Fatalf("speed at third boundary=%d", got)
	}
}

func TestSpeedIncreaseIsCappedAndReset(t *testing.T) {
	s := New()
	s.Time = 3600
	if got := s.renderPlan().SpeedPercent; got != 160 {
		t.Fatalf("capped speed=%d", got)
	}
	s.Reset()
	if got := s.renderPlan().SpeedPercent; got != 100 {
		t.Fatalf("reset speed=%d", got)
	}
}

func TestStepDoesNotAllocate(t *testing.T) {
	s := New()
	in := Input{CrankDelta: 1, DeltaSeconds: 1.0 / 30}
	if got := testing.AllocsPerRun(1000, func() { s.Step(in) }); got != 0 {
		t.Fatalf("Step allocated %v", got)
	}
}
