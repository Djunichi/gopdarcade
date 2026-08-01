# Crank Courier

An external, target-independent `gopdsdk` consumer and compact Playdate arcade.
Turn the crank (or use left/right) to steer. A pauses/resumes and retries after a
collision; B performs a deterministic reset.

Gameplay lives in `game` as a pure-Go deterministic state machine. The root
package translates immutable render/audio plans to public `playdate` APIs and
owns all native resources. Assets live under `resources/images`,
`resources/audio`, and `resources/fonts`.

## Verification

`go test ./...` is the portable gate. SDK integration uses:

```text
gopdsdk run --sdk <official-sdk> .
gopdsdk probe device --install --sdk <official-sdk>
```

Simulator, hard-float device compilation, physical install/run, ten-minute soak,
33 ms frame timing, bounded live heap, device logs, and rollback must be recorded
from real runs; CI alone does not prove them.
