//+build windows

package sdl

import "errors"

// Audio format masks.
// (https://wiki.libsdl.org/SDL_AudioFormat)
const (
	AUDIO_MASK_BITSIZE  = 0xFF
	AUDIO_MASK_DATATYPE = 1 << 8
	AUDIO_MASK_ENDIAN   = 1 << 12
	AUDIO_MASK_SIGNED   = 1 << 15
)

// Audio format values.
// (https://wiki.libsdl.org/SDL_AudioFormat)
const (
	AUDIO_S8 = 0x8008 // unsigned 8-bit samples
	AUDIO_U8 = 0x0008 // signed 8-bit samples

	AUDIO_S16LSB = 0x8010       // signed 16-bit samples in little-endian byte order
	AUDIO_S16MSB = 0x9010       // signed 16-bit samples in big-endian byte order
	AUDIO_S16SYS = AUDIO_S16LSB // signed 16-bit samples in native byte order
	AUDIO_S16    = AUDIO_S16LSB // AUDIO_S16LSB
	AUDIO_U16LSB = 0x0010       // unsigned 16-bit samples in little-endian byte order
	AUDIO_U16MSB = 0x1010       // unsigned 16-bit samples in big-endian byte order
	AUDIO_U16SYS = AUDIO_U16LSB // unsigned 16-bit samples in native byte order
	AUDIO_U16    = AUDIO_U16LSB // AUDIO_U16LSB

	AUDIO_S32LSB = 0x8020       // 32-bit integer samples in little-endian byte order
	AUDIO_S32MSB = 0x9020       // 32-bit integer samples in big-endian byte order
	AUDIO_S32SYS = AUDIO_S32LSB // 32-bit integer samples in native byte order
	AUDIO_S32    = AUDIO_S32LSB // AUDIO_S32LSB

	AUDIO_F32LSB = 0x8120       // 32-bit floating point samples in little-endian byte order
	AUDIO_F32MSB = 0x9120       // 32-bit floating point samples in big-endian byte order
	AUDIO_F32SYS = AUDIO_F32LSB // 32-bit floating point samples in native byte order
	AUDIO_F32    = AUDIO_F32LSB // AUDIO_F32LSB
)

// AllowedChanges flags specify how SDL should behave when a device cannot offer a specific feature. If the application requests a feature that the hardware doesn't offer, SDL will always try to get the closest equivalent. Used in OpenAudioDevice().
// (https://wiki.libsdl.org/SDL_OpenAudioDevice)
const (
	AUDIO_ALLOW_FREQUENCY_CHANGE = 0x00000001
	AUDIO_ALLOW_FORMAT_CHANGE    = 0x00000002
	AUDIO_ALLOW_CHANNELS_CHANGE  = 0x00000004
	AUDIO_ALLOW_SAMPLES_CHANGE   = 0x00000008
	AUDIO_ALLOW_ANY_CHANGE       = (AUDIO_ALLOW_FREQUENCY_CHANGE | AUDIO_ALLOW_FORMAT_CHANGE | AUDIO_ALLOW_CHANNELS_CHANGE | AUDIO_ALLOW_SAMPLES_CHANGE)
)

// An enumeration of audio device states used in GetAudioDeviceStatus() and GetAudioStatus().
// (https://wiki.libsdl.org/SDL_AudioStatus)
const (
	AUDIO_STOPPED AudioStatus = iota // audio device is stopped
	AUDIO_PLAYING                    // audio device is playing
	AUDIO_PAUSED                     // audio device is paused
)

const (
	BLENDMODE_NONE    = 0x00000000 // no blending
	BLENDMODE_BLEND   = 0x00000001 // alpha blending
	BLENDMODE_ADD     = 0x00000002 // additive blending
	BLENDMODE_MOD     = 0x00000004 // color modulate
	BLENDMODE_INVALID = 0x7FFFFFFF
)

const (
	BLENDOPERATION_ADD          = 0x1 // dst + src: supported by all renderers
	BLENDOPERATION_SUBTRACT     = 0x2 // dst - src : supported by D3D9, D3D11, OpenGL, OpenGLES
	BLENDOPERATION_REV_SUBTRACT = 0x3 // src - dst : supported by D3D9, D3D11, OpenGL, OpenGLES
	BLENDOPERATION_MINIMUM      = 0x4 // min(dst, src) : supported by D3D11
	BLENDOPERATION_MAXIMUM      = 0x5 // max(dst, src) : supported by D3D11
)

const (
	BLENDFACTOR_ZERO                = 0x1 // 0, 0, 0, 0
	BLENDFACTOR_ONE                 = 0x2 // 1, 1, 1, 1
	BLENDFACTOR_SRC_COLOR           = 0x3 // srcR, srcG, srcB, srcA
	BLENDFACTOR_ONE_MINUS_SRC_COLOR = 0x4 // 1-srcR, 1-srcG, 1-srcB, 1-srcA
	BLENDFACTOR_SRC_ALPHA           = 0x5 // srcA, srcA, srcA, srcA
	BLENDFACTOR_ONE_MINUS_SRC_ALPHA = 0x6 // 1-srcA, 1-srcA, 1-srcA, 1-srcA
	BLENDFACTOR_DST_COLOR           = 0x7 // dstR, dstG, dstB, dstA
	BLENDFACTOR_ONE_MINUS_DST_COLOR = 0x8 // 1-dstR, 1-dstG, 1-dstB, 1-dstA
	BLENDFACTOR_DST_ALPHA           = 0x9 // dstA, dstA, dstA, dstA
	BLENDFACTOR_ONE_MINUS_DST_ALPHA = 0xA // 1-dstA, 1-dstA, 1-dstA, 1-dstA
)

// Endian-specific values.
// (https://wiki.libsdl.org/CategoryEndian)
const (
	BYTEORDER  = LIL_ENDIAN // macro that corresponds to the byte order used by the processor type it was compiled for
	LIL_ENDIAN = 1234       // byte order is 1234, where the least significant byte is stored first
	BIG_ENDIAN = 4321       // byte order is 4321, where the most significant byte is stored first
)

// SDL error codes with their corresponding predefined strings.
const (
	ENOMEM      ErrorCode = iota // out of memory
	EFREAD                       // error reading from datastream
	EFWRITE                      // error writing to datastream
	EFSEEK                       // error seeking in datastream
	UNSUPPORTED                  // that operation is not supported
	LASTERROR                    // the highest numbered predefined error
)

// Enumeration of the types of events that can be delivered.
// (https://wiki.libsdl.org/SDL_EventType)
const (
	FIRSTEVENT = 0 // do not remove (unused)

	// Application events
	QUIT = 0x100 // user-requested quit

	// Android, iOS and WinRT events
	APP_TERMINATING         = 0x100 + 1 // OS is terminating the application
	APP_LOWMEMORY           = 0x100 + 2 // OS is low on memory; free some
	APP_WILLENTERBACKGROUND = 0x100 + 3 // application is entering background
	APP_DIDENTERBACKGROUND  = 0x100 + 4 //application entered background
	APP_WILLENTERFOREGROUND = 0x100 + 5 // application is entering foreground
	APP_DIDENTERFOREGROUND  = 0x100 + 6 // application entered foreground

	// Window events
	WINDOWEVENT = 0x200     // window state change
	SYSWMEVENT  = 0x200 + 1 // system specific event

	// Keyboard events
	KEYDOWN       = 0x300     // key pressed
	KEYUP         = 0x300 + 1 // key released
	TEXTEDITING   = 0x300 + 2 // keyboard text editing (composition)
	TEXTINPUT     = 0x300 + 3 // keyboard text input
	KEYMAPCHANGED = 0x300 + 4 // keymap changed due to a system event such as an input language or keyboard layout change (>= SDL 2.0.4)

	// Mouse events
	MOUSEMOTION     = 0x400     // mouse moved
	MOUSEBUTTONDOWN = 0x400 + 1 // mouse button pressed
	MOUSEBUTTONUP   = 0x400 + 2 // mouse button released
	MOUSEWHEEL      = 0x400 + 3 // mouse wheel motion

	// Joystick events
	JOYAXISMOTION    = 0x600     // joystick axis motion
	JOYBALLMOTION    = 0x600 + 1 // joystick trackball motion
	JOYHATMOTION     = 0x600 + 2 // joystick hat position change
	JOYBUTTONDOWN    = 0x600 + 3 // joystick button pressed
	JOYBUTTONUP      = 0x600 + 4 // joystick button released
	JOYDEVICEADDED   = 0x600 + 5 // joystick connected
	JOYDEVICEREMOVED = 0x600 + 6 // joystick disconnected

	// Game controller events
	CONTROLLERAXISMOTION     = 0x650     // controller axis motion
	CONTROLLERBUTTONDOWN     = 0x650 + 1 // controller button pressed
	CONTROLLERBUTTONUP       = 0x650 + 2 // controller button released
	CONTROLLERDEVICEADDED    = 0x650 + 3 // controller connected
	CONTROLLERDEVICEREMOVED  = 0x650 + 4 // controller disconnected
	CONTROLLERDEVICEREMAPPED = 0x650 + 5 // controller mapping updated

	// Touch events
	FINGERDOWN   = 0x700     // user has touched input device
	FINGERUP     = 0x700 + 1 // user stopped touching input device
	FINGERMOTION = 0x700 + 3 // user is dragging finger on input device

	// Gesture events
	DOLLARGESTURE = 0x800
	DOLLARRECORD  = 0x800 + 1
	MULTIGESTURE  = 0x800 + 2

	// Clipboard events
	CLIPBOARDUPDATE = 0x900 // the clipboard changed

	// Drag and drop events
	DROPFILE     = 0x1000     // the system requests a file open
	DROPTEXT     = 0x1000 + 1 // text/plain drag-and-drop event
	DROPBEGIN    = 0x1000 + 2 // a new set of drops is beginning (NULL filename)
	DROPCOMPLETE = 0x1000 + 3 // current set of drops is now complete (NULL filename)

	// Audio hotplug events
	AUDIODEVICEADDED   = 0x1100     // a new audio device is available (>= SDL 2.0.4)
	AUDIODEVICEREMOVED = 0x1100 + 1 // an audio device has been removed (>= SDL 2.0.4)

	// Sensor events
	SENSORUPDATE = 0x1200 // a sensor was updated

	// Render events
	RENDER_TARGETS_RESET = 0x2000     // the render targets have been reset and their contents need to be updated (>= SDL 2.0.2)
	RENDER_DEVICE_RESET  = 0x2000 + 1 // the device has been reset and all textures need to be recreated (>= SDL 2.0.4)

	// These are for your use, and should be allocated with RegisterEvents()
	USEREVENT = 0x8000 // a user-specified event
	LASTEVENT = 0xFFFF // (only for bounding internal arrays)
)

// Actions for PeepEvents().
// (https://wiki.libsdl.org/SDL_PeepEvents)
const (
	ADDEVENT  = iota // up to numevents events will be added to the back of the event queue
	PEEKEVENT        // up to numevents events at the front of the event queue, within the specified minimum and maximum type, will be returned and will not be removed from the queue
	GETEVENT         // up to numevents events at the front of the event queue, within the specified minimum and maximum type, will be returned and will be removed from the queue
)

// Toggles for different event state functions.
const (
	QUERY   = -1
	IGNORE  = 0
	DISABLE = 0
	ENABLE  = 1
)

// Types of game controller inputs.
const (
	CONTROLLER_BINDTYPE_NONE = iota
	CONTROLLER_BINDTYPE_BUTTON
	CONTROLLER_BINDTYPE_AXIS
	CONTROLLER_BINDTYPE_HAT
)

// An enumeration of axes available from a controller.
// (https://wiki.libsdl.org/SDL_GameControllerAxis)
const (
	CONTROLLER_AXIS_INVALID = iota - 1
	CONTROLLER_AXIS_LEFTX
	CONTROLLER_AXIS_LEFTY
	CONTROLLER_AXIS_RIGHTX
	CONTROLLER_AXIS_RIGHTY
	CONTROLLER_AXIS_TRIGGERLEFT
	CONTROLLER_AXIS_TRIGGERRIGHT
	CONTROLLER_AXIS_MAX
)

// An enumeration of buttons available from a controller.
// (https://wiki.libsdl.org/SDL_GameControllerButton)
const (
	CONTROLLER_BUTTON_INVALID = iota - 1
	CONTROLLER_BUTTON_A
	CONTROLLER_BUTTON_B
	CONTROLLER_BUTTON_X
	CONTROLLER_BUTTON_Y
	CONTROLLER_BUTTON_BACK
	CONTROLLER_BUTTON_GUIDE
	CONTROLLER_BUTTON_START
	CONTROLLER_BUTTON_LEFTSTICK
	CONTROLLER_BUTTON_RIGHTSTICK
	CONTROLLER_BUTTON_LEFTSHOULDER
	CONTROLLER_BUTTON_RIGHTSHOULDER
	CONTROLLER_BUTTON_DPAD_UP
	CONTROLLER_BUTTON_DPAD_DOWN
	CONTROLLER_BUTTON_DPAD_LEFT
	CONTROLLER_BUTTON_DPAD_RIGHT
	CONTROLLER_BUTTON_MAX
)

