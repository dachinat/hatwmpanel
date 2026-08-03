project(
    'hatwmpanel',
    version: '0.3.0',
    meson_version: '>=1.1.0',
)

go = find_program('go', required: true)
wayland_scanner = find_program('wayland-scanner', required: true)

wayland_client = dependency('wayland-client', required: true)
cairo = dependency('cairo', required: true)
pangocairo = dependency('pangocairo', required: true)
librsvg = dependency('librsvg-2.0', required: true)
gtk3 = dependency('gtk+-3.0', required: true)
gtk_layer_shell = dependency('gtk-layer-shell-0', required: true)
dbusmenu_gtk3 = dependency('dbusmenu-gtk3-0.4', required: true)

source_root = meson.project_source_root()
build_root = meson.project_build_root()
protocol_xml = files('protocols/wlr-layer-shell-unstable-v1.xml')

protocol_header = custom_target(
    'wlr-layer-shell-client-header',
    input: protocol_xml,
    output: 'wlr-layer-shell-unstable-v1-client-protocol.h',
    command: [wayland_scanner, 'client-header', '@INPUT@', '@OUTPUT@'],
)

protocol_code = custom_target(
    'wlr-layer-shell-private-code',
    input: protocol_xml,
    output: 'wlr-layer-shell-unstable-v1-protocol.c',
    command: [wayland_scanner, 'private-code', '@INPUT@', '@OUTPUT@'],
)

# Go owns the cgo compile and final link. Every non-test input is listed so
# Ninja notices changes to Go packages, embedded assets, native implementation
# fragments, and generated Wayland protocol sources. The forced Go rebuild is
# retained because panel.c includes the panel_*.inc implementation files.
hatwmpanel_sources = files(
    'go.mod',
    'go.sum',
    'config_bridge.go',
    'main.go',
    'module_helpers.go',
    'module_runner.go',
    'native_bridge.go',
    'internal/system/battery/battery.go',
    'internal/connectivity/bluetooth/bluetooth.go',
    'internal/system/brightness/brightness.go',
    'config/config.go',
    'config/module.go',
    'config/separator.go',
    'internal/system/cpu/cpu.go',
    'internal/system/gpu/gpu.go',
    'internal/hatwm/keyboardlayout/keyboardlayout.go',
    'internal/system/memory/memory.go',
    'internal/audio/microphone/microphone.go',
    'internal/connectivity/netstat/netstat.go',
    'internal/connectivity/network/network.go',
    'native/battery.go',
    'native/bluetooth.go',
    'native/brightness.go',
    'native/button_action.go',
    'native/buttons.go',
    'native/clock.go',
    'native/cpu_gpu.go',
    'native/keyboard_layout.go',
    'native/launcher.go',
    'native/layout.go',
    'native/logo.go',
    'native/logo.svg',
    'native/microphone.go',
    'native/native.go',
    'native/network.go',
    'native/ram.go',
    'native/separators.go',
    'native/storage.go',
    'native/tray.go',
    'native/volume.go',
    'native/workspaces.go',
    'internal/system/storage/storage.go',
    'internal/desktop/tray/tray.go',
    'internal/audio/volume/volume.go',
    'internal/hatwm/workspaces/client.go',
    'native/layer_shell_protocol.c',
    'native/panel.c',
    'native/panel.h',
    'native/panel_calendar.inc',
    'native/panel_connections_menu.inc',
    'native/panel_dispatch.inc',
    'native/panel_input.inc',
    'native/panel_internal.inc',
    'native/panel_launcher.inc',
    'native/panel_lifecycle.inc',
    'native/panel_render.inc',
    'native/panel_state.inc',
    'native/panel_style.inc',
    'native/panel_tray_menu.inc',
)

go_env = environment()
go_env.set('CGO_CFLAGS', '-I' + build_root)

hatwmpanel = custom_target(
    'hatwmpanel',
    input: hatwmpanel_sources + [protocol_header, protocol_code],
    output: 'hatwmpanel',
    command: [
        go,
        '-C',
        source_root,
        'build',
        '-a',
        '-buildvcs=false',
        '-o',
        join_paths(build_root, 'hatwmpanel'),
        '.',
    ],
    env: go_env,
    build_by_default: true,
    install: true,
    install_dir: get_option('bindir'),
    console: true,
)

test(
    'go test',
    go,
    args: ['-C', source_root, 'test', './...'],
    env: go_env,
    workdir: source_root,
    depends: [protocol_header, protocol_code],
    timeout: 180,
)

test(
    'go vet',
    go,
    args: ['-C', source_root, 'vet', './...'],
    env: go_env,
    workdir: source_root,
    depends: [protocol_header, protocol_code],
    timeout: 180,
)

run_target(
    'fmt',
    command: [go, '-C', source_root, 'fmt', './...'],
)

run_target(
    'vet',
    command: [go, '-C', source_root, 'vet', './...'],
    env: go_env,
    depends: [protocol_header, protocol_code],
)

summary(
    {
        'Go': go.full_path(),
        'Wayland scanner': wayland_scanner.full_path(),
        'Wayland client': wayland_client.version(),
        'Cairo': cairo.version(),
        'Pango Cairo': pangocairo.version(),
        'librsvg': librsvg.version(),
        'GTK 3': gtk3.version(),
        'GTK Layer Shell': gtk_layer_shell.version(),
        'DBusMenu GTK 3': dbusmenu_gtk3.version(),
        'Install directory': join_paths(
            get_option('prefix'),
            get_option('bindir'),
        ),
    },
    section: 'hatwmpanel',
)
