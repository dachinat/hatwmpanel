#include "panel.h"

#include <cairo/cairo.h>
#include <errno.h>
#include <fcntl.h>
#include <math.h>
#include <pango/pangocairo.h>
#include <librsvg/rsvg.h>
#include <gtk/gtk.h>
#include <gtk-layer-shell.h>
#include <libdbusmenu-gtk/menu.h>
#include <poll.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/syscall.h>
#include <unistd.h>
#include <wayland-client.h>

#include "wlr-layer-shell-unstable-v1-client-protocol.h"

// wlr-layer-shell's generated protocol metadata references xdg_popup for the
// optional get_popup request. HatWM Panel never issues that request, so a
// minimal interface descriptor is sufficient and avoids pulling in xdg-shell.
const struct wl_interface xdg_popup_interface = {
    "xdg_popup", 6, 0, NULL, 0, NULL
};

typedef struct PanelBuffer {
    struct wl_buffer *buffer;
    void *data;
    size_t size;
} PanelBuffer;

typedef struct PanelButton {
    int group;
    char *text;
    cairo_surface_t *icon;
    int origin_x;
    int width;
    int order;
} PanelButton;

typedef struct PanelSeparator {
    int group;
    int width;
    int order;
} PanelSeparator;

typedef struct PanelNetwork {
    char *ssid;
    int signal;
    int secured;
    int active;
    int wired;
} PanelNetwork;

typedef struct PanelBluetoothDevice {
    char *address;
    char *name;
    int connected;
} PanelBluetoothDevice;

enum PanelComponentKind {
    PANEL_COMPONENT_LAUNCHER = 0,
    PANEL_COMPONENT_WORKSPACES = 1,
    PANEL_COMPONENT_TRAY = 2,
    PANEL_COMPONENT_CLOCK = 3,
    PANEL_COMPONENT_VOLUME = 4,
    PANEL_COMPONENT_BRIGHTNESS = 5,
    PANEL_COMPONENT_MICROPHONE = 6,
    PANEL_COMPONENT_NETWORK = 7,
    PANEL_COMPONENT_BLUETOOTH = 8,
    PANEL_COMPONENT_BATTERY = 9,
    PANEL_COMPONENT_KEYBOARD_LAYOUT = 10,
    PANEL_COMPONENT_STORAGE = 11,
    PANEL_COMPONENT_RAM = 12,
    PANEL_COMPONENT_CPU = 13,
    PANEL_COMPONENT_GPU = 14,
    PANEL_COMPONENT_NETSTAT = 15,
    PANEL_COMPONENT_SEPARATORS = 16,
    PANEL_COMPONENT_BUTTONS = 17,
    PANEL_COMPONENT_TEXT = 18,
    PANEL_COMPONENT_COUNT = 19,
};

struct HatWMPanel {
    struct wl_display *display;
    struct wl_registry *registry;
    struct wl_compositor *compositor;
    struct wl_shm *shm;
    struct wl_seat *seat;
    struct wl_pointer *pointer;
    struct zwlr_layer_shell_v1 *layer_shell;
    struct wl_surface *surface;
    struct zwlr_layer_surface_v1 *layer_surface;