// Haptic effects.
// (https://wiki.libsdl.org/SDL_HapticEffect)
const (
	HAPTIC_CONSTANT     = 1 << iota // constant haptic effect
	HAPTIC_SINE                     // periodic haptic effect that simulates sine waves
	HAPTIC_LEFTRIGHT                // haptic effect for direct control over high/low frequency motors
	HAPTIC_TRIANGLE                 // periodic haptic effect that simulates triangular waves
	HAPTIC_SAWTOOTHUP               // periodic haptic effect that simulates saw tooth up waves
	HAPTIC_SAWTOOTHDOWN             // periodic haptic effect that simulates saw tooth down waves
	HAPTIC_RAMP                     // ramp haptic effect
	HAPTIC_SPRING                   // condition haptic effect that simulates a spring.  Effect is based on the axes position
	HAPTIC_DAMPER                   // condition haptic effect that simulates dampening.  Effect is based on the axes velocity
	HAPTIC_INERTIA                  // condition haptic effect that simulates inertia.  Effect is based on the axes acceleration
	HAPTIC_FRICTION                 // condition haptic effect that simulates friction.  Effect is based on the axes movement
	HAPTIC_CUSTOM                   // user defined custom haptic effect
	HAPTIC_GAIN                     // device supports setting the global gain
	HAPTIC_AUTOCENTER               // device supports setting autocenter
	HAPTIC_STATUS                   // device can be queried for effect status
	HAPTIC_PAUSE                    // device can be paused
)

// Direction encodings.
// (https://wiki.libsdl.org/SDL_HapticDirection)
const (
	HAPTIC_POLAR     = 0          // uses polar coordinates for the direction
	HAPTIC_CARTESIAN = 1          // uses cartesian coordinates for the direction
	HAPTIC_SPHERICAL = 2          // uses spherical coordinates for the direction
	HAPTIC_INFINITY  = 4294967295 // used to play a device an infinite number of times
)

// Configuration hints
// (https://wiki.libsdl.org/CategoryHints)
const (
	HINT_FRAMEBUFFER_ACCELERATION                 = "SDL_FRAMEBUFFER_ACCELERATION"                 // specifies how 3D acceleration is used with Window.GetSurface()
	HINT_RENDER_DRIVER                            = "SDL_RENDER_DRIVER"                            // specifies which render driver to use
	HINT_RENDER_OPENGL_SHADERS                    = "SDL_RENDER_OPENGL_SHADERS"                    // specifies whether the OpenGL render driver uses shaders
	HINT_RENDER_DIRECT3D_THREADSAFE               = "SDL_RENDER_DIRECT3D_THREADSAFE"               // specifies whether the Direct3D device is initialized for thread-safe operations
	HINT_RENDER_DIRECT3D11_DEBUG                  = "SDL_RENDER_DIRECT3D11_DEBUG"                  // specifies a variable controlling whether to enable Direct3D 11+'s Debug Layer
	HINT_RENDER_SCALE_QUALITY                     = "SDL_RENDER_SCALE_QUALITY"                     // specifies scaling quality
	HINT_RENDER_VSYNC                             = "SDL_RENDER_VSYNC"                             // specifies whether sync to vertical refresh is enabled or disabled in CreateRenderer() to avoid tearing
	HINT_VIDEO_ALLOW_SCREENSAVER                  = "SDL_VIDEO_ALLOW_SCREENSAVER"                  // specifies whether the screensaver is enabled
	HINT_VIDEO_X11_NET_WM_PING                    = "SDL_VIDEO_X11_NET_WM_PING"                    // specifies whether the X11 _NET_WM_PING protocol should be supported
	HINT_VIDEO_X11_XVIDMODE                       = "SDL_VIDEO_X11_XVIDMODE"                       // specifies whether the X11 VidMode extension should be used
	HINT_VIDEO_X11_XINERAMA                       = "SDL_VIDEO_X11_XINERAMA"                       // specifies whether the X11 Xinerama extension should be used
	HINT_VIDEO_X11_XRANDR                         = "SDL_VIDEO_X11_XRANDR"                         // specifies whether the X11 XRandR extension should be used
	HINT_GRAB_KEYBOARD                            = "SDL_GRAB_KEYBOARD"                            // specifies whether grabbing input grabs the keyboard
	HINT_MOUSE_RELATIVE_MODE_WARP                 = "SDL_MOUSE_RELATIVE_MODE_WARP"                 // specifies whether relative mouse mode is implemented using mouse warping
	HINT_VIDEO_MINIMIZE_ON_FOCUS_LOSS             = "SDL_VIDEO_MINIMIZE_ON_FOCUS_LOSS"             // specifies if a Window is minimized if it loses key focus when in fullscreen mode
	HINT_IDLE_TIMER_DISABLED                      = "SDL_IOS_IDLE_TIMER_DISABLED"                  // specifies a variable controlling whether the idle timer is disabled on iOS
	HINT_IME_INTERNAL_EDITING                     = "SDL_IME_INTERNAL_EDITING"                     // specifies whether certain IMEs should handle text editing internally instead of sending TextEditingEvents
	HINT_ORIENTATIONS                             = "SDL_IOS_ORIENTATIONS"                         // specifies a variable controlling which orientations are allowed on iOS
	HINT_ACCELEROMETER_AS_JOYSTICK                = "SDL_ACCELEROMETER_AS_JOYSTICK"                // specifies whether the Android / iOS built-in accelerometer should be listed as a joystick device, rather than listing actual joysticks only
	HINT_XINPUT_ENABLED                           = "SDL_XINPUT_ENABLED"                           // specifies if Xinput gamepad devices are detected
	HINT_XINPUT_USE_OLD_JOYSTICK_MAPPING          = "SDL_XINPUT_USE_OLD_JOYSTICK_MAPPING"          // specifies that SDL should use the old axis and button mapping for XInput devices
	HINT_GAMECONTROLLERCONFIG                     = "SDL_GAMECONTROLLERCONFIG"                     // specifies extra gamecontroller db entries
	HINT_JOYSTICK_ALLOW_BACKGROUND_EVENTS         = "SDL_JOYSTICK_ALLOW_BACKGROUND_EVENTS"         // specifies if joystick (and gamecontroller) events are enabled even when the application is in the background
	HINT_ALLOW_TOPMOST                            = "SDL_ALLOW_TOPMOST"                            // specifies if top most bit on an SDL Window can be set
	HINT_THREAD_STACK_SIZE                        = "SDL_THREAD_STACK_SIZE"                        // specifies a variable specifying SDL's threads stack size in bytes or "0" for the backend's default size
	HINT_TIMER_RESOLUTION                         = "SDL_TIMER_RESOLUTION"                         // specifies the timer resolution in milliseconds
	HINT_VIDEO_HIGHDPI_DISABLED                   = "SDL_VIDEO_HIGHDPI_DISABLED"                   // specifies if high-DPI windows ("Retina" on Mac and iOS) are not allowed
	HINT_MAC_BACKGROUND_APP                       = "SDL_MAC_BACKGROUND_APP"                       // specifies if the SDL app should not be forced to become a foreground process on Mac OS X
	HINT_MAC_CTRL_CLICK_EMULATE_RIGHT_CLICK       = "SDL_MAC_CTRL_CLICK_EMULATE_RIGHT_CLICK"       // specifies whether ctrl+click should generate a right-click event on Mac
	HINT_VIDEO_WIN_D3DCOMPILER                    = "SDL_VIDEO_WIN_D3DCOMPILER"                    // specifies which shader compiler to preload when using the Chrome ANGLE binaries
	HINT_VIDEO_WINDOW_SHARE_PIXEL_FORMAT          = "SDL_VIDEO_WINDOW_SHARE_PIXEL_FORMAT"          // specifies the address of another Window* (as a hex string formatted with "%p")
	HINT_WINRT_PRIVACY_POLICY_URL                 = "SDL_WINRT_PRIVACY_POLICY_URL"                 // specifies a URL to a WinRT app's privacy policy
	HINT_WINRT_PRIVACY_POLICY_LABEL               = "SDL_WINRT_PRIVACY_POLICY_LABEL"               // specifies a label text for a WinRT app's privacy policy link
	HINT_WINRT_HANDLE_BACK_BUTTON                 = "SDL_WINRT_HANDLE_BACK_BUTTON"                 // specifies a variable to allow back-button-press events on Windows Phone to be marked as handled
	HINT_VIDEO_MAC_FULLSCREEN_SPACES              = "SDL_VIDEO_MAC_FULLSCREEN_SPACES"              // specifies policy for fullscreen Spaces on Mac OS X
	HINT_NO_SIGNAL_HANDLERS                       = "SDL_NO_SIGNAL_HANDLERS"                       // specifies not to catch the SIGINT or SIGTERM signals
	HINT_WINDOW_FRAME_USABLE_WHILE_CURSOR_HIDDEN  = "SDL_WINDOW_FRAME_USABLE_WHILE_CURSOR_HIDDEN"  // specifies whether the window frame and title bar are interactive when the cursor is hidden
	HINT_WINDOWS_ENABLE_MESSAGELOOP               = "SDL_WINDOWS_ENABLE_MESSAGELOOP"               // specifies whether the windows message loop is processed by SDL
	HINT_WINDOWS_NO_CLOSE_ON_ALT_F4               = "SDL_WINDOWS_NO_CLOSE_ON_ALT_F4"               // specifies that SDL should not to generate WINDOWEVENT_CLOSE events for Alt+F4 on Microsoft Windows
	HINT_ANDROID_SEPARATE_MOUSE_AND_TOUCH         = "SDL_ANDROID_SEPARATE_MOUSE_AND_TOUCH"         // specifies a variable to control whether mouse and touch events are to be treated together or separately
	HINT_ANDROID_APK_EXPANSION_MAIN_FILE_VERSION  = "SDL_ANDROID_APK_EXPANSION_MAIN_FILE_VERSION"  // specifies the Android APK expansion main file version
	HINT_ANDROID_APK_EXPANSION_PATCH_FILE_VERSION = "SDL_ANDROID_APK_EXPANSION_PATCH_FILE_VERSION" // specifies the Android APK expansion patch file version
	HINT_AUDIO_RESAMPLING_MODE                    = "SDL_AUDIO_RESAMPLING_MODE"                    // specifies a variable controlling speed/quality tradeoff of audio resampling
	HINT_RENDER_LOGICAL_SIZE_MODE                 = "SDL_RENDER_LOGICAL_SIZE_MODE"                 // specifies a variable controlling the scaling policy for SDL_RenderSetLogicalSize
	HINT_MOUSE_NORMAL_SPEED_SCALE                 = "SDL_MOUSE_NORMAL_SPEED_SCALE"                 // specifies a variable setting the speed scale for mouse motion, in floating point, when the mouse is not in relative mode
	HINT_MOUSE_RELATIVE_SPEED_SCALE               = "SDL_MOUSE_RELATIVE_SPEED_SCALE"               // specifies a variable setting the scale for mouse motion, in floating point, when the mouse is in relative mode
	HINT_TOUCH_MOUSE_EVENTS                       = "SDL_TOUCH_MOUSE_EVENTS"                       // specifies a variable controlling whether touch events should generate synthetic mouse events
	HINT_WINDOWS_INTRESOURCE_ICON                 = "SDL_WINDOWS_INTRESOURCE_ICON"                 // specifies a variable to specify custom icon resource id from RC file on Windows platform
	HINT_WINDOWS_INTRESOURCE_ICON_SMALL           = "SDL_WINDOWS_INTRESOURCE_ICON_SMALL"           // specifies a variable to specify custom icon resource id from RC file on Windows platform
	HINT_IOS_HIDE_HOME_INDICATOR                  = "SDL_IOS_HIDE_HOME_INDICATOR"                  // specifies a variable controlling whether the home indicator bar on iPhone X should be hidden.
	HINT_RETURN_KEY_HIDES_IME                     = "SDL_RETURN_KEY_HIDES_IME"                     // specifies a variable to control whether the return key on the soft keyboard should hide the soft keyboard on Android and iOS.
	HINT_TV_REMOTE_AS_JOYSTICK                    = "SDL_TV_REMOTE_AS_JOYSTICK"                    // specifies a variable controlling whether the Android / tvOS remotes  should be listed as joystick devices, instead of sending keyboard events.
	HINT_VIDEO_X11_NET_WM_BYPASS_COMPOSITOR       = "SDL_VIDEO_X11_NET_WM_BYPASS_COMPOSITOR"       // specifies a variable controlling whether the X11 _NET_WM_BYPASS_COMPOSITOR hint should be used.
	HINT_VIDEO_DOUBLE_BUFFER                      = "SDL_VIDEO_DOUBLE_BUFFER"                      // specifies a variable that tells the video driver that we only want a double buffer.
)

