package sdl_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gonutz/check"
	"github.com/gonutz/go-sdl2/sdl"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 {
		if err := sdl.LoadDLL(os.Args[1]); err != nil {
			panic(err)
		}
	}
	os.Exit(m.Run())
}

func test(f func()) {
	sdl.Main(func() {
		sdl.Do(f)
	})
}

func TestAddAndDeleteEventWatch(t *testing.T) {
	test(func() {
		sdl.Init(0)
		defer sdl.Quit()

		// add event watch and trigger a known event
		var f eventFilter
		handle := sdl.AddEventWatch(&f, "test")
		e := sdl.UserEvent{Type: sdl.USEREVENT, Code: 42}
		sdl.PushEvent(&e)
		check.Eq(t, f.event, &e)
		check.Eq(t, f.userdata, "test")

		// delete event watch, clear the variables, re-trigger the event, this
		// time the variables must not be updated
		sdl.DelEventWatch(handle)
		f.event = nil
		f.userdata = nil
		sdl.PushEvent(&e)
		check.Eq(t, f.event, nil)
		check.Eq(t, f.userdata, nil)
	})
}

type eventFilter struct {
	event    sdl.Event
	userdata interface{}
}

func (f *eventFilter) FilterEvent(e sdl.Event, userdata interface{}) bool {
	f.event = e
	f.userdata = userdata
	return true
}

func TestAddAndDeleteEventWatchFunc(t *testing.T) {
	test(func() {
		sdl.Init(0)
		defer sdl.Quit()

		// add event watch and trigger a known event
		var (
			event    sdl.Event
			userdata interface{}
		)
		handle := sdl.AddEventWatchFunc(func(e sdl.Event, u interface{}) bool {
			event = e
			userdata = u
			return true
		}, "test")
		e := sdl.UserEvent{Type: sdl.USEREVENT, Code: 42}
		sdl.PushEvent(&e)
		check.Eq(t, event, &e)
		check.Eq(t, userdata, "test")

		// delete event watch, clear the variables, re-trigger the event, this
		// time the variables must not be updated
		sdl.DelEventWatch(handle)
		event = nil
		userdata = nil
		sdl.PushEvent(&e)
		check.Eq(t, event, nil)
		check.Eq(t, userdata, nil)
	})
}

func TestAddAndDeleteHintCallback(t *testing.T) {
	test(func() {
		sdl.Init(0)
		defer sdl.Quit()
		var x []interface{} // holds all callback data
		sdl.AddHintCallback(
			sdl.HINT_RENDER_SCALE_QUALITY,
			func(data interface{}, name, oldValue, newValue string) {
				x = append(x, data, name, oldValue, newValue)
			},
			"test",
		)
		// NOTE that the first time we add a hint callback, it is immediately
		// called with the old and new value being the same
		sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1") // change "" -> "1"
		sdl.SetHint(sdl.HINT_RENDER_DRIVER, "1")        // should not trigger
		check.Eq(t, x, []interface{}{
			"test", sdl.HINT_RENDER_SCALE_QUALITY, "", "", // initial call
			"test", sdl.HINT_RENDER_SCALE_QUALITY, "", "1",
		})

		// delete the hint callback, x should stay the same
		sdl.DelHintCallback(sdl.HINT_RENDER_SCALE_QUALITY)
		sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "")
		check.Eq(t, x, []interface{}{
			"test", sdl.HINT_RENDER_SCALE_QUALITY, "", "",
			"test", sdl.HINT_RENDER_SCALE_QUALITY, "", "1",
		})
	})
}

func TestDelay(t *testing.T) {
	start := time.Now()
	sdl.Delay(100)
	end := time.Now()
	check.EqEps(t, end.Sub(start).Seconds(), 0.100, 0.01)
}

func TestBtoi(t *testing.T) {
	check.Eq(t, sdl.Btoi(false), 0)
	check.Eq(t, sdl.Btoi(true), 1)
}

func TestRWFromMem(t *testing.T) {
	rw, err := sdl.RWFromMem(make([]byte, 10))
	check.Eq(t, err, nil)
	n, err := rw.Size()
	check.Eq(t, err, nil)
	check.Eq(t, n, 10)
}

