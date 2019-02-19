//+build windows,386

package sdl

import "unsafe"

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

// WriteBE64 writes 64 bits in native format to the RWops as big-endian data.
// (https://wiki.libsdl.org/SDL_WriteBE64)
func (rwops *RWops) WriteBE64(value uint64) uint {
	if rwops == nil {
		return 0
	}
	a := uint32(value)
	b := uint32(value >> 32)
	ret, _, _ := writeBE64.Call(
		uintptr(unsafe.Pointer(rwops)),
		uintptr(a),
		uintptr(b),
	)
	return uint(ret)
}

// WriteLE64 writes 64 bits in native format to the RWops as little-endian data.
// (https://wiki.libsdl.org/SDL_WriteLE64)
func (rwops *RWops) WriteLE64(value uint64) uint {
	if rwops == nil {
		return 0
	}
	a := uint32(value)
	b := uint32(value >> 32)
	ret, _, _ := writeLE64.Call(
		uintptr(unsafe.Pointer(rwops)),
		uintptr(a),
		uintptr(b),
	)
	return uint(ret)
}