// An enumeration of hint priorities.
// (https://wiki.libsdl.org/SDL_HintPriority)
const (
	HINT_DEFAULT  = iota // low priority, used for default values
	HINT_NORMAL          // medium priority
	HINT_OVERRIDE        // high priority
)

// Hat positions.
// (https://wiki.libsdl.org/SDL_JoystickGetHat)
const (
	HAT_CENTERED  = 0x00
	HAT_UP        = 0x01
	HAT_RIGHT     = 0x02
	HAT_DOWN      = 0x04
	HAT_LEFT      = 0x08
	HAT_RIGHTUP   = HAT_RIGHT | HAT_UP
	HAT_RIGHTDOWN = HAT_RIGHT | HAT_DOWN
	HAT_LEFTUP    = HAT_LEFT | HAT_UP
	HAT_LEFTDOWN  = HAT_LEFT | HAT_DOWN
)

// Types of a joystick.
const (
	JOYSTICK_TYPE_UNKNOWN = iota
	JOYSTICK_TYPE_GAMECONTROLLER
	JOYSTICK_TYPE_WHEEL
	JOYSTICK_TYPE_ARCADE_STICK
	JOYSTICK_TYPE_FLIGHT_STICK
	JOYSTICK_TYPE_DANCE_PAD
	JOYSTICK_TYPE_GUITAR
	JOYSTICK_TYPE_DRUM_KIT
	JOYSTICK_TYPE_ARCADE_PAD
	JOYSTICK_TYPE_THROTTLE
)

// An enumeration of battery levels of a joystick.
// (https://wiki.libsdl.org/SDL_JoystickPowerLevel)
const (
	JOYSTICK_POWER_UNKNOWN = iota - 1
	JOYSTICK_POWER_EMPTY
	JOYSTICK_POWER_LOW
	JOYSTICK_POWER_MEDIUM
	JOYSTICK_POWER_FULL
	JOYSTICK_POWER_WIRED
	JOYSTICK_POWER_MAX
)