func TestRWFromFileAndRWClose(t *testing.T) {
	// create a test file
	path := filepath.Join(os.Getenv("APPDATA"), "sdl_test_rw")
	err := ioutil.WriteFile(path, make([]byte, 10), 0666)

	// open the test file in read/write mode
	rw := sdl.RWFromFile(path, "rwb")
	n, err := rw.Size()
	check.Eq(t, err, nil)
	check.Eq(t, n, 10)

	// at this point the RWops have not yet been closed, this means that Windows
	// should prevent us from being able to delete the file
	if err := os.Remove(path); err == nil {
		t.Error("error expected when deleting unclosed RWops")
	}
	// close the RWops and we should be able to delete the file
	check.Eq(t, rw.Close(), nil)
	check.Eq(t, os.Remove(path), nil)
}

func TestRWReadWrite(t *testing.T) {
	buf := make([]byte, 1+2+2+4+4+8+8+3+6)
	rw, err := sdl.RWFromMem(buf)
	check.Eq(t, err, nil)

	check.Eq(t, rw.WriteU8(0x01), 1)
	check.Eq(t, rw.WriteBE16(0x02+0x03<<8), 1)
	check.Eq(t, rw.WriteLE16(0x04+0x05<<8), 1)
	check.Eq(t, rw.WriteBE32(0x06+0x07<<8+0x08<<16+0x09<<24), 1)
	check.Eq(t, rw.WriteLE32(0x0A+0x0B<<8+0x0C<<16+0x0D<<24), 1)
	check.Eq(t, rw.WriteBE64(0x0E+0x0F<<8+0x10<<16+0x11<<24+0x12<<32+0x13<<40+0x14<<48+0x15<<56), 1)
	check.Eq(t, rw.WriteLE64(0x16+0x17<<8+0x18<<16+0x19<<24+0x1A<<32+0x1B<<40+0x1C<<48+0x1D<<56), 1)

	n, err := rw.Write([]byte("abc"))
	check.Eq(t, err, nil)
	check.Eq(t, n, 3)

	n, err = rw.Write2([]byte("123456789"), 3, 2)
	check.Eq(t, err, nil)
	check.Eq(t, n, 2)

	check.Eq(t, rw.WriteU8(0x00), 0) // the buffer is full, 0 written

	check.Eq(t, buf, []byte{
		0x01,
		0x03, 0x02,
		0x04, 0x05,
		0x09, 0x08, 0x07, 0x06,
		0x0A, 0x0B, 0x0C, 0x0D,
		0x15, 0x14, 0x13, 0x12, 0x11, 0x10, 0x0F, 0x0E,
		0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D,
		'a', 'b', 'c',
		'1', '2', '3', '4', '5', '6',
	})

	size, err := rw.Size()
	check.Eq(t, err, nil)
	check.Eq(t, size, len(buf))

	pos, err := rw.Tell()
	check.Eq(t, err, nil)
	check.Eq(t, pos, len(buf))

	pos, err = rw.Seek(0, sdl.RW_SEEK_SET)
	check.Eq(t, err, nil)
	check.Eq(t, pos, 0)

	check.Eq(t, rw.ReadU8(), 0x01)
	check.Eq(t, rw.ReadBE16(), 0x02+0x03<<8)
	check.Eq(t, rw.ReadLE16(), 0x04+0x05<<8)
	check.Eq(t, rw.ReadBE32(), 0x06+0x07<<8+0x08<<16+0x09<<24)
	check.Eq(t, rw.ReadLE32(), 0x0A+0x0B<<8+0x0C<<16+0x0D<<24)
	check.Eq(t, rw.ReadBE64(), uint64(0x0E+0x0F<<8+0x10<<16+0x11<<24+0x12<<32+0x13<<40+0x14<<48+0x15<<56))
	check.Eq(t, rw.ReadLE64(), uint64(0x16+0x17<<8+0x18<<16+0x19<<24+0x1A<<32+0x1B<<40+0x1C<<48+0x1D<<56))

	var readBuf [3]byte
	n, err = rw.Read(readBuf[:])
	check.Eq(t, err, nil)
	check.Eq(t, n, 3)
	check.Eq(t, readBuf[:], []byte("abc"))

	var read2Buf [6]byte
	n, err = rw.Read2(read2Buf[:], 3, 2)
	check.Eq(t, err, nil)
	check.Eq(t, n, 2)
	check.Eq(t, read2Buf[:], []byte("123456"))
}

