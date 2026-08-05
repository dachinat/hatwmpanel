#ifndef HATWM_PANEL_H
#define HATWM_PANEL_H

#include <stdint.h>

typedef struct HatWMPanel HatWMPanel;

enum HatWMPanelGroup {
    HATWM_PANEL_GROUP_NONE = 0,
    HATWM_PANEL_GROUP_LEFT = 1,
    HATWM_PANEL_GROUP_CENTER = 2,
    HATWM_PANEL_GROUP_RIGHT = 3,
};

typedef struct HatWMPanelConfig {
    const char *position;
    const char *layer;
    int height;
    int exclusive_zone;
    uint32_t background_color;
    uint32_t foreground_color;
    double panel_opacity;
    int border_width;
    uint32_t border_color;
    int shadow_size;
    uint32_t shadow_color;
    const char *font;
    int padding;
    int margin_top;
    int margin_right;
    int margin_bottom;
    int margin_left;
    const char *separator;

    int launcher_group;
    const char *launcher_icon;
    uint32_t launcher_icon_color;
    int launcher_icon_tint;
    int workspace_group;
    int tray_group;
    int clock_group;
    const char *clock_icon;
    uint32_t clock_icon_color;
    int clock_icon_tint;
    int volume_group;
    const char *volume_icon;
    uint32_t volume_icon_color;
    int volume_icon_tint;
    int volume_slider_width;
    int volume_step;
    int brightness_group;
    const char *brightness_icon;
    uint32_t brightness_icon_color;
    int brightness_icon_tint;
    int brightness_slider_width;
    int brightness_step;
    int microphone_group;
    const char *microphone_icon;
    uint32_t microphone_icon_color;
    int microphone_icon_tint;
    int microphone_slider_width;
    int microphone_step;
    int network_group;
    const char *network_icon;
    const char *network_wired_icon;
    uint32_t network_icon_color;
    int network_icon_tint;
    int network_wired_icon_tint;
    int network_show_text;
    int bluetooth_group;
    const char *bluetooth_icon;
    uint32_t bluetooth_icon_color;
    int bluetooth_icon_tint;
    int bluetooth_show_text;
    int battery_group;
    const char *battery_icon;
    uint32_t battery_icon_color;
    int battery_icon_tint;
    int battery_dynamic_icon;
    int battery_show_text;
    int keyboard_layout_group;
    const char *keyboard_layout_icon;
    uint32_t keyboard_layout_icon_color;
    int keyboard_layout_icon_tint;
    int storage_group;
    const char *storage_icon;
    uint32_t storage_icon_color;
    int storage_icon_tint;
    int ram_group;
    const char *ram_icon;
    uint32_t ram_icon_color;
    int ram_icon_tint;
    int cpu_group;
    const char *cpu_icon;
    uint32_t cpu_icon_color;
    int cpu_icon_tint;
    int gpu_group;
    const char *gpu_icon;
    uint32_t gpu_icon_color;
    int gpu_icon_tint;
    int netstat_group;
    const char *netstat_icon;
    uint32_t netstat_icon_color;
    int netstat_icon_tint;
    uint32_t module_background_color;
    uint32_t module_icon_color;
    int module_padding;
    int module_radius;
    int module_icon_size;
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
    uint32_t menu_background_color;
    uint32_t menu_foreground_color;
    uint32_t menu_selected_background_color;
    uint32_t menu_selected_foreground_color;
    uint32_t menu_border_color;
    const char *menu_font;
    int menu_padding;
    int menu_border_width;
    int menu_radius;
} HatWMPanelConfig;

typedef struct HatWMPanelWorkspace {
    int number;
    int active;
    int focused;
    int urgent;
    int windows;
} HatWMPanelWorkspace;

typedef struct HatWMPanelWindow {
    uint64_t id;
    int workspace;
    int mapped;
    int focused;
    int fullscreen;
    int x;
    int y;
    int width;
    int height;
} HatWMPanelWindow;

typedef struct HatWMPanelOutput {
    int x;
    int y;
    int width;
    int height;
    int usable_x;
    int usable_y;
    int usable_width;
    int usable_height;
} HatWMPanelOutput;

typedef struct HatWMPanelButton {
    int group;
    const char *text;
    const char *icon;
    uint32_t icon_color;
    int icon_tint;
    int order;
} HatWMPanelButton;

typedef struct HatWMPanelSeparator {
    int group;
    int width;
    int order;
} HatWMPanelSeparator;

typedef struct HatWMPanelNetwork {
    const char *ssid;
    int signal;
    int secured;
    int active;
    int wired;
} HatWMPanelNetwork;

typedef struct HatWMPanelBluetoothDevice {
    const char *address;
    const char *name;
    int connected;
} HatWMPanelBluetoothDevice;