// The SDL virtual key representation.
// (https://wiki.libsdl.org/SDL_Keycode)
// (https://wiki.libsdl.org/SDLKeycodeLookup)
const (
	K_UNKNOWN = 0 // "" (no name, empty string)

	K_RETURN     = '\r'   // "Return" (the Enter key (main keyboard))
	K_ESCAPE     = '\033' // "Escape" (the Esc key)
	K_BACKSPACE  = '\b'   // "Backspace"
	K_TAB        = '\t'   // "Tab" (the Tab key)
	K_SPACE      = ' '    // "Space" (the Space Bar key(s))
	K_EXCLAIM    = '!'    // "!"
	K_QUOTEDBL   = '"'    // """
	K_HASH       = '#'    // "#"
	K_PERCENT    = '%'    // "%"
	K_DOLLAR     = '$'    // "$"
	K_AMPERSAND  = '&'    // "&"
	K_QUOTE      = '\''   // "'"
	K_LEFTPAREN  = '('    // "("
	K_RIGHTPAREN = ')'    // ")"
	K_ASTERISK   = '*'    // "*"
	K_PLUS       = '+'    // "+"
	K_COMMA      = ','    // ","
	K_MINUS      = '-'    // "-"
	K_PERIOD     = '.'    // "."
	K_SLASH      = '/'    // "/"
	K_0          = '0'    // "0"
	K_1          = '1'    // "1"
	K_2          = '2'    // "2"
	K_3          = '3'    // "3"
	K_4          = '4'    // "4"
	K_5          = '5'    // "5"
	K_6          = '6'    // "6"
	K_7          = '7'    // "7"
	K_8          = '8'    // "8"
	K_9          = '9'    // "9"
	K_COLON      = ':'    // ":"
	K_SEMICOLON  = ';'    // ";"
	K_LESS       = '<'    // "<"
	K_EQUALS     = '='    // "="
	K_GREATER    = '>'    // ">"
	K_QUESTION   = '?'    // "?"
	K_AT         = '@'    // "@"
	/*
	   Skip uppercase letters
	*/
	K_LEFTBRACKET  = '['  // "["
	K_BACKSLASH    = '\\' // "\"
	K_RIGHTBRACKET = ']'  // "]"
	K_CARET        = '^'  // "^"
	K_UNDERSCORE   = '_'  // "_"
	K_BACKQUOTE    = '`'  // "`"
	K_a            = 'a'  // "A"
	K_b            = 'b'  // "B"
	K_c            = 'c'  // "C"
	K_d            = 'd'  // "D"
	K_e            = 'e'  // "E"
	K_f            = 'f'  // "F"
	K_g            = 'g'  // "G"
	K_h            = 'h'  // "H"
	K_i            = 'i'  // "I"
	K_j            = 'j'  // "J"
	K_k            = 'k'  // "K"
	K_l            = 'l'  // "L"
	K_m            = 'm'  // "M"
	K_n            = 'n'  // "N"
	K_o            = 'o'  // "O"
	K_p            = 'p'  // "P"
	K_q            = 'q'  // "Q"
	K_r            = 'r'  // "R"
	K_s            = 's'  // "S"
	K_t            = 't'  // "T"
	K_u            = 'u'  // "U"
	K_v            = 'v'  // "V"
	K_w            = 'w'  // "W"
	K_x            = 'x'  // "X"
	K_y            = 'y'  // "Y"
	K_z            = 'z'  // "Z"

	K_CAPSLOCK = SCANCODE_CAPSLOCK | K_SCANCODE_MASK // "CapsLock"

	K_F1  = SCANCODE_F1 | K_SCANCODE_MASK  // "F1"
	K_F2  = SCANCODE_F2 | K_SCANCODE_MASK  // "F2"
	K_F3  = SCANCODE_F3 | K_SCANCODE_MASK  // "F3"
	K_F4  = SCANCODE_F4 | K_SCANCODE_MASK  // "F4"
	K_F5  = SCANCODE_F5 | K_SCANCODE_MASK  // "F5"
	K_F6  = SCANCODE_F6 | K_SCANCODE_MASK  // "F6"
	K_F7  = SCANCODE_F7 | K_SCANCODE_MASK  // "F7"
	K_F8  = SCANCODE_F8 | K_SCANCODE_MASK  // "F8"
	K_F9  = SCANCODE_F9 | K_SCANCODE_MASK  // "F9"
	K_F10 = SCANCODE_F10 | K_SCANCODE_MASK // "F10"
	K_F11 = SCANCODE_F11 | K_SCANCODE_MASK // "F11"
	K_F12 = SCANCODE_F12 | K_SCANCODE_MASK // "F12"

	K_PRINTSCREEN = SCANCODE_PRINTSCREEN | K_SCANCODE_MASK // "PrintScreen"
	K_SCROLLLOCK  = SCANCODE_SCROLLLOCK | K_SCANCODE_MASK  // "ScrollLock"
	K_PAUSE       = SCANCODE_PAUSE | K_SCANCODE_MASK       // "Pause" (the Pause / Break key)
	K_INSERT      = SCANCODE_INSERT | K_SCANCODE_MASK      // "Insert" (insert on PC, help on some Mac keyboards (but does send code 73, not 117))
	K_HOME        = SCANCODE_HOME | K_SCANCODE_MASK        // "Home"
	K_PAGEUP      = SCANCODE_PAGEUP | K_SCANCODE_MASK      // "PageUp"
	K_DELETE      = '\177'                                 // "Delete"
	K_END         = SCANCODE_END | K_SCANCODE_MASK         // "End"
	K_PAGEDOWN    = SCANCODE_PAGEDOWN | K_SCANCODE_MASK    // "PageDown"
	K_RIGHT       = SCANCODE_RIGHT | K_SCANCODE_MASK       // "Right" (the Right arrow key (navigation keypad))
	K_LEFT        = SCANCODE_LEFT | K_SCANCODE_MASK        // "Left" (the Left arrow key (navigation keypad))
	K_DOWN        = SCANCODE_DOWN | K_SCANCODE_MASK        // "Down" (the Down arrow key (navigation keypad))
	K_UP          = SCANCODE_UP | K_SCANCODE_MASK          // "Up" (the Up arrow key (navigation keypad))

	K_NUMLOCKCLEAR = SCANCODE_NUMLOCKCLEAR | K_SCANCODE_MASK // "Numlock" (the Num Lock key (PC) / the Clear key (Mac))
	K_KP_DIVIDE    = SCANCODE_KP_DIVIDE | K_SCANCODE_MASK    // "Keypad /" (the / key (numeric keypad))
	K_KP_MULTIPLY  = SCANCODE_KP_MULTIPLY | K_SCANCODE_MASK  // "Keypad *" (the * key (numeric keypad))
	K_KP_MINUS     = SCANCODE_KP_MINUS | K_SCANCODE_MASK     // "Keypad -" (the - key (numeric keypad))
	K_KP_PLUS      = SCANCODE_KP_PLUS | K_SCANCODE_MASK      // "Keypad +" (the + key (numeric keypad))
	K_KP_ENTER     = SCANCODE_KP_ENTER | K_SCANCODE_MASK     // "Keypad Enter" (the Enter key (numeric keypad))
	K_KP_1         = SCANCODE_KP_1 | K_SCANCODE_MASK         // "Keypad 1" (the 1 key (numeric keypad))
	K_KP_2         = SCANCODE_KP_2 | K_SCANCODE_MASK         // "Keypad 2" (the 2 key (numeric keypad))
	K_KP_3         = SCANCODE_KP_3 | K_SCANCODE_MASK         // "Keypad 3" (the 3 key (numeric keypad))
	K_KP_4         = SCANCODE_KP_4 | K_SCANCODE_MASK         // "Keypad 4" (the 4 key (numeric keypad))
	K_KP_5         = SCANCODE_KP_5 | K_SCANCODE_MASK         // "Keypad 5" (the 5 key (numeric keypad))
	K_KP_6         = SCANCODE_KP_6 | K_SCANCODE_MASK         // "Keypad 6" (the 6 key (numeric keypad))
	K_KP_7         = SCANCODE_KP_7 | K_SCANCODE_MASK         // "Keypad 7" (the 7 key (numeric keypad))
	K_KP_8         = SCANCODE_KP_8 | K_SCANCODE_MASK         // "Keypad 8" (the 8 key (numeric keypad))
	K_KP_9         = SCANCODE_KP_9 | K_SCANCODE_MASK         // "Keypad 9" (the 9 key (numeric keypad))
	K_KP_0         = SCANCODE_KP_0 | K_SCANCODE_MASK         // "Keypad 0" (the 0 key (numeric keypad))
	K_KP_PERIOD    = SCANCODE_KP_PERIOD | K_SCANCODE_MASK    // "Keypad ." (the . key (numeric keypad))

	K_APPLICATION    = SCANCODE_APPLICATION | K_SCANCODE_MASK    // "Application" (the Application / Compose / Context Menu (Windows) key)
	K_POWER          = SCANCODE_POWER | K_SCANCODE_MASK          // "Power" (The USB document says this is a status flag, not a physical key - but some Mac keyboards do have a power key.)
	K_KP_EQUALS      = SCANCODE_EQUALS | K_SCANCODE_MASK         // "Keypad =" (the = key (numeric keypad))
	K_F13            = SCANCODE_F13 | K_SCANCODE_MASK            // "F13"
	K_F14            = SCANCODE_F14 | K_SCANCODE_MASK            // "F14"
	K_F15            = SCANCODE_F15 | K_SCANCODE_MASK            // "F15"
	K_F16            = SCANCODE_F16 | K_SCANCODE_MASK            // "F16"
	K_F17            = SCANCODE_F17 | K_SCANCODE_MASK            // "F17"
	K_F18            = SCANCODE_F18 | K_SCANCODE_MASK            // "F18"
	K_F19            = SCANCODE_F19 | K_SCANCODE_MASK            // "F19"
	K_F20            = SCANCODE_F20 | K_SCANCODE_MASK            // "F20"
	K_F21            = SCANCODE_F21 | K_SCANCODE_MASK            // "F21"
	K_F22            = SCANCODE_F22 | K_SCANCODE_MASK            // "F22"
	K_F23            = SCANCODE_F23 | K_SCANCODE_MASK            // "F23"
	K_F24            = SCANCODE_F24 | K_SCANCODE_MASK            // "F24"
	K_EXECUTE        = SCANCODE_EXECUTE | K_SCANCODE_MASK        // "Execute"
	K_HELP           = SCANCODE_HELP | K_SCANCODE_MASK           // "Help"
	K_MENU           = SCANCODE_MENU | K_SCANCODE_MASK           // "Menu"
	K_SELECT         = SCANCODE_SELECT | K_SCANCODE_MASK         // "Select"
	K_STOP           = SCANCODE_STOP | K_SCANCODE_MASK           // "Stop"
	K_AGAIN          = SCANCODE_AGAIN | K_SCANCODE_MASK          // "Again" (the Again key (Redo))
	K_UNDO           = SCANCODE_UNDO | K_SCANCODE_MASK           // "Undo"
	K_CUT            = SCANCODE_CUT | K_SCANCODE_MASK            // "Cut"
	K_COPY           = SCANCODE_COPY | K_SCANCODE_MASK           // "Copy"
	K_PASTE          = SCANCODE_PASTE | K_SCANCODE_MASK          // "Paste"
	K_FIND           = SCANCODE_FIND | K_SCANCODE_MASK           // "Find"
	K_MUTE           = SCANCODE_MUTE | K_SCANCODE_MASK           // "Mute"
	K_VOLUMEUP       = SCANCODE_VOLUMEUP | K_SCANCODE_MASK       // "VolumeUp"
	K_VOLUMEDOWN     = SCANCODE_VOLUMEDOWN | K_SCANCODE_MASK     // "VolumeDown"
	K_KP_COMMA       = SCANCODE_KP_COMMA | K_SCANCODE_MASK       // "Keypad ," (the Comma key (numeric keypad))
	K_KP_EQUALSAS400 = SCANCODE_KP_EQUALSAS400 | K_SCANCODE_MASK // "Keypad = (AS400)" (the Equals AS400 key (numeric keypad))

	K_ALTERASE   = SCANCODE_ALTERASE | K_SCANCODE_MASK   // "AltErase" (Erase-Eaze)
	K_SYSREQ     = SCANCODE_SYSREQ | K_SCANCODE_MASK     // "SysReq" (the SysReq key)
	K_CANCEL     = SCANCODE_CANCEL | K_SCANCODE_MASK     // "Cancel"
	K_CLEAR      = SCANCODE_CLEAR | K_SCANCODE_MASK      // "Clear"
	K_PRIOR      = SCANCODE_PRIOR | K_SCANCODE_MASK      // "Prior"
	K_RETURN2    = SCANCODE_RETURN2 | K_SCANCODE_MASK    // "Return"
	K_SEPARATOR  = SCANCODE_SEPARATOR | K_SCANCODE_MASK  // "Separator"
	K_OUT        = SCANCODE_OUT | K_SCANCODE_MASK        // "Out"
	K_OPER       = SCANCODE_OPER | K_SCANCODE_MASK       // "Oper"
	K_CLEARAGAIN = SCANCODE_CLEARAGAIN | K_SCANCODE_MASK // "Clear / Again"
	K_CRSEL      = SCANCODE_CRSEL | K_SCANCODE_MASK      // "CrSel"
	K_EXSEL      = SCANCODE_EXSEL | K_SCANCODE_MASK      // "ExSel"

	K_KP_00              = SCANCODE_KP_00 | K_SCANCODE_MASK              // "Keypad 00" (the 00 key (numeric keypad))
	K_KP_000             = SCANCODE_KP_000 | K_SCANCODE_MASK             // "Keypad 000" (the 000 key (numeric keypad))
	K_THOUSANDSSEPARATOR = SCANCODE_THOUSANDSSEPARATOR | K_SCANCODE_MASK // "ThousandsSeparator" (the Thousands Separator key)
	K_DECIMALSEPARATOR   = SCANCODE_DECIMALSEPARATOR | K_SCANCODE_MASK   // "DecimalSeparator" (the Decimal Separator key)
	K_CURRENCYUNIT       = SCANCODE_CURRENCYUNIT | K_SCANCODE_MASK       // "CurrencyUnit" (the Currency Unit key)
	K_CURRENCYSUBUNIT    = SCANCODE_CURRENCYSUBUNIT | K_SCANCODE_MASK    // "CurrencySubUnit" (the Currency Subunit key)
	K_KP_LEFTPAREN       = SCANCODE_KP_LEFTPAREN | K_SCANCODE_MASK       // "Keypad (" (the Left Parenthesis key (numeric keypad))
	K_KP_RIGHTPAREN      = SCANCODE_KP_RIGHTPAREN | K_SCANCODE_MASK      // "Keypad )" (the Right Parenthesis key (numeric keypad))
	K_KP_LEFTBRACE       = SCANCODE_KP_LEFTBRACE | K_SCANCODE_MASK       // "Keypad {" (the Left Brace key (numeric keypad))
	K_KP_RIGHTBRACE      = SCANCODE_KP_RIGHTBRACE | K_SCANCODE_MASK      // "Keypad }" (the Right Brace key (numeric keypad))
	K_KP_TAB             = SCANCODE_KP_TAB | K_SCANCODE_MASK             // "Keypad Tab" (the Tab key (numeric keypad))
	K_KP_BACKSPACE       = SCANCODE_KP_BACKSPACE | K_SCANCODE_MASK       // "Keypad Backspace" (the Backspace key (numeric keypad))
	K_KP_A               = SCANCODE_KP_A | K_SCANCODE_MASK               // "Keypad A" (the A key (numeric keypad))
	K_KP_B               = SCANCODE_KP_B | K_SCANCODE_MASK               // "Keypad B" (the B key (numeric keypad))
	K_KP_C               = SCANCODE_KP_C | K_SCANCODE_MASK               // "Keypad C" (the C key (numeric keypad))
	K_KP_D               = SCANCODE_KP_D | K_SCANCODE_MASK               // "Keypad D" (the D key (numeric keypad))
	K_KP_E               = SCANCODE_KP_E | K_SCANCODE_MASK               // "Keypad E" (the E key (numeric keypad))
	K_KP_F               = SCANCODE_KP_F | K_SCANCODE_MASK               // "Keypad F" (the F key (numeric keypad))
	K_KP_XOR             = SCANCODE_KP_XOR | K_SCANCODE_MASK             // "Keypad XOR" (the XOR key (numeric keypad))
	K_KP_POWER           = SCANCODE_KP_POWER | K_SCANCODE_MASK           // "Keypad ^" (the Power key (numeric keypad))
	K_KP_PERCENT         = SCANCODE_KP_PERCENT | K_SCANCODE_MASK         // "Keypad %" (the Percent key (numeric keypad))
	K_KP_LESS            = SCANCODE_KP_LESS | K_SCANCODE_MASK            // "Keypad <" (the Less key (numeric keypad))
	K_KP_GREATER         = SCANCODE_KP_GREATER | K_SCANCODE_MASK         // "Keypad >" (the Greater key (numeric keypad))
	K_KP_AMPERSAND       = SCANCODE_KP_AMPERSAND | K_SCANCODE_MASK       // "Keypad &" (the & key (numeric keypad))
	K_KP_DBLAMPERSAND    = SCANCODE_KP_DBLAMPERSAND | K_SCANCODE_MASK    // "Keypad &&" (the && key (numeric keypad))
	K_KP_VERTICALBAR     = SCANCODE_KP_VERTICALBAR | K_SCANCODE_MASK     // "Keypad |" (the | key (numeric keypad))
	K_KP_DBLVERTICALBAR  = SCANCODE_KP_DBLVERTICALBAR | K_SCANCODE_MASK  // "Keypad ||" (the || key (numeric keypad))
	K_KP_COLON           = SCANCODE_KP_COLON | K_SCANCODE_MASK           // "Keypad :" (the : key (numeric keypad))
	K_KP_HASH            = SCANCODE_KP_HASH | K_SCANCODE_MASK            // "Keypad #" (the # key (numeric keypad))
	K_KP_SPACE           = SCANCODE_KP_SPACE | K_SCANCODE_MASK           // "Keypad Space" (the Space key (numeric keypad))
	K_KP_AT              = SCANCODE_KP_AT | K_SCANCODE_MASK              // "Keypad @" (the @ key (numeric keypad))
	K_KP_EXCLAM          = SCANCODE_KP_EXCLAM | K_SCANCODE_MASK          // "Keypad !" (the ! key (numeric keypad))
	K_KP_MEMSTORE        = SCANCODE_KP_MEMSTORE | K_SCANCODE_MASK        // "Keypad MemStore" (the Mem Store key (numeric keypad))
	K_KP_MEMRECALL       = SCANCODE_KP_MEMRECALL | K_SCANCODE_MASK       // "Keypad MemRecall" (the Mem Recall key (numeric keypad))
	K_KP_MEMCLEAR        = SCANCODE_KP_MEMCLEAR | K_SCANCODE_MASK        // "Keypad MemClear" (the Mem Clear key (numeric keypad))
	K_KP_MEMADD          = SCANCODE_KP_MEMADD | K_SCANCODE_MASK          // "Keypad MemAdd" (the Mem Add key (numeric keypad))
	K_KP_MEMSUBTRACT     = SCANCODE_KP_MEMSUBTRACT | K_SCANCODE_MASK     // "Keypad MemSubtract" (the Mem Subtract key (numeric keypad))
	K_KP_MEMMULTIPLY     = SCANCODE_KP_MEMMULTIPLY | K_SCANCODE_MASK     // "Keypad MemMultiply" (the Mem Multiply key (numeric keypad))
	K_KP_MEMDIVIDE       = SCANCODE_KP_MEMDIVIDE | K_SCANCODE_MASK       // "Keypad MemDivide" (the Mem Divide key (numeric keypad))
	K_KP_PLUSMINUS       = SCANCODE_KP_PLUSMINUS | K_SCANCODE_MASK       // "Keypad +/-" (the +/- key (numeric keypad))
	K_KP_CLEAR           = SCANCODE_KP_CLEAR | K_SCANCODE_MASK           // "Keypad Clear" (the Clear key (numeric keypad))
	K_KP_CLEARENTRY      = SCANCODE_KP_CLEARENTRY | K_SCANCODE_MASK      // "Keypad ClearEntry" (the Clear Entry key (numeric keypad))
	K_KP_BINARY          = SCANCODE_KP_BINARY | K_SCANCODE_MASK          // "Keypad Binary" (the Binary key (numeric keypad))
	K_KP_OCTAL           = SCANCODE_KP_OCTAL | K_SCANCODE_MASK           // "Keypad Octal" (the Octal key (numeric keypad))
	K_KP_DECIMAL         = SCANCODE_KP_DECIMAL | K_SCANCODE_MASK         // "Keypad Decimal" (the Decimal key (numeric keypad))
	K_KP_HEXADECIMAL     = SCANCODE_KP_HEXADECIMAL | K_SCANCODE_MASK     // "Keypad Hexadecimal" (the Hexadecimal key (numeric keypad))

	K_LCTRL  = SCANCODE_LCTRL | K_SCANCODE_MASK  // "Left Ctrl"
	K_LSHIFT = SCANCODE_LSHIFT | K_SCANCODE_MASK // "Left Shift"
	K_LALT   = SCANCODE_LALT | K_SCANCODE_MASK   // "Left Alt" (alt, option)
	K_LGUI   = SCANCODE_LGUI | K_SCANCODE_MASK   // "Left GUI" (windows, command (apple), meta)
	K_RCTRL  = SCANCODE_RCTRL | K_SCANCODE_MASK  // "Right Ctrl"
	K_RSHIFT = SCANCODE_RSHIFT | K_SCANCODE_MASK // "Right Shift"
	K_RALT   = SCANCODE_RALT | K_SCANCODE_MASK   // "Right Alt" (alt, option)
	K_RGUI   = SCANCODE_RGUI | K_SCANCODE_MASK   // "Right GUI" (windows, command (apple), meta)

	K_MODE = SCANCODE_MODE | K_SCANCODE_MASK // "ModeSwitch" (I'm not sure if this is really not covered by any of the above, but since there's a special KMOD_MODE for it I'm adding it here)

	K_AUDIONEXT    = SCANCODE_AUDIONEXT | K_SCANCODE_MASK    // "AudioNext" (the Next Track media key)
	K_AUDIOPREV    = SCANCODE_AUDIOPREV | K_SCANCODE_MASK    // "AudioPrev" (the Previous Track media key)
	K_AUDIOSTOP    = SCANCODE_AUDIOSTOP | K_SCANCODE_MASK    // "AudioStop" (the Stop media key)
	K_AUDIOPLAY    = SCANCODE_AUDIOPLAY | K_SCANCODE_MASK    // "AudioPlay" (the Play media key)
	K_AUDIOMUTE    = SCANCODE_AUDIOMUTE | K_SCANCODE_MASK    // "AudioMute" (the Mute volume key)
	K_MEDIASELECT  = SCANCODE_MEDIASELECT | K_SCANCODE_MASK  // "MediaSelect" (the Media Select key)
	K_WWW          = SCANCODE_WWW | K_SCANCODE_MASK          // "WWW" (the WWW/World Wide Web key)
	K_MAIL         = SCANCODE_MAIL | K_SCANCODE_MASK         // "Mail" (the Mail/eMail key)
	K_CALCULATOR   = SCANCODE_CALCULATOR | K_SCANCODE_MASK   // "Calculator" (the Calculator key)
	K_COMPUTER     = SCANCODE_COMPUTER | K_SCANCODE_MASK     // "Computer" (the My Computer key)
	K_AC_SEARCH    = SCANCODE_AC_SEARCH | K_SCANCODE_MASK    // "AC Search" (the Search key (application control keypad))
	K_AC_HOME      = SCANCODE_AC_HOME | K_SCANCODE_MASK      // "AC Home" (the Home key (application control keypad))
	K_AC_BACK      = SCANCODE_AC_BACK | K_SCANCODE_MASK      // "AC Back" (the Back key (application control keypad))
	K_AC_FORWARD   = SCANCODE_AC_FORWARD | K_SCANCODE_MASK   // "AC Forward" (the Forward key (application control keypad))
	K_AC_STOP      = SCANCODE_AC_STOP | K_SCANCODE_MASK      // "AC Stop" (the Stop key (application control keypad))
	K_AC_REFRESH   = SCANCODE_AC_REFRESH | K_SCANCODE_MASK   // "AC Refresh" (the Refresh key (application control keypad))
	K_AC_BOOKMARKS = SCANCODE_AC_BOOKMARKS | K_SCANCODE_MASK // "AC Bookmarks" (the Bookmarks key (application control keypad))

	K_BRIGHTNESSDOWN = SCANCODE_BRIGHTNESSDOWN | K_SCANCODE_MASK // "BrightnessDown" (the Brightness Down key)
	K_BRIGHTNESSUP   = SCANCODE_BRIGHTNESSUP | K_SCANCODE_MASK   // "BrightnessUp" (the Brightness Up key)
	K_DISPLAYSWITCH  = SCANCODE_DISPLAYSWITCH | K_SCANCODE_MASK  // "DisplaySwitch" (display mirroring/dual display switch, video mode switch)
	K_KBDILLUMTOGGLE = SCANCODE_KBDILLUMTOGGLE | K_SCANCODE_MASK // "KBDIllumToggle" (the Keyboard Illumination Toggle key)
	K_KBDILLUMDOWN   = SCANCODE_KBDILLUMDOWN | K_SCANCODE_MASK   // "KBDIllumDown" (the Keyboard Illumination Down key)
	K_KBDILLUMUP     = SCANCODE_KBDILLUMUP | K_SCANCODE_MASK     // "KBDIllumUp" (the Keyboard Illumination Up key)
	K_EJECT          = SCANCODE_EJECT | K_SCANCODE_MASK          // "Eject" (the Eject key)
	K_SLEEP          = SCANCODE_SLEEP | K_SCANCODE_MASK          // "Sleep" (the Sleep key)
	K_APP1           = SCANCODE_APP1 | K_SCANCODE_MASK
	K_APP2           = SCANCODE_APP2 | K_SCANCODE_MASK

	K_AUDIOREWIND      = SCANCODE_AUDIOREWIND | K_SCANCODE_MASK
	K_AUDIOFASTFORWARD = SCANCODE_AUDIOFASTFORWARD | K_SCANCODE_MASK
)