    int configured;
    int closed;
    int width;
    int height;
    int requested_height;
    int padding;
    uint32_t background_color;
    uint32_t foreground_color;
    double panel_opacity;
    int border_width;
    uint32_t border_color;
    int shadow_size;
    uint32_t shadow_color;
    char *font;
    char *separator;
    char *left;
    char *center;
    char *right;
    char *clock;
    cairo_surface_t *clock_icon;
    int clock_group;
    int clock_origin_x;
    int clock_width;
    int pending_clock_click;
    cairo_surface_t *volume_icon;
    int volume_group;
    int volume_origin_x;
    int volume_width;
    int volume_slider_origin_x;
    int volume_slider_width;
    int volume_percent;
    int volume_muted;
    int volume_step;
    uint32_t volume_color;
    int volume_dragging;
    int pending_volume_change;
    cairo_surface_t *brightness_icon;
    int brightness_group;
    int brightness_origin_x;
    int brightness_width;
    int brightness_slider_origin_x;
    int brightness_slider_width;
    int brightness_percent;
    int brightness_step;
    int brightness_dragging;
    int pending_brightness_change;
    uint32_t brightness_color;
    cairo_surface_t *microphone_icon;
    int microphone_group;
    int microphone_origin_x;
    int microphone_width;
    int microphone_slider_origin_x;
    int microphone_slider_width;
    int microphone_percent;
    int microphone_muted;
    int microphone_step;
    int microphone_dragging;
    int pending_microphone_change;
    uint32_t microphone_color;
    cairo_surface_t *network_icon;
    cairo_surface_t *network_wired_icon;
    int network_active_wired;
    int network_group;
    int network_origin_x;
    int network_width;
    int network_show_text;
    char *network_label;
    PanelNetwork *networks;
    int network_count;
    int pending_network_menu;
    int pending_network_click;
    int pending_network_disconnect;
    GtkWidget *network_menu;
    cairo_surface_t *bluetooth_icon;
    int bluetooth_group;
    int bluetooth_origin_x;
    int bluetooth_width;
    int bluetooth_show_text;
    char *bluetooth_label;
    PanelBluetoothDevice *bluetooth_devices;
    int bluetooth_device_count;
    int pending_bluetooth_menu;
    int pending_bluetooth_click;
    int pending_bluetooth_disconnect;
    GtkWidget *bluetooth_menu;
    cairo_surface_t *battery_icon;
    char *battery_icon_name;
    uint32_t battery_icon_color;
    int battery_icon_tint;
    int battery_dynamic_icon;
    int battery_group;
    int battery_origin_x;
    int battery_width;
    int battery_show_text;
    char battery_label[16];
    int battery_percent;
    int battery_charging;
    int battery_full;
    cairo_surface_t *keyboard_layout_icon;
    int keyboard_layout_group;
    int keyboard_layout_origin_x;
    int keyboard_layout_width;
    char *keyboard_layout;
    int pending_keyboard_layout_click;
    cairo_surface_t *storage_icon;
    int storage_group;
    int storage_origin_x;
    int storage_width;
    char *storage_text;
    cairo_surface_t *ram_icon;
    int ram_group;
    int ram_origin_x;
    int ram_width;
    char *ram_text;
    cairo_surface_t *cpu_icon;
    int cpu_group;
    int cpu_origin_x;
    int cpu_width;
    char *cpu_text;
    cairo_surface_t *gpu_icon;
    int gpu_group;
    int gpu_origin_x;
    int gpu_width;
    char *gpu_text;
    cairo_surface_t *netstat_icon;
    int netstat_group;
    int netstat_origin_x;
    int netstat_width;
    char *netstat_text;
    uint32_t module_background_color;
    uint32_t module_icon_color;
    int module_padding;
    int module_radius;
    int module_icon_size;
    GtkWidget *calendar_window;
    int calendar_left_margin;
    GtkWidget *launcher_window;
    GtkWidget *launcher_entry;
    GtkWidget *launcher_hint;
    GtkWidget *launcher_results;
    GtkWidget *launcher_list;
    GList *launcher_apps;
    int launcher_left_margin;
    PanelButton *buttons;
    int button_count;
    int pending_button_click;
    PanelSeparator *separators;
    int separator_count;
    int group_order[4][PANEL_COMPONENT_COUNT];

    int launcher_group;
    cairo_surface_t *launcher_logo;
    int launcher_origin_x;
    int launcher_width;
    int launcher_size;
    int pending_launcher_click;
    int workspace_group;
    int workspace_width;
    int workspace_height;
    int workspace_gap;
    int workspace_inset;
    uint32_t workspace_color;
    uint32_t workspace_active_color;
    uint32_t workspace_urgent_color;
    uint32_t workspace_border_color;
    uint32_t workspace_window_color;
    uint32_t workspace_focused_window_color;
    HatWMPanelWorkspace *workspaces;
    int workspace_count;
    HatWMPanelWindow *windows;
    int window_count;
    HatWMPanelOutput output;
    int workspace_origin_x;
    int workspace_origin_y;
    int workspace_draw_width;
    int workspace_draw_height;
    int pending_workspace_click;
    int tray_group;
    cairo_surface_t **tray_icons;
    int tray_count;
    int tray_origin_x;
    int tray_origin_y;
    int tray_item_size;
    int pending_tray_click;
    int pending_tray_button;
    double pointer_x;
    double pointer_y;
    int pointer_left_pressed;
    int panel_at_bottom;
};

static char last_error[512];
static int gtk_initialized;
static int active_tray_menus;
static GtkCssProvider *tray_menu_style;
static int panel_draw(HatWMPanel *panel);
static int clamp_int(int value, int minimum, int maximum);