HatWMPanel *hatwm_panel_create(const HatWMPanelConfig *config);
void hatwm_panel_destroy(HatWMPanel *panel);
void hatwm_panel_set_text(HatWMPanel *panel, const char *left, const char *center, const char *right);
void hatwm_panel_set_clock(HatWMPanel *panel, const char *text);
void hatwm_panel_set_volume(HatWMPanel *panel, int percent, int muted);
void hatwm_panel_set_brightness(HatWMPanel *panel, int percent);
void hatwm_panel_set_microphone(HatWMPanel *panel, int percent, int muted);
void hatwm_panel_set_networks(HatWMPanel *panel,
                              const HatWMPanelNetwork *networks, int network_count);
void hatwm_panel_set_bluetooth_devices(
    HatWMPanel *panel, const HatWMPanelBluetoothDevice *devices, int device_count);
void hatwm_panel_set_battery(HatWMPanel *panel, int percent,
                             int charging, int full);
void hatwm_panel_set_keyboard_layout(HatWMPanel *panel, const char *layout);
void hatwm_panel_set_storage(HatWMPanel *panel, const char *text);
void hatwm_panel_set_ram(HatWMPanel *panel, const char *text);
void hatwm_panel_set_cpu(HatWMPanel *panel, const char *text);
void hatwm_panel_set_gpu(HatWMPanel *panel, const char *text);
void hatwm_panel_set_netstat(HatWMPanel *panel, const char *text);
void hatwm_panel_set_launcher_logo_png(HatWMPanel *panel,
                                       const uint8_t *data, int length);
void hatwm_panel_set_workspaces(HatWMPanel *panel,
                                const HatWMPanelWorkspace *workspaces, int workspace_count,
                                const HatWMPanelWindow *windows, int window_count,
                                const HatWMPanelOutput *output);
void hatwm_panel_set_tray(HatWMPanel *panel, int count);
void hatwm_panel_set_tray_icon(HatWMPanel *panel, int index,
                               const uint32_t *pixels, int width, int height);
void hatwm_panel_set_tray_icon_file(HatWMPanel *panel, int index, const char *path);
void hatwm_panel_set_buttons(HatWMPanel *panel,
                             const HatWMPanelButton *buttons, int button_count);
void hatwm_panel_set_separators(
    HatWMPanel *panel, const HatWMPanelSeparator *separators,
    int separator_count);
void hatwm_panel_set_group_order(HatWMPanel *panel, int group,
                                 int launcher_order, int workspace_order,
                                 int tray_order,
                                 int clock_order, int volume_order,
                                 int brightness_order, int microphone_order,
                                 int network_order, int bluetooth_order,
                                 int battery_order, int keyboard_layout_order,
                                 int storage_order,
                                 int ram_order,
                                 int cpu_order, int gpu_order,
                                 int netstat_order,
                                 int button_order,
                                 int text_order);
void hatwm_panel_redraw(HatWMPanel *panel);
int hatwm_panel_take_workspace_click(HatWMPanel *panel);
int hatwm_panel_take_tray_click(HatWMPanel *panel);
int hatwm_panel_take_tray_button(HatWMPanel *panel);
int hatwm_panel_take_clock_click(HatWMPanel *panel);
int hatwm_panel_take_launcher_click(HatWMPanel *panel);
int hatwm_panel_take_button_click(HatWMPanel *panel);
int hatwm_panel_take_volume_change(HatWMPanel *panel);
int hatwm_panel_take_brightness_change(HatWMPanel *panel);
int hatwm_panel_take_microphone_change(HatWMPanel *panel);
int hatwm_panel_take_network_menu_request(HatWMPanel *panel);
int hatwm_panel_show_network_menu(HatWMPanel *panel);
int hatwm_panel_take_network_click(HatWMPanel *panel);
int hatwm_panel_take_network_disconnect(HatWMPanel *panel);
int hatwm_panel_take_bluetooth_menu_request(HatWMPanel *panel);
int hatwm_panel_show_bluetooth_menu(HatWMPanel *panel);
int hatwm_panel_take_bluetooth_click(HatWMPanel *panel);
int hatwm_panel_take_bluetooth_disconnect(HatWMPanel *panel);
int hatwm_panel_take_keyboard_layout_click(HatWMPanel *panel);
int hatwm_panel_show_calendar(HatWMPanel *panel);
int hatwm_panel_show_launcher(HatWMPanel *panel);
int hatwm_panel_show_tray_menu(HatWMPanel *panel, const char *service, const char *path);
int hatwm_panel_dispatch(HatWMPanel *panel, int timeout_ms);
int hatwm_panel_closed(HatWMPanel *panel);
const char *hatwm_panel_last_error(void);

#endif