// An enumeration of key modifier masks.
// (https://wiki.libsdl.org/SDL_Keymod)
const (
	KMOD_NONE     = 0x0000                    // 0 (no modifier is applicable)
	KMOD_LSHIFT   = 0x0001                    // the left Shift key is down
	KMOD_RSHIFT   = 0x0002                    // the right Shift key is down
	KMOD_LCTRL    = 0x0040                    // the left Ctrl (Control) key is down
	KMOD_RCTRL    = 0x0080                    // the right Ctrl (Control) key is down
	KMOD_LALT     = 0x0100                    // the left Alt key is down
	KMOD_RALT     = 0x0200                    // the right Alt key is down
	KMOD_LGUI     = 0x0400                    // the left GUI key (often the Windows key) is down
	KMOD_RGUI     = 0x0800                    // the right GUI key (often the Windows key) is down
	KMOD_NUM      = 0x1000                    // the Num Lock key (may be located on an extended keypad) is down
	KMOD_CAPS     = 0x2000                    // the Caps Lock key is down
	KMOD_MODE     = 0x4000                    // the AltGr key is down
	KMOD_CTRL     = KMOD_LCTRL | KMOD_RCTRL   // (KMOD_LCTRL|KMOD_RCTRL)
	KMOD_SHIFT    = KMOD_LSHIFT | KMOD_RSHIFT // (KMOD_LSHIFT|KMOD_RSHIFT)
	KMOD_ALT      = KMOD_LALT | KMOD_RALT     // (KMOD_LALT|KMOD_RALT)
	KMOD_GUI      = KMOD_LGUI | KMOD_RGUI     // (KMOD_LGUI|KMOD_RGUI)
	KMOD_RESERVED = 0x8000                    // reserved for future use
)

// An enumeration of the predefined log categories.
// (https://wiki.libsdl.org/SDL_LOG_CATEGORY)
const (
	LOG_CATEGORY_APPLICATION = iota // application log
	LOG_CATEGORY_ERROR              // error log
	LOG_CATEGORY_ASSERT             // assert log
	LOG_CATEGORY_SYSTEM             // system log
	LOG_CATEGORY_AUDIO              // audio log
	LOG_CATEGORY_VIDEO              // video log
	LOG_CATEGORY_RENDER             // render log
	LOG_CATEGORY_INPUT              // input log
	LOG_CATEGORY_TEST               // test log
	LOG_CATEGORY_RESERVED1          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED2          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED3          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED4          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED5          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED6          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED7          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED8          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED9          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED10         // reserved for future SDL library use
	LOG_CATEGORY_CUSTOM             // reserved for application use
)

// An enumeration of the predefined log priorities.
// (https://wiki.libsdl.org/SDL_LogPriority)
const (
	LOG_PRIORITY_VERBOSE  = iota + 1 // verbose
	LOG_PRIORITY_DEBUG               // debug
	LOG_PRIORITY_INFO                // info
	LOG_PRIORITY_WARN                // warn
	LOG_PRIORITY_ERROR               // error
	LOG_PRIORITY_CRITICAL            // critical
	NUM_LOG_PRIORITIES               // (internal use)
)

// Cursor types for CreateSystemCursor()
const (
	SYSTEM_CURSOR_ARROW     = iota // arrow
	SYSTEM_CURSOR_IBEAM            // i-beam
	SYSTEM_CURSOR_WAIT             // wait
	SYSTEM_CURSOR_CROSSHAIR        // crosshair
	SYSTEM_CURSOR_WAITARROW        // small wait cursor (or wait if not available)
	SYSTEM_CURSOR_SIZENWSE         // double arrow pointing northwest and southeast
	SYSTEM_CURSOR_SIZENESW         // double arrow pointing northeast and southwest
	SYSTEM_CURSOR_SIZEWE           // double arrow pointing west and east
	SYSTEM_CURSOR_SIZENS           // double arrow pointing north and south
	SYSTEM_CURSOR_SIZEALL          // four pointed arrow pointing north, south, east, and west
	SYSTEM_CURSOR_NO               // slashed circle or crossbones
	SYSTEM_CURSOR_HAND             // hand
	NUM_SYSTEM_CURSORS             // (only for bounding internal arrays)
)

// Scroll direction types for the Scroll event
const (
	MOUSEWHEEL_NORMAL  = iota // the scroll direction is normal
	MOUSEWHEEL_FLIPPED        // the scroll direction is flipped / natural
)

// Used as a mask when testing buttons in buttonstate.
const (
	BUTTON_LEFT   = 1 // left mouse button
	BUTTON_MIDDLE = 2 // middle mouse button
	BUTTON_RIGHT  = 3 // right mouse button
	BUTTON_X1     = 4 // x1 mouse button
	BUTTON_X2     = 5 // x2 mouse button
)

// Pixel types.
const (
	PIXELTYPE_UNKNOWN = iota
	PIXELTYPE_INDEX1
	PIXELTYPE_INDEX4
	PIXELTYPE_INDEX8
	PIXELTYPE_PACKED8
	PIXELTYPE_PACKED16
	PIXELTYPE_PACKED32
	PIXELTYPE_ARRAYU8
	PIXELTYPE_ARRAYU16
	PIXELTYPE_ARRAYU32
	PIXELTYPE_ARRAYF16
	PIXELTYPE_ARRAYF32
)

// Bitmap pixel order high bit -> low bit.
const (
	BITMAPORDER_NONE = iota
	BITMAPORDER_4321
	BITMAPORDER_1234
)

// Packed component order high bit -> low bit.
const (
	PACKEDORDER_NONE = iota
	PACKEDORDER_XRGB
	PACKEDORDER_RGBX
	PACKEDORDER_ARGB
	PACKEDORDER_RGBA
	PACKEDORDER_XBGR
	PACKEDORDER_BGRX
	PACKEDORDER_ABGR
	PACKEDORDER_BGRA
)

// Array component order low byte -> high byte.
const (
	ARRAYORDER_NONE = iota
	ARRAYORDER_RGB
	ARRAYORDER_RGBA
	ARRAYORDER_ARGB
	ARRAYORDER_BGR
	ARRAYORDER_BGRA
	ARRAYORDER_ABGR
)

// Packed component layout.
const (
	PACKEDLAYOUT_NONE = iota
	PACKEDLAYOUT_332
	PACKEDLAYOUT_4444
	PACKEDLAYOUT_1555
	PACKEDLAYOUT_5551
	PACKEDLAYOUT_565
	PACKEDLAYOUT_8888
	PACKEDLAYOUT_2101010
	PACKEDLAYOUT_1010102
)

// Pixel format values.
const (
	PIXELFORMAT_UNKNOWN      = 0
	PIXELFORMAT_INDEX1LSB    = (1 << 28) | ((PIXELTYPE_INDEX1) << 24) | ((BITMAPORDER_4321) << 20) | ((0) << 16) | ((1) << 8) | ((0) << 0)
	PIXELFORMAT_INDEX1MSB    = (1 << 28) | ((PIXELTYPE_INDEX1) << 24) | ((BITMAPORDER_1234) << 20) | ((0) << 16) | ((1) << 8) | ((0) << 0)
	PIXELFORMAT_INDEX4LSB    = (1 << 28) | ((PIXELTYPE_INDEX4) << 24) | ((BITMAPORDER_4321) << 20) | ((0) << 16) | ((4) << 8) | ((0) << 0)
	PIXELFORMAT_INDEX4MSB    = (1 << 28) | ((PIXELTYPE_INDEX4) << 24) | ((BITMAPORDER_1234) << 20) | ((0) << 16) | ((4) << 8) | ((0) << 0)
	PIXELFORMAT_INDEX8       = (1 << 28) | ((PIXELTYPE_INDEX8) << 24) | ((0) << 20) | ((0) << 16) | ((8) << 8) | ((1) << 0)
	PIXELFORMAT_RGB332       = (1 << 28) | ((PIXELTYPE_PACKED8) << 24) | ((PACKEDORDER_XRGB) << 20) | ((PACKEDLAYOUT_332) << 16) | ((8) << 8) | ((1) << 0)
	PIXELFORMAT_RGB444       = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_XRGB) << 20) | ((PACKEDLAYOUT_4444) << 16) | ((12) << 8) | ((2) << 0)
	PIXELFORMAT_RGB555       = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_XRGB) << 20) | ((PACKEDLAYOUT_1555) << 16) | ((15) << 8) | ((2) << 0)
	PIXELFORMAT_BGR555       = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_XBGR) << 20) | ((PACKEDLAYOUT_1555) << 16) | ((15) << 8) | ((2) << 0)
	PIXELFORMAT_ARGB4444     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_ARGB) << 20) | ((PACKEDLAYOUT_4444) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_RGBA4444     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_RGBA) << 20) | ((PACKEDLAYOUT_4444) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_ABGR4444     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_ABGR) << 20) | ((PACKEDLAYOUT_4444) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_BGRA4444     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_BGRA) << 20) | ((PACKEDLAYOUT_4444) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_ARGB1555     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_ARGB) << 20) | ((PACKEDLAYOUT_1555) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_RGBA5551     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_RGBA) << 20) | ((PACKEDLAYOUT_5551) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_ABGR1555     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_ABGR) << 20) | ((PACKEDLAYOUT_1555) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_BGRA5551     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_BGRA) << 20) | ((PACKEDLAYOUT_5551) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_RGB565       = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_XRGB) << 20) | ((PACKEDLAYOUT_565) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_BGR565       = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_XBGR) << 20) | ((PACKEDLAYOUT_565) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_RGB24        = (1 << 28) | ((PIXELTYPE_ARRAYU8) << 24) | ((ARRAYORDER_RGB) << 20) | ((0) << 16) | ((24) << 8) | ((3) << 0)
	PIXELFORMAT_BGR24        = (1 << 28) | ((PIXELTYPE_ARRAYU8) << 24) | ((ARRAYORDER_BGR) << 20) | ((0) << 16) | ((24) << 8) | ((3) << 0)
	PIXELFORMAT_RGB888       = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_XRGB) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((24) << 8) | ((4) << 0)
	PIXELFORMAT_RGBX8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_RGBX) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((24) << 8) | ((4) << 0)
	PIXELFORMAT_BGR888       = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_XBGR) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((24) << 8) | ((4) << 0)
	PIXELFORMAT_BGRX8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_BGRX) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((24) << 8) | ((4) << 0)
	PIXELFORMAT_ARGB8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_ARGB) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((32) << 8) | ((4) << 0)
	PIXELFORMAT_RGBA8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_RGBA) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((32) << 8) | ((4) << 0)
	PIXELFORMAT_ABGR8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_ABGR) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((32) << 8) | ((4) << 0)
	PIXELFORMAT_BGRA8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_BGRA) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((32) << 8) | ((4) << 0)
	PIXELFORMAT_ARGB2101010  = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_ARGB) << 20) | ((PACKEDLAYOUT_2101010) << 16) | ((32) << 8) | ((4) << 0)
	PIXELFORMAT_YV12         = ('Y' << 0) | ('V' << 8) | ('1' << 16) | ('2' << 24)
	PIXELFORMAT_IYUV         = ('I' << 0) | ('Y' << 8) | ('U' << 16) | ('V' << 24)
	PIXELFORMAT_YUY2         = ('Y' << 0) | ('U' << 8) | ('Y' << 16) | ('2' << 24)
	PIXELFORMAT_UYVY         = ('U' << 0) | ('Y' << 8) | ('V' << 16) | ('Y' << 24)
	PIXELFORMAT_YVYU         = ('Y' << 0) | ('V' << 8) | ('Y' << 16) | ('U' << 24)
	PIXELFORMAT_NV12         = ('N' << 0) | ('V' << 8) | ('1' << 16) | ('2' << 24)
	PIXELFORMAT_NV21         = ('N' << 0) | ('V' << 8) | ('2' << 16) | ('1' << 24)
	PIXELFORMAT_EXTERNAL_OES = ('O' << 0) | ('E' << 8) | ('S' << 16) | (' ' << 24)
)