func TestLog(t *testing.T) {
	var x []interface{}
	f := func(data interface{}, category int, pri sdl.LogPriority, message string) {
		x = append(x, data, category, pri, message)
	}

	sdl.LogSetOutputFunction(f, "test")
	check.Eq(t, len(x), 0)
	haveF, haveData := sdl.LogGetOutputFunction()
	check.Eq(t, haveF, sdl.LogOutputFunction(f))
	check.Eq(t, haveData, "test")
	sdl.Log("%d", 123)
	check.Eq(t, x, []interface{}{
		"test",
		sdl.LOG_CATEGORY_APPLICATION,
		sdl.LogPriority(sdl.LOG_PRIORITY_INFO),
		"123",
	})

	x = nil
	sdl.LogCritical(sdl.LOG_CATEGORY_RENDER, "%.1f", 1.1)
	check.Eq(t, x, []interface{}{
		"test",
		sdl.LOG_CATEGORY_RENDER,
		sdl.LogPriority(sdl.LOG_PRIORITY_CRITICAL),
		"1.1",
	})

	x = nil
	sdl.LogDebug(sdl.LOG_CATEGORY_APPLICATION, "debug is inactive")
	check.Eq(t, len(x), 0)

	x = nil
	sdl.LogError(sdl.LOG_CATEGORY_APPLICATION, "%s%s", "a", "b")
	check.Eq(t, x, []interface{}{
		"test",
		sdl.LOG_CATEGORY_APPLICATION,
		sdl.LogPriority(sdl.LOG_PRIORITY_ERROR),
		"ab",
	})

	x = nil
	sdl.LogInfo(sdl.LOG_CATEGORY_APPLICATION, "%v info", true)
	check.Eq(t, x, []interface{}{
		"test",
		sdl.LOG_CATEGORY_APPLICATION,
		sdl.LogPriority(sdl.LOG_PRIORITY_INFO),
		"true info",
	})

	x = nil
	sdl.LogMessage(sdl.LOG_CATEGORY_APPLICATION, sdl.LOG_PRIORITY_WARN, "warning")
	check.Eq(t, x, []interface{}{
		"test",
		sdl.LOG_CATEGORY_APPLICATION,
		sdl.LogPriority(sdl.LOG_PRIORITY_WARN),
		"warning",
	})

	x = nil
	sdl.LogVerbose(sdl.LOG_CATEGORY_APPLICATION, "verbose is inactive")
	check.Eq(t, len(x), 0)

	x = nil
	sdl.LogWarn(sdl.LOG_CATEGORY_APPLICATION, "warn me")
	check.Eq(t, x, []interface{}{
		"test",
		sdl.LOG_CATEGORY_APPLICATION,
		sdl.LogPriority(sdl.LOG_PRIORITY_WARN),
		"warn me",
	})

	sdl.LogSetPriority(sdl.LOG_CATEGORY_INPUT, sdl.LOG_PRIORITY_VERBOSE)
	check.Eq(t, sdl.LogGetPriority(sdl.LOG_CATEGORY_INPUT), sdl.LOG_PRIORITY_VERBOSE)
	x = nil
	sdl.LogVerbose(sdl.LOG_CATEGORY_INPUT, "verbose now active")
	check.Eq(t, x, []interface{}{
		"test",
		sdl.LOG_CATEGORY_INPUT,
		sdl.LogPriority(sdl.LOG_PRIORITY_VERBOSE),
		"verbose now active",
	})

	sdl.LogSetPriority(sdl.LOG_CATEGORY_INPUT, sdl.LOG_PRIORITY_DEBUG)
	check.Eq(t, sdl.LogGetPriority(sdl.LOG_CATEGORY_INPUT), sdl.LOG_PRIORITY_DEBUG)
	x = nil
	sdl.LogDebug(sdl.LOG_CATEGORY_INPUT, "debug now active")
	check.Eq(t, x, []interface{}{
		"test",
		sdl.LOG_CATEGORY_INPUT,
		sdl.LogPriority(sdl.LOG_PRIORITY_DEBUG),
		"debug now active",
	})

	sdl.LogResetPriorities()
	x = nil
	sdl.LogDebug(sdl.LOG_CATEGORY_APPLICATION, "debug inactive again")
	check.Eq(t, len(x), 0)

	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_DEBUG)
	x = nil
	sdl.LogDebug(sdl.LOG_CATEGORY_INPUT, "debug active again")
	check.Eq(t, x, []interface{}{
		"test",
		sdl.LOG_CATEGORY_INPUT,
		sdl.LogPriority(sdl.LOG_PRIORITY_DEBUG),
		"debug active again",
	})
}
