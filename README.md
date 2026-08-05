[![PayPal](https://img.shields.io/badge/PayPal-Support-0070BA?logo=paypal&logoColor=white)](https://paypal.me/dachina)

# hatwmpanel

`hatwmpanel` is a native Wayland panel written in Go for the HatWM compositor.
It uses `wlr-layer-shell-unstable-v1`, so HatWM can place the bar on the top or
bottom edge and reserve space for tiled windows.

## Features

- native Wayland layer-shell panel; no GTK window
- top or bottom placement
- configurable layer, height, margins, colors, font, padding, and exclusive zone
- left / center / right module groups
- `launcher`, `text`, interactive `clock` and `button`, periodically refreshed `exec`,
  HatWM `workspaces`, StatusNotifierItem `tray`, and interactive network modules
- GNOME 2/early GNOME-style workspace pager:
  - one small desktop square per workspace
  - miniature windows drawn at their actual relative positions
  - active workspace and focused window highlighting
  - click a square to switch workspace
- persistent direct connection to HatWM's Unix-socket IPC
- automatic IPC reconnect when HatWM restarts
- automatic config creation and reload
- config syntax follows HatWM's INI-like configuration style

## Dependencies

On Arch Linux:

```sh
sudo pacman -S --needed go wayland cairo pango librsvg gtk3 gtk-layer-shell \
  libdbusmenu-gtk3 networkmanager bluez-utils zenity pkgconf meson ninja
```

`wayland-scanner` is supplied by the Wayland package. The project includes the
`wlr-layer-shell-unstable-v1.xml` protocol description and generates the client
header/code in Meson's build directory during compilation.

## Required HatWM version

The workspace pager needs the geometry fields added to HatWM IPC protocol v1:

- `get_state.result.output`
- `get_windows.result[].x`
- `get_windows.result[].y`
- `get_windows.result[].width`
- `get_windows.result[].height`

Use the updated HatWM source included with this delivery. Older HatWM builds can
still provide workspace counts, but cannot draw accurate miniature window
positions.

## Build

```sh
meson setup build
meson compile -C build
meson test -C build
./build/hatwmpanel
```

Meson uses Ninja as its default backend. Build files, generated Wayland
protocol sources, and the executable remain outside the source tree.

## Source layout

```text
internal/
  audio/                 volume and microphone backends
  connectivity/          NetworkManager, BlueZ, and traffic backends
  desktop/               desktop-service integrations such as the system tray
  hatwm/                 HatWM IPC and keyboard-layout support
  system/                battery, display, processor, memory, and disk readers
native/
  *.go                   rendering, module UI, and typed Go-to-C bridges
  logo.png               bundled 32 px launcher artwork
  panel.h                public C API
  panel.c                native translation-unit index
  panel_*.inc            focused Wayland/Cairo/GTK implementation units
config/
  *.go                   configuration model, parsing, and validation
config_bridge.go         root compatibility aliases for configuration
native_bridge.go         root compatibility aliases for the native panel
module_runner.go         static text and periodic command rendering
main.go                  application lifecycle and component coordination
```

The root Go package coordinates components. Private backend integrations are
grouped by responsibility under `internal`, configuration belongs under
`config`, and module rendering and desktop-library integration live together
under `native`.

Install system-wide:

```sh
sudo meson install -C build
```

The default prefix is `/usr/local`. Choose another prefix during initial setup,
for example `meson setup build --prefix=/usr`. Format or vet the Go sources
with `meson compile -C build fmt` and `meson compile -C build vet`. Remove files
installed by Meson with `sudo ninja -C build uninstall`.

Then start it from `~/.config/hatwm/config`:

```ini
[autostart]
panel = hatwmpanel
```

HatWM exports `HATWM_SOCKET` before starting autostart commands. If the panel is
started another way, it falls back to:

```text
$XDG_RUNTIME_DIR/hatwm/ipc.sock
```

The panel talks to the socket directly. It does not spawn `hatwmctl`, which keeps
workspace updates event-driven and avoids repeatedly creating processes.

## Configuration

On first launch, the panel creates:

```text
~/.config/hatwmpanel/config
```

Example:

The default **HatWM Midnight** palette uses deep ink backgrounds, steel-blue
surfaces, Hat blue for selection, and distinct semantic status accents.

```ini
[settings]
position = top
layer = top
height = 32
exclusive_zone = true
background_color = 0x10151dff
foreground_color = 0xe6edf6ff
panel_opacity = 1.0
border_width = 0
border_color = 0x34465fff
shadow_size = 0
shadow_color = 0x070a0f99
font = Iosevka Nerd Font SemiBold 10
padding = 10
separator = "  |  "

workspace_width = 24
workspace_height = 22
workspace_gap = 5
workspace_inset = 2
workspace_color = 0x202d3fff
workspace_active_color = 0x315b86ff
workspace_urgent_color = 0xf07178ff
workspace_border_color = 0x34465fff
workspace_window_color = 0x8b9bb0ff
workspace_focused_window_color = 0x67d4c0ff

# Tray context-menu appearance
menu_background_color = 0x10151dff
menu_foreground_color = 0xe6edf6ff
menu_selected_background_color = 0x315b86ff
menu_selected_foreground_color = 0xe6edf6ff
menu_border_color = 0x34465fff
menu_font = Iosevka Nerd Font 10
menu_padding = 6
menu_border_width = 1
menu_radius = 6
module_background_color = 0x182231ff
module_icon_color = 0x6aa9e9ff
module_padding = 7
module_radius = 6
module_icon_size = 16
volume_slider_width = 80
brightness_slider_width = 80
microphone_slider_width = 80

margin_top = 0
margin_right = 0
margin_bottom = 0
margin_left = 0

[left]
launcher = launcher
workspaces = workspaces
storage = storage path=/ icon=drive-harddisk-symbolic
ram = ram icon=computer-symbolic
cpu = cpu icon=applications-system-symbolic
gpu = gpu icon=video-display-symbolic
netstat = netstat icon=network-transmit-receive-symbolic
wm = text HatWM

[center]
host = exec 30 hostname

[right]
network = network icon=auto text=on
bluetooth = bluetooth icon=bluetooth-active-symbolic text=on
battery = battery icon=auto icon_color=0x7bd88fff text=on
keyboard = keyboard_layout icon=input-keyboard-symbolic
tray = tray
clock = clock icon=preferences-system-time-symbolic %a %d %b  %H:%M
power = button icon=system-shutdown-symbolic text=none action=wlogout
volume = volume icon=audio-volume-high-symbolic step=5
brightness = brightness icon=display-brightness-symbolic step=5
microphone = microphone icon=audio-input-microphone-symbolic step=5
```

`panel_opacity` controls the transparency of the complete panel, including
modules, and accepts values from `0.0` (invisible) to `1.0` (opaque).
`border_width` draws an outline inside the panel surface. `shadow_size` draws
a gradient along the desktop-facing edge, automatically using the bottom edge
for a top panel and the top edge for a bottom panel. Set either size to `0` to
disable that effect.

### Module syntax

```ini
name = launcher [icon=auto|system-icon-name] [icon_color=RGBA]
name = separator width=PIXELS
name = workspaces
name = tray
name = text arbitrary text
name = clock [icon=system-icon-name|none] [icon_color=RGBA] %a %d %b  %H:%M
name = button [icon=name|none] [icon_color=RGBA] [text=label|none] action=command
name = volume [icon=name|none] [icon_color=RGBA] [step=1..25]
name = brightness [icon=name|none] [icon_color=RGBA] [step=1..25]
name = microphone [icon=name|none] [icon_color=RGBA] [step=1..25]
name = network [icon=auto|name|none] [wireless_icon=name|none] [wired_icon=name|none] [icon_color=RGBA] [text=on|none]
name = bluetooth [icon=name|none] [icon_color=RGBA] [text=on|none]
name = battery [icon=auto|name|none] [icon_color=RGBA] [text=on|none]
name = keyboard_layout [icon=name|none] [icon_color=RGBA]
name = storage [path=/] [icon=name|none] [icon_color=RGBA]
name = ram [icon=name|none] [icon_color=RGBA]
name = cpu [icon=name|none] [icon_color=RGBA]
name = gpu [icon=name|none] [icon_color=RGBA]
name = netstat [icon=name|none] [icon_color=RGBA]
name = exec 5 command --with arguments
```

For every module and button, regular system icons preserve their original SVG
or raster artwork. Icons ending in `-symbolic` are tinted with
`module_icon_color`; an explicit `icon_color` forces tinting for any icon.

The network module switches icons automatically by connection type:

```ini
network = network icon=auto
# Or override either system icon:
network = network wireless_icon=network-wireless-symbolic wired_icon=network-wired-symbolic
```

For compatibility, `icon=name` customizes the wireless icon while retaining
the default wired icon. `icon=none` disables both icons; either can then be
enabled again with its specific option.

Without an `icon` option, the `launcher` module displays
`native/logo.png` exactly as authored, including its original colors and
transparency. Set a system icon name to replace the bundled PNG completely:

```ini
launcher = launcher icon=start-here-symbolic icon_color=0x78a9ffff
```

`icon=auto` explicitly selects the bundled PNG. The `icon_color` option applies
only when the launcher uses a system icon. Regular system SVG icons retain
their original artwork and colors; names ending in `-symbolic` use
`module_icon_color`, while an explicit `icon_color` tints any configured
system icon.

Clicking the module opens a search popup directly beneath it. Results update
while typing and use
application names and icons from installed desktop entries. Click a result to
close the popup and launch it.

```ini
[left]
launcher = launcher
workspaces = workspaces
```

The first matching application is selected automatically. Press Enter to
launch the selected result. When there is no matching application, Enter
executes the entered text through `/bin/sh -c`, including arguments and shell
syntax:

```text
kitty
```

Press Escape or click the launcher again to close the popup.

HatwmPanel toggles the launcher when it receives `SIGUSR1`. This allows the
compositor to provide a configurable global shortcut. For HatWM:

```ini
[keybindings]
Mod4+r = exec pkill -USR1 -x hatwmpanel
```

Change `Mod4+r` to any unused HatWM key combination. Sending the signal again
toggles the launcher closed.

The `separator` module reserves an exact number of horizontal pixels without
drawing anything. It can be used more than once and keeps its configured
position among other modules. Both forms below are accepted:

```ini
[left]
launcher = launcher
launcher_gap = separator width=12
workspaces = workspaces

# Short form:
# launcher_gap = separator 12
```

Only the first `workspaces` module is used. It can be placed in `[left]`,
`[center]`, or `[right]`. In that group, the pager is drawn before the regular
text modules.

Only the first `tray` module is used. It hosts modern
StatusNotifierItem/AppIndicator applications on the session D-Bus. Each item
is represented by a compact marker derived from its title, and a left-click
calls the item's `Activate` method. Legacy XEmbed-only tray icons are not
supported on Wayland.

Only the first `clock` module is used. It displays date/time text using the
formatting tokens below; click it to open a calendar below (or above, for a
bottom panel) the module.

Interactive modules have their own background and use monochrome symbolic
icons from the system icon theme. For the clock, omit `icon=...` to use
`preferences-system-time-symbolic`, choose another installed symbolic icon
name, or set `icon=none` to hide it. The shared appearance is configured with
`module_background_color`, `module_icon_color`, `module_padding`,
`module_radius`, and `module_icon_size`.

Set `icon_color=0xRRGGBBAA` on a clock or button to override
`module_icon_color` for that module only.

`button` modules are reusable launchers. Each button may show an icon, text, or
both, and `action=` contains the shell command launched on click. For example:

```ini
power = button icon=system-shutdown-symbolic icon_color=0xf07178ff text=none action=wlogout
lock = button icon=system-lock-screen-symbolic icon_color=0x78a9ffff text=Lock action=loginctl lock-session
terminal = button icon=none text=Terminal action=foot
```

Add any number of buttons to the left, center, or right group. Actions launch
asynchronously, so the panel remains responsive.

The `volume` module controls the default audio sink through `wpctl`, with a
`pactl` fallback. Click or drag its slider to set the level, or scroll anywhere
over the module to adjust by `step` percentage points. `volume_slider_width`
controls the track width:

```ini
volume = volume icon=audio-volume-high-symbolic icon_color=0x78a9ffff step=5
```

The `brightness` module uses `brightnessctl` and provides the same click,
drag, and scroll interactions. Brightness is clamped to 1–100%:

```ini
brightness_slider_width = 80
brightness = brightness icon=display-brightness-symbolic icon_color=0xeea846ff step=5
```

The `microphone` module controls the default audio input and mirrors the
speaker volume interactions. Adjusting it also unmutes the input:

```ini
microphone_slider_width = 80
microphone = microphone icon=audio-input-microphone-symbolic icon_color=0xbe95ffff step=5
```

The `network` module shows the active Wi-Fi SSID or Ethernet profile. Click it
to open nearby Wi-Fi networks and saved Ethernet profiles, then click an entry
to connect through NetworkManager. Saved and open Wi-Fi networks connect
directly; a new secured network opens a masked Zenity password prompt.
Ethernet profiles use the system wired-network icon while active. The dropdown
also provides a `Disconnect` button for the active connection. Set `text=none`
for an icon-only module:

```ini
network = network icon=auto icon_color=0x33b1ffff text=on
```

The `bluetooth` module shows the connected device name, or `Bluetooth` when
nothing is connected. Click it to open a centered list of BlueZ devices.
Connected devices have a checkmark; selecting another device connects it,
pairing and trusting it first when necessary. A `Disconnect` action is shown
while a device is connected:

```ini
bluetooth = bluetooth icon=bluetooth-active-symbolic icon_color=0x8cb6ffff text=on
```

The `battery` module reads `/sys/class/power_supply`, displays the current
percentage, and automatically chooses a system symbolic icon for the charge
level and charging state. Use a system icon name instead of `auto` for a fixed
icon, or `icon=none`/`text=none` to hide either part:

```ini
battery = battery icon=auto icon_color=0x7bd88fff text=on
```

The `keyboard_layout` module reads HatWM's current XKB layout over its
persistent IPC connection and displays a short language code such as `EN`,
`GE`, `FR`, or `RU`. Click the module to send HatWM's
`toggle_keyboard_layout` IPC command and advance to the next layout configured
by `keyboard_layouts` in HatWM:

```ini
keyboard = keyboard_layout icon=input-keyboard-symbolic icon_color=0x08bdbaff
```

The `storage` module shows used and available filesystem space in a compact
`used / available` form. It reads `/` by default and accepts another mount
point through `path=`:

```ini
[left]
workspaces = workspaces
storage = storage path=/ icon=drive-harddisk-symbolic
```

The `ram` module reads `/proc/meminfo` and displays used and available memory
with one decimal place:

```ini
[left]
workspaces = workspaces
storage = storage path=/ icon=drive-harddisk-symbolic
ram = ram icon=computer-symbolic
cpu = cpu icon=applications-system-symbolic
gpu = gpu icon=video-display-symbolic
netstat = netstat icon=network-transmit-receive-symbolic
```

The `cpu` module samples aggregate utilization from `/proc/stat` once per
second. The `gpu` module updates every five seconds, preferring the Linux DRM
`gpu_busy_percent` interface and falling back to `nvidia-smi`. It displays
`N/A` when the active GPU driver does not expose either source.

The `netstat` module reads `/proc/net/dev` once per second and displays
aggregate download and upload speeds for all non-loopback interfaces:

```ini
netstat = netstat icon=network-transmit-receive-symbolic
```

For `exec`, the optional number after `exec` is the refresh interval in seconds.
Without it, the interval is 5 seconds. Commands run through `sh -c` with a
two-second timeout.

Supported clock tokens include:

- `%a`, `%A` — abbreviated/full weekday
- `%b`, `%B` — abbreviated/full month
- `%d`, `%m`, `%Y`, `%y` — date
- `%H`, `%I`, `%M`, `%S`, `%p`, `%z` — time
- `%%` — literal percent sign

## Workspace behavior

The panel subscribes to workspace, window, focus, layout, fullscreen, and config
events. It also refreshes geometry periodically so floating move/resize changes
remain accurate even if a client changes its own size without a dedicated IPC
event.

When an application on another workspace requests activation, HatWM marks that
workspace urgent and the pager uses `workspace_urgent_color`. The highlight is
cleared when the requested window receives focus.

A left-click sends this IPC command directly to HatWM:

```json
{"type":"command","command":"workspace","workspace":2}
```

## Layout mode module

The `layout_mode` module reflects HatWM's current global layout and toggles it
through the persistent IPC connection when clicked:

```ini
layout = layout_mode tiling_icon=view-grid-symbolic floating_icon=window-restore-symbolic icon_color=0x6aa9e9ff
```

Both icons come from the current system icon theme. The options are optional;
the values above are the defaults. The icon also updates when the layout is
changed through a keybinding or another IPC client.