// These define alpha as the opacity of a surface.
const (
	ALPHA_OPAQUE      = 255
	ALPHA_TRANSPARENT = 0
)

// An enumeration of the basic state of the system's power supply.
// (https://wiki.libsdl.org/SDL_PowerState)
const (
	POWERSTATE_UNKNOWN    = iota // cannot determine power status
	POWERSTATE_ON_BATTERY        // not plugged in, running on the battery
	POWERSTATE_NO_BATTERY        // plugged in, no battery available
	POWERSTATE_CHARGING          // plugged in, charging battery
	POWERSTATE_CHARGED           // plugged in, battery charged
)

// An enumeration of flags used when creating a rendering context.
// (https://wiki.libsdl.org/SDL_RendererFlags)
const (
	RENDERER_SOFTWARE      = 0x00000001 // the renderer is a software fallback
	RENDERER_ACCELERATED   = 0x00000002 // the renderer uses hardware acceleration
	RENDERER_PRESENTVSYNC  = 0x00000004 // present is synchronized with the refresh rate
	RENDERER_TARGETTEXTURE = 0x00000008 // the renderer supports rendering to texture
)

// An enumeration of texture access patterns..
// (https://wiki.libsdl.org/SDL_TextureAccess)
const (
	TEXTUREACCESS_STATIC    = iota // changes rarely, not lockable
	TEXTUREACCESS_STREAMING        // changes frequently, lockable
	TEXTUREACCESS_TARGET           // can be used as a render target
)

// An enumeration of the texture channel modulation used in Renderer.Copy().
// (https://wiki.libsdl.org/SDL_TextureModulate)
const (
	TEXTUREMODULATE_NONE  = 0x00000000 // no modulation
	TEXTUREMODULATE_COLOR = 0x00000001 // srcC = srcC * color
	TEXTUREMODULATE_ALPHA = 0x00000002 // srcA = srcA * alpha
)

// An enumeration of flags that can be used in the flip parameter for Renderer.CopyEx().
// (https://wiki.libsdl.org/SDL_RendererFlip)
const (
	FLIP_NONE       RendererFlip = 0x00000000 // do not flip
	FLIP_HORIZONTAL              = 0x00000001 // flip horizontally
	FLIP_VERTICAL                = 0x00000002 // flip vertically
)

// RWops types
const (
	RWOPS_UNKNOWN   = 0 // unknown stream type
	RWOPS_WINFILE   = 1 // win32 file
	RWOPS_STDFILE   = 2 // stdio file
	RWOPS_JNIFILE   = 3 // android asset
	RWOPS_MEMORY    = 4 // memory stream
	RWOPS_MEMORY_RO = 5 // read-only memory stream
)

// RWops seek from
const (
	RW_SEEK_SET = 0 // seek from the beginning of data
	RW_SEEK_CUR = 1 // seek relative to current read point
	RW_SEEK_END = 2 // seek relative to the end of data
)

