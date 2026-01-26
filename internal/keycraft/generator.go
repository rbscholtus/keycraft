package keycraft

// GenerateInput captures CLI inputs for generation.
// This is a stub that will be expanded when the full generator is implemented.
type GenerateInput struct {
	ConfigPath      string // resolved .gen file path
	MaxLayouts      int    // from --max-layouts flag (default 1500, 0=all)
	Seed            uint64 // from --seed flag (0=timestamp)
	Optimize        bool   // from --optimize flag
	KeepUnoptimized bool   // from --keep-unoptimized flag
}
