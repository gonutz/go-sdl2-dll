//+build windows,386

package sdl

// GetPerformanceCounter returns the current value of the high resolution counter.
// (https://wiki.libsdl.org/SDL_GetPerformanceCounter)
func GetPerformanceCounter() uint64 {
	r1, r2, _ := getPerformanceCounter.Call()
	return uint64(r2)<<32 + uint64(r1)
}

// GetPerformanceFrequency returns the count per second of the high resolution counter.
// (https://wiki.libsdl.org/SDL_GetPerformanceFrequency)
func GetPerformanceFrequency() uint64 {
	r1, r2, _ := getPerformanceFrequency.Call()
	return uint64(r2)<<32 + uint64(r1)
}