// The SDL keyboard scancode representation.
// (https://wiki.libsdl.org/SDL_Scancode)
// (https://wiki.libsdl.org/SDLScancodeLookup)
const (
	SCANCODE_UNKNOWN = 0 // "" (no name, empty string)

	SCANCODE_A = 4  // "A"
	SCANCODE_B = 5  // "B"
	SCANCODE_C = 6  // "C"
	SCANCODE_D = 7  // "D"
	SCANCODE_E = 8  // "E"
	SCANCODE_F = 9  // "F"
	SCANCODE_G = 10 // "G"
	SCANCODE_H = 11 // "H"
	SCANCODE_I = 12 // "I"
	SCANCODE_J = 13 // "J"
	SCANCODE_K = 14 // "K"
	SCANCODE_L = 15 // "L"
	SCANCODE_M = 16 // "M"
	SCANCODE_N = 17 // "N"
	SCANCODE_O = 18 // "O"
	SCANCODE_P = 19 // "P"
	SCANCODE_Q = 20 // "Q"
	SCANCODE_R = 21 // "R"
	SCANCODE_S = 22 // "S"
	SCANCODE_T = 23 // "T"
	SCANCODE_U = 24 // "U"
	SCANCODE_V = 25 // "V"
	SCANCODE_W = 26 // "W"
	SCANCODE_X = 27 // "X"
	SCANCODE_Y = 28 // "Y"
	SCANCODE_Z = 29 // "Z"

	SCANCODE_1 = 30 // "1"
	SCANCODE_2 = 31 // "2"
	SCANCODE_3 = 32 // "3"
	SCANCODE_4 = 33 // "4"
	SCANCODE_5 = 34 // "5"
	SCANCODE_6 = 35 // "6"
	SCANCODE_7 = 36 // "7"
	SCANCODE_8 = 37 // "8"
	SCANCODE_9 = 38 // "9"
	SCANCODE_0 = 39 // "0"

	SCANCODE_RETURN    = 40 // "Return"
	SCANCODE_ESCAPE    = 41 // "Escape" (the Esc key)
	SCANCODE_BACKSPACE = 42 // "Backspace"
	SCANCODE_TAB       = 43 // "Tab" (the Tab key)
	SCANCODE_SPACE     = 44 // "Space" (the Space Bar key(s))

	SCANCODE_MINUS        = 45 // "-"
	SCANCODE_EQUALS       = 46 // "="
	SCANCODE_LEFTBRACKET  = 47 // "["
	SCANCODE_RIGHTBRACKET = 48 // "]"
	SCANCODE_BACKSLASH    = 49 // "\"
	SCANCODE_NONUSHASH    = 50 // "#" (ISO USB keyboards actually use this code instead of 49 for the same key, but all OSes I've seen treat the two codes identically. So, as an implementor, unless your keyboard generates both of those codes and your OS treats them differently, you should generate SDL_SCANCODE_BACKSLASH instead of this code. As a user, you should not rely on this code because SDL will never generate it with most (all?) keyboards.)
	SCANCODE_SEMICOLON    = 51 // ";"
	SCANCODE_APOSTROPHE   = 52 // "'"
	SCANCODE_GRAVE        = 53 // "`"
	SCANCODE_COMMA        = 54 // ","
	SCANCODE_PERIOD       = 55 // "."
	SCANCODE_SLASH        = 56 // "/"
	SCANCODE_CAPSLOCK     = 57 // "CapsLock"
	SCANCODE_F1           = 58 // "F1"
	SCANCODE_F2           = 59 // "F2"
	SCANCODE_F3           = 60 // "F3"
	SCANCODE_F4           = 61 // "F4"
	SCANCODE_F5           = 62 // "F5"
	SCANCODE_F6           = 63 // "F6"
	SCANCODE_F7           = 64 // "F7"
	SCANCODE_F8           = 65 // "F8"
	SCANCODE_F9           = 66 // "F9"
	SCANCODE_F10          = 67 // "F10"
	SCANCODE_F11          = 68 // "F11"
	SCANCODE_F12          = 69 // "F12"
	SCANCODE_PRINTSCREEN  = 70 // "PrintScreen"
	SCANCODE_SCROLLLOCK   = 71 // "ScrollLock"
	SCANCODE_PAUSE        = 72 // "Pause" (the Pause / Break key)
	SCANCODE_INSERT       = 73 // "Insert" (insert on PC, help on some Mac keyboards (but does send code 73, not 117))
	SCANCODE_HOME         = 74 // "Home"
	SCANCODE_PAGEUP       = 75 // "PageUp"
	SCANCODE_DELETE       = 76 // "Delete"
	SCANCODE_END          = 77 // "End"
	SCANCODE_PAGEDOWN     = 78 // "PageDown"
	SCANCODE_RIGHT        = 79 // "Right" (the Right arrow key (navigation keypad))
	SCANCODE_LEFT         = 80 // "Left" (the Left arrow key (navigation keypad))
	SCANCODE_DOWN         = 81 // "Down" (the Down arrow key (navigation keypad))
	SCANCODE_UP           = 82 // "Up" (the Up arrow key (navigation keypad))

	SCANCODE_NUMLOCKCLEAR = 83 // "Numlock" (the Num Lock key (PC) / the Clear key (Mac))
	SCANCODE_KP_DIVIDE    = 84 // "Keypad /" (the / key (numeric keypad))
	SCANCODE_KP_MULTIPLY  = 85 // "Keypad *" (the * key (numeric keypad))
	SCANCODE_KP_MINUS     = 86 // "Keypad -" (the - key (numeric keypad))
	SCANCODE_KP_PLUS      = 87 // "Keypad +" (the + key (numeric keypad))
	SCANCODE_KP_ENTER     = 88 // "Keypad Enter" (the Enter key (numeric keypad))
	SCANCODE_KP_1         = 89 // "Keypad 1" (the 1 key (numeric keypad))
	SCANCODE_KP_2         = 90 // "Keypad 2" (the 2 key (numeric keypad))
	SCANCODE_KP_3         = 91 // "Keypad 3" (the 3 key (numeric keypad))
	SCANCODE_KP_4         = 92 // "Keypad 4" (the 4 key (numeric keypad))
	SCANCODE_KP_5         = 93 // "Keypad 5" (the 5 key (numeric keypad))
	SCANCODE_KP_6         = 94 // "Keypad 6" (the 6 key (numeric keypad))
	SCANCODE_KP_7         = 95 // "Keypad 7" (the 7 key (numeric keypad))
	SCANCODE_KP_8         = 96 // "Keypad 8" (the 8 key (numeric keypad))
	SCANCODE_KP_9         = 97 // "Keypad 9" (the 9 key (numeric keypad))
	SCANCODE_KP_0         = 98 // "Keypad 0" (the 0 key (numeric keypad))
	SCANCODE_KP_PERIOD    = 99 // "Keypad ." (the . key (numeric keypad))

	SCANCODE_NONUSBACKSLASH = 100 // "" (no name, empty string; This is the additional key that ISO keyboards have over ANSI ones, located between left shift and Y. Produces GRAVE ACCENT and TILDE in a US or UK Mac layout, REVERSE SOLIDUS (backslash) and VERTICAL LINE in a US or UK Windows layout, and LESS-THAN SIGN and GREATER-THAN SIGN in a Swiss German, German, or French layout.)
	SCANCODE_APPLICATION    = 101 // "Application" (the Application / Compose / Context Menu (Windows) key)
	SCANCODE_POWER          = 102 // "Power" (The USB document says this is a status flag, not a physical key - but some Mac keyboards do have a power key.)
	SCANCODE_KP_EQUALS      = 103 // "Keypad =" (the = key (numeric keypad))
	SCANCODE_F13            = 104 // "F13"
	SCANCODE_F14            = 105 // "F14"
	SCANCODE_F15            = 106 // "F15"
	SCANCODE_F16            = 107 // "F16"
	SCANCODE_F17            = 108 // "F17"
	SCANCODE_F18            = 109 // "F18"
	SCANCODE_F19            = 110 // "F19"
	SCANCODE_F20            = 111 // "F20"
	SCANCODE_F21            = 112 // "F21"
	SCANCODE_F22            = 113 // "F22"
	SCANCODE_F23            = 114 // "F23"
	SCANCODE_F24            = 115 // "F24"
	SCANCODE_EXECUTE        = 116 // "Execute"
	SCANCODE_HELP           = 117 // "Help"
	SCANCODE_MENU           = 118 // "Menu"
	SCANCODE_SELECT         = 119 // "Select"
	SCANCODE_STOP           = 120 // "Stop"
	SCANCODE_AGAIN          = 121 // "Again" (the Again key (Redo))
	SCANCODE_UNDO           = 122 // "Undo"
	SCANCODE_CUT            = 123 // "Cut"
	SCANCODE_COPY           = 124 // "Copy"
	SCANCODE_PASTE          = 125 // "Paste"
	SCANCODE_FIND           = 126 // "Find"
	SCANCODE_MUTE           = 127 // "Mute"
	SCANCODE_VOLUMEUP       = 128 // "VolumeUp"
	SCANCODE_VOLUMEDOWN     = 129 // "VolumeDown"
	SCANCODE_KP_COMMA       = 133 // "Keypad ," (the Comma key (numeric keypad))
	SCANCODE_KP_EQUALSAS400 = 134 // "Keypad = (AS400)" (the Equals AS400 key (numeric keypad))

	SCANCODE_INTERNATIONAL1 = 135 // "" (no name, empty string; used on Asian keyboards, see footnotes in USB doc)
	SCANCODE_INTERNATIONAL2 = 136 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL3 = 137 // "" (no name, empty string; Yen)
	SCANCODE_INTERNATIONAL4 = 138 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL5 = 139 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL6 = 140 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL7 = 141 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL8 = 142 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL9 = 143 // "" (no name, empty string)
	SCANCODE_LANG1          = 144 // "" (no name, empty string; Hangul/English toggle)
	SCANCODE_LANG2          = 145 // "" (no name, empty string; Hanja conversion)
	SCANCODE_LANG3          = 146 // "" (no name, empty string; Katakana)
	SCANCODE_LANG4          = 147 // "" (no name, empty string; Hiragana)
	SCANCODE_LANG5          = 148 // "" (no name, empty string; Zenkaku/Hankaku)
	SCANCODE_LANG6          = 149 // "" (no name, empty string; reserved)
	SCANCODE_LANG7          = 150 // "" (no name, empty string; reserved)
	SCANCODE_LANG8          = 151 // "" (no name, empty string; reserved)
	SCANCODE_LANG9          = 152 // "" (no name, empty string; reserved)

	SCANCODE_ALTERASE   = 153 // "AltErase" (Erase-Eaze)
	SCANCODE_SYSREQ     = 154 // "SysReq" (the SysReq key)
	SCANCODE_CANCEL     = 155 // "Cancel"
	SCANCODE_CLEAR      = 156 // "Clear"
	SCANCODE_PRIOR      = 157 // "Prior"
	SCANCODE_RETURN2    = 158 // "Return"
	SCANCODE_SEPARATOR  = 159 // "Separator"
	SCANCODE_OUT        = 160 // "Out"
	SCANCODE_OPER       = 161 // "Oper"
	SCANCODE_CLEARAGAIN = 162 // "Clear / Again"
	SCANCODE_CRSEL      = 163 // "CrSel"
	SCANCODE_EXSEL      = 164 // "ExSel"

	SCANCODE_KP_00              = 176 // "Keypad 00" (the 00 key (numeric keypad))
	SCANCODE_KP_000             = 177 // "Keypad 000" (the 000 key (numeric keypad))
	SCANCODE_THOUSANDSSEPARATOR = 178 // "ThousandsSeparator" (the Thousands Separator key)
	SCANCODE_DECIMALSEPARATOR   = 179 // "DecimalSeparator" (the Decimal Separator key)
	SCANCODE_CURRENCYUNIT       = 180 // "CurrencyUnit" (the Currency Unit key)
	SCANCODE_CURRENCYSUBUNIT    = 181 // "CurrencySubUnit" (the Currency Subunit key)
	SCANCODE_KP_LEFTPAREN       = 182 // "Keypad (" (the Left Parenthesis key (numeric keypad))
	SCANCODE_KP_RIGHTPAREN      = 183 // "Keypad )" (the Right Parenthesis key (numeric keypad))
	SCANCODE_KP_LEFTBRACE       = 184 // "Keypad {" (the Left Brace key (numeric keypad))
	SCANCODE_KP_RIGHTBRACE      = 185 // "Keypad }" (the Right Brace key (numeric keypad))
	SCANCODE_KP_TAB             = 186 // "Keypad Tab" (the Tab key (numeric keypad))
	SCANCODE_KP_BACKSPACE       = 187 // "Keypad Backspace" (the Backspace key (numeric keypad))
	SCANCODE_KP_A               = 188 // "Keypad A" (the A key (numeric keypad))
	SCANCODE_KP_B               = 189 // "Keypad B" (the B key (numeric keypad))
	SCANCODE_KP_C               = 190 // "Keypad C" (the C key (numeric keypad))
	SCANCODE_KP_D               = 191 // "Keypad D" (the D key (numeric keypad))
	SCANCODE_KP_E               = 192 // "Keypad E" (the E key (numeric keypad))
	SCANCODE_KP_F               = 193 // "Keypad F" (the F key (numeric keypad))
	SCANCODE_KP_XOR             = 194 // "Keypad XOR" (the XOR key (numeric keypad))
	SCANCODE_KP_POWER           = 195 // "Keypad ^" (the Power key (numeric keypad))
	SCANCODE_KP_PERCENT         = 196 // "Keypad %" (the Percent key (numeric keypad))
	SCANCODE_KP_LESS            = 197 // "Keypad <" (the Less key (numeric keypad))
	SCANCODE_KP_GREATER         = 198 // "Keypad >" (the Greater key (numeric keypad))
	SCANCODE_KP_AMPERSAND       = 199 // "Keypad &" (the & key (numeric keypad))
	SCANCODE_KP_DBLAMPERSAND    = 200 // "Keypad &&" (the && key (numeric keypad))
	SCANCODE_KP_VERTICALBAR     = 201 // "Keypad |" (the | key (numeric keypad))
	SCANCODE_KP_DBLVERTICALBAR  = 202 // "Keypad ||" (the || key (numeric keypad))
	SCANCODE_KP_COLON           = 203 // "Keypad :" (the : key (numeric keypad))
	SCANCODE_KP_HASH            = 204 // "Keypad #" (the # key (numeric keypad))
	SCANCODE_KP_SPACE           = 205 // "Keypad Space" (the Space key (numeric keypad))
	SCANCODE_KP_AT              = 206 // "Keypad @" (the @ key (numeric keypad))
	SCANCODE_KP_EXCLAM          = 207 // "Keypad !" (the ! key (numeric keypad))
	SCANCODE_KP_MEMSTORE        = 208 // "Keypad MemStore" (the Mem Store key (numeric keypad))
	SCANCODE_KP_MEMRECALL       = 209 // "Keypad MemRecall" (the Mem Recall key (numeric keypad))
	SCANCODE_KP_MEMCLEAR        = 210 // "Keypad MemClear" (the Mem Clear key (numeric keypad))
	SCANCODE_KP_MEMADD          = 211 // "Keypad MemAdd" (the Mem Add key (numeric keypad))
	SCANCODE_KP_MEMSUBTRACT     = 212 // "Keypad MemSubtract" (the Mem Subtract key (numeric keypad))
	SCANCODE_KP_MEMMULTIPLY     = 213 // "Keypad MemMultiply" (the Mem Multiply key (numeric keypad))
	SCANCODE_KP_MEMDIVIDE       = 214 // "Keypad MemDivide" (the Mem Divide key (numeric keypad))
	SCANCODE_KP_PLUSMINUS       = 215 // "Keypad +/-" (the +/- key (numeric keypad))
	SCANCODE_KP_CLEAR           = 216 // "Keypad Clear" (the Clear key (numeric keypad))
	SCANCODE_KP_CLEARENTRY      = 217 // "Keypad ClearEntry" (the Clear Entry key (numeric keypad))
	SCANCODE_KP_BINARY          = 218 // "Keypad Binary" (the Binary key (numeric keypad))
	SCANCODE_KP_OCTAL           = 219 // "Keypad Octal" (the Octal key (numeric keypad))
	SCANCODE_KP_DECIMAL         = 220 // "Keypad Decimal" (the Decimal key (numeric keypad))
	SCANCODE_KP_HEXADECIMAL     = 221 // "Keypad Hexadecimal" (the Hexadecimal key (numeric keypad))

	SCANCODE_LCTRL  = 224 // "Left Ctrl"
	SCANCODE_LSHIFT = 225 // "Left Shift"
	SCANCODE_LALT   = 226 // "Left Alt" (alt, option)
	SCANCODE_LGUI   = 227 // "Left GUI" (windows, command (apple), meta)
	SCANCODE_RCTRL  = 228 // "Right Ctrl"
	SCANCODE_RSHIFT = 229 // "Right Shift"
	SCANCODE_RALT   = 230 // "Right Alt" (alt gr, option)
	SCANCODE_RGUI   = 231 // "Right GUI" (windows, command (apple), meta)

	SCANCODE_MODE             = 257 // "ModeSwitch" (I'm not sure if this is really not covered by any of the above, but since there's a special KMOD_MODE for it I'm adding it here)
	SCANCODE_AUDIONEXT        = 258 // "AudioNext" (the Next Track media key)
	SCANCODE_AUDIOPREV        = 259 // "AudioPrev" (the Previous Track media key)
	SCANCODE_AUDIOSTOP        = 260 // "AudioStop" (the Stop media key)
	SCANCODE_AUDIOPLAY        = 261 // "AudioPlay" (the Play media key)
	SCANCODE_AUDIOMUTE        = 262 // "AudioMute" (the Mute volume key)
	SCANCODE_MEDIASELECT      = 263 // "MediaSelect" (the Media Select key)
	SCANCODE_WWW              = 264 // "WWW" (the WWW/World Wide Web key)
	SCANCODE_MAIL             = 265 // "Mail" (the Mail/eMail key)
	SCANCODE_CALCULATOR       = 266 // "Calculator" (the Calculator key)
	SCANCODE_COMPUTER         = 267 // "Computer" (the My Computer key)
	SCANCODE_AC_SEARCH        = 268 // "AC Search" (the Search key (application control keypad))
	SCANCODE_AC_HOME          = 269 // "AC Home" (the Home key (application control keypad))
	SCANCODE_AC_BACK          = 270 // "AC Back" (the Back key (application control keypad))
	SCANCODE_AC_FORWARD       = 271 // "AC Forward" (the Forward key (application control keypad))
	SCANCODE_AC_STOP          = 272 // "AC Stop" (the Stop key (application control keypad))
	SCANCODE_AC_REFRESH       = 273 // "AC Refresh" (the Refresh key (application control keypad))
	SCANCODE_AC_BOOKMARKS     = 274 // "AC Bookmarks" (the Bookmarks key (application control keypad))
	SCANCODE_BRIGHTNESSDOWN   = 275 // "BrightnessDown" (the Brightness Down key)
	SCANCODE_BRIGHTNESSUP     = 276 // "BrightnessUp" (the Brightness Up key)
	SCANCODE_DISPLAYSWITCH    = 277 // "DisplaySwitch" (display mirroring/dual display switch, video mode switch)
	SCANCODE_KBDILLUMTOGGLE   = 278 // "KBDIllumToggle" (the Keyboard Illumination Toggle key)
	SCANCODE_KBDILLUMDOWN     = 279 // "KBDIllumDown" (the Keyboard Illumination Down key)
	SCANCODE_KBDILLUMUP       = 280 // "KBDIllumUp" (the Keyboard Illumination Up key)
	SCANCODE_EJECT            = 281 // "Eject" (the Eject key)
	SCANCODE_SLEEP            = 282 // "Sleep" (the Sleep key)
	SCANCODE_APP1             = 283
	SCANCODE_APP2             = 284
	SCANCODE_AUDIOREWIND      = 285
	SCANCODE_AUDIOFASTFORWARD = 286
	NUM_SCANCODES             = 512
)

// These are the flags which may be passed to SDL_Init().
// (https://wiki.libsdl.org/SDL_Init)
const (
	INIT_TIMER          = 0x00000001 // timer subsystem
	INIT_AUDIO          = 0x00000010 // audio subsystem
	INIT_VIDEO          = 0x00000020 // video subsystem; automatically initializes the events subsystem
	INIT_JOYSTICK       = 0x00000200 // joystick subsystem; automatically initializes the events subsystem
	INIT_HAPTIC         = 0x00001000 // haptic (force feedback) subsystem
	INIT_GAMECONTROLLER = 0x00002000 // controller subsystem; automatically initializes the joystick subsystem
	INIT_EVENTS         = 0x00004000 // events subsystem
	INIT_NOPARACHUTE    = 0x00100000 // compatibility; this flag is ignored
	INIT_SENSOR         = 0x00008000 // sensor subsystem
	INIT_EVERYTHING     = INIT_TIMER | INIT_AUDIO | INIT_VIDEO | INIT_EVENTS |
		INIT_JOYSTICK | INIT_HAPTIC | INIT_GAMECONTROLLER |
		INIT_SENSOR // all of the above subsystems
)

const (
	RELEASED = 0
	PRESSED  = 1
)

// Surface flags (internal use)
const (
	SWSURFACE = 0          // just here for compatibility
	PREALLOC  = 0x00000001 // surface uses preallocated memory
	RLEACCEL  = 0x00000002 // surface is RLE encoded
	DONTFREE  = 0x00000004 // surface is referenced internally
)

// YUV Conversion Modes
const (
	YUV_CONVERSION_JPEG      YUV_CONVERSION_MODE = iota // Full range JPEG
	YUV_CONVERSION_BT601                                // BT.601 (the default)
	YUV_CONVERSION_BT709                                // BT.709
	YUV_CONVERSION_AUTOMATIC                            // BT.601 for SD content, BT.709 for HD content
)

// Various supported windowing subsystems.
const (
	SYSWM_UNKNOWN  = iota
	SYSWM_WINDOWS  // Microsoft Windows
	SYSWM_X11      // X Window System
	SYSWM_DIRECTFB // DirectFB
	SYSWM_COCOA    // Apple Mac OS X
	SYSWM_UIKIT    // Apple iOS
	SYSWM_WAYLAND  // Wayland (>= SDL 2.0.2)
	SYSWM_MIR      // Mir (>= SDL 2.0.2)
	SYSWM_WINRT    // WinRT (>= SDL 2.0.3)
	SYSWM_ANDROID  // Android (>= SDL 2.0.4)
	SYSWM_VIVANTE  // Vivante (>= SDL 2.0.5)
	SYSWM_OS2
)

// The version of SDL in use. NOTE that this is currently the version that this
// Go wrapper was created with. You should check your DLL's version using
// TODO(insert the right function name here).
const (
	MAJOR_VERSION = 2 // major version
	MINOR_VERSION = 0 // minor version
	PATCHLEVEL    = 9 // update version (patchlevel)
)

// An enumeration of window states.
// (https://wiki.libsdl.org/SDL_WindowFlags)
const (
	WINDOW_FULLSCREEN         = 0x00000001                     // fullscreen window
	WINDOW_OPENGL             = 0x00000002                     // window usable with OpenGL context
	WINDOW_SHOWN              = 0x00000004                     // window is visible
	WINDOW_HIDDEN             = 0x00000008                     // window is not visible
	WINDOW_BORDERLESS         = 0x00000010                     // no window decoration
	WINDOW_RESIZABLE          = 0x00000020                     // window can be resized
	WINDOW_MINIMIZED          = 0x00000040                     // window is minimized
	WINDOW_MAXIMIZED          = 0x00000080                     // window is maximized
	WINDOW_INPUT_GRABBED      = 0x00000100                     // window has grabbed input focus
	WINDOW_INPUT_FOCUS        = 0x00000200                     // window has input focus
	WINDOW_MOUSE_FOCUS        = 0x00000400                     // window has mouse focus
	WINDOW_FULLSCREEN_DESKTOP = WINDOW_FULLSCREEN | 0x00001000 // fullscreen window at the current desktop resolution
	WINDOW_FOREIGN            = 0x00000800                     // window not created by SDL
	WINDOW_ALLOW_HIGHDPI      = 0x00002000                     // window should be created in high-DPI mode if supported (>= SDL 2.0.1)
	WINDOW_MOUSE_CAPTURE      = 0x00004000                     // window has mouse captured (unrelated to INPUT_GRABBED, >= SDL 2.0.4)
	WINDOW_ALWAYS_ON_TOP      = 0x00008000                     // window should always be above others (X11 only, >= SDL 2.0.5)
	WINDOW_SKIP_TASKBAR       = 0x00010000                     // window should not be added to the taskbar (X11 only, >= SDL 2.0.5)
	WINDOW_UTILITY            = 0x00020000                     // window should be treated as a utility window (X11 only, >= SDL 2.0.5)
	WINDOW_TOOLTIP            = 0x00040000                     // window should be treated as a tooltip (X11 only, >= SDL 2.0.5)
	WINDOW_POPUP_MENU         = 0x00080000                     // window should be treated as a popup menu (X11 only, >= SDL 2.0.5)
	WINDOW_VULKAN             = 0x10000000                     // window usable for Vulkan surface (>= SDL 2.0.6)
)

// An enumeration of window events.
// (https://wiki.libsdl.org/SDL_WindowEventID)
const (
	WINDOWEVENT_NONE         = iota // (never used)
	WINDOWEVENT_SHOWN               // window has been shown
	WINDOWEVENT_HIDDEN              // window has been hidden
	WINDOWEVENT_EXPOSED             // window has been exposed and should be redrawn
	WINDOWEVENT_MOVED               // window has been moved to data1, data2
	WINDOWEVENT_RESIZED             // window has been resized to data1xdata2; this event is always preceded by WINDOWEVENT_SIZE_CHANGED
	WINDOWEVENT_SIZE_CHANGED        // window size has changed, either as a result of an API call or through the system or user changing the window size; this event is followed by WINDOWEVENT_RESIZED if the size was changed by an external event, i.e. the user or the window manager
	WINDOWEVENT_MINIMIZED           // window has been minimized
	WINDOWEVENT_MAXIMIZED           // window has been maximized
	WINDOWEVENT_RESTORED            // window has been restored to normal size and position
	WINDOWEVENT_ENTER               // window has gained mouse focus
	WINDOWEVENT_LEAVE               // window has lost mouse focus
	WINDOWEVENT_FOCUS_GAINED        // window has gained keyboard focus
	WINDOWEVENT_FOCUS_LOST          // window has lost keyboard focus
	WINDOWEVENT_CLOSE               // the window manager requests that the window be closed
	WINDOWEVENT_TAKE_FOCUS          // window is being offered a focus (should SDL_SetWindowInputFocus() on itself or a subwindow, or ignore) (>= SDL 2.0.5)
	WINDOWEVENT_HIT_TEST            // window had a hit test that wasn't SDL_HITTEST_NORMAL (>= SDL 2.0.5)
)

// Window position flags.
// (https://wiki.libsdl.org/SDL_CreateWindow)
const (
	WINDOWPOS_UNDEFINED_MASK = 0x1FFF0000 // used to indicate that you don't care what the window position is
	WINDOWPOS_UNDEFINED      = 0x1FFF0000 // used to indicate that you don't care what the window position is
	WINDOWPOS_CENTERED_MASK  = 0x2FFF0000 // used to indicate that the window position should be centered
	WINDOWPOS_CENTERED       = 0x2FFF0000 // used to indicate that the window position should be centered
)

// An enumeration of message box flags (e.g. if supported message box will display warning icon).
// (https://wiki.libsdl.org/SDL_MessageBoxFlags)
const (
	MESSAGEBOX_ERROR       = 0x00000010 // error dialog
	MESSAGEBOX_WARNING     = 0x00000020 // warning dialog
	MESSAGEBOX_INFORMATION = 0x00000040 // informational dialog
)

// Flags for MessageBoxButtonData.
const (
	MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT = 0x00000001 // marks the default button when return is hit
	MESSAGEBOX_BUTTON_ESCAPEKEY_DEFAULT = 0x00000002 // marks the default button when escape is hit
)

// OpenGL configuration attributes.
// (https://wiki.libsdl.org/SDL_GL_SetAttribute)
const (
	GL_RED_SIZE                   = iota // the minimum number of bits for the red channel of the color buffer; defaults to 3
	GL_GREEN_SIZE                        // the minimum number of bits for the green channel of the color buffer; defaults to 3
	GL_BLUE_SIZE                         // the minimum number of bits for the blue channel of the color buffer; defaults to 2
	GL_ALPHA_SIZE                        // the minimum number of bits for the alpha channel of the color buffer; defaults to 0
	GL_BUFFER_SIZE                       // the minimum number of bits for frame buffer size; defaults to 0
	GL_DOUBLEBUFFER                      // whether the output is single or double buffered; defaults to double buffering on
	GL_DEPTH_SIZE                        // the minimum number of bits in the depth buffer; defaults to 16
	GL_STENCIL_SIZE                      // the minimum number of bits in the stencil buffer; defaults to 0
	GL_ACCUM_RED_SIZE                    // the minimum number of bits for the red channel of the accumulation buffer; defaults to 0
	GL_ACCUM_GREEN_SIZE                  // the minimum number of bits for the green channel of the accumulation buffer; defaults to 0
	GL_ACCUM_BLUE_SIZE                   // the minimum number of bits for the blue channel of the accumulation buffer; defaults to 0
	GL_ACCUM_ALPHA_SIZE                  // the minimum number of bits for the alpha channel of the accumulation buffer; defaults to 0
	GL_STEREO                            // whether the output is stereo 3D; defaults to off
	GL_MULTISAMPLEBUFFERS                // the number of buffers used for multisample anti-aliasing; defaults to 0; see Remarks for details
	GL_MULTISAMPLESAMPLES                // the number of samples used around the current pixel used for multisample anti-aliasing; defaults to 0; see Remarks for details
	GL_ACCELERATED_VISUAL                // set to 1 to require hardware acceleration, set to 0 to force software rendering; defaults to allow either
	GL_RETAINED_BACKING                  // not used (deprecated)
	GL_CONTEXT_MAJOR_VERSION             // OpenGL context major version
	GL_CONTEXT_MINOR_VERSION             // OpenGL context minor version
	GL_CONTEXT_EGL                       // not used (deprecated)
	GL_CONTEXT_FLAGS                     // some combination of 0 or more of elements of the GLcontextFlag enumeration; defaults to 0 (https://wiki.libsdl.org/SDL_GLcontextFlag)
	GL_CONTEXT_PROFILE_MASK              // type of GL context (Core, Compatibility, ES); default value depends on platform (https://wiki.libsdl.org/SDL_GLprofile)
	GL_SHARE_WITH_CURRENT_CONTEXT        // OpenGL context sharing; defaults to 0
	GL_FRAMEBUFFER_SRGB_CAPABLE          // requests sRGB capable visual; defaults to 0 (>= SDL 2.0.1)
	GL_CONTEXT_RELEASE_BEHAVIOR          // sets context the release behavior; defaults to 1 (>= SDL 2.0.4)
	GL_CONTEXT_RESET_NOTIFICATION        // (>= SDL 2.0.6)
	GL_CONTEXT_NO_ERROR                  // (>= SDL 2.0.6)
)

// An enumeration of OpenGL profiles.
// (https://wiki.libsdl.org/SDL_GLprofile)
const (
	GL_CONTEXT_PROFILE_CORE          = 0x0001 // OpenGL core profile - deprecated functions are disabled
	GL_CONTEXT_PROFILE_COMPATIBILITY = 0x0002 // OpenGL compatibility profile - deprecated functions are allowed
	GL_CONTEXT_PROFILE_ES            = 0x0004 // OpenGL ES profile - only a subset of the base OpenGL functionality is available
)

// An enumeration of OpenGL context configuration flags.
// (https://wiki.libsdl.org/SDL_GLcontextFlag)
const (
	GL_CONTEXT_DEBUG_FLAG              = 0x0001 // intended to put the GL into a "debug" mode which might offer better developer insights, possibly at a loss of performance
	GL_CONTEXT_FORWARD_COMPATIBLE_FLAG = 0x0002 // intended to put the GL into a "forward compatible" mode, which means that no deprecated functionality will be supported, possibly at a gain in performance, and only applies to GL 3.0 and later contexts
	GL_CONTEXT_ROBUST_ACCESS_FLAG      = 0x0004 // intended to require a GL context that supports the GL_ARB_robustness extension--a mode that offers a few APIs that are safer than the usual defaults (think snprintf() vs sprintf())
	GL_CONTEXT_RESET_ISOLATION_FLAG    = 0x0008 // intended to require the GL to make promises about what to do in the face of driver or hardware failure
)

// CACHELINE_SIZE is a cacheline size used for padding.
const CACHELINE_SIZE = 128

const K_SCANCODE_MASK = 1 << 30

// MIX_MAXVOLUME is the full audio volume value used in MixAudioFormat() and AudioFormat().
// (https://wiki.libsdl.org/SDL_MixAudioFormat)
const MIX_MAXVOLUME = 128 // full audio volume

// TOUCH_MOUSEID is the device ID for mouse events simulated with touch input
const TOUCH_MOUSEID = 0xFFFFFFFF // uint32(-1)

const (
	PIXELFORMAT_RGBA32 = PIXELFORMAT_ABGR8888
	PIXELFORMAT_ARGB32 = PIXELFORMAT_BGRA8888
	PIXELFORMAT_BGRA32 = PIXELFORMAT_ARGB8888
	PIXELFORMAT_ABGR32 = PIXELFORMAT_RGBA8888
)

const SDL_STANDARD_GRAVITY = 9.80665

var ErrInvalidParameters = errors.New("Invalid Parameters")
