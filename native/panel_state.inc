static void replace_string(char **target, const char *value) {
    char *copy = strdup(value ? value : "");
    if (!copy) return;
    free(*target);
    *target = copy;
}

void hatwm_panel_set_text(HatWMPanel *panel, const char *left, const char *center, const char *right) {
    if (!panel) return;
    replace_string(&panel->left, left);
    replace_string(&panel->center, center);
    replace_string(&panel->right, right);
    panel_draw(panel);
}

void hatwm_panel_set_clock(HatWMPanel *panel, const char *text) {
    if (!panel) return;
    replace_string(&panel->clock, text);
    panel_draw(panel);
}

void hatwm_panel_set_launcher_logo_svg(HatWMPanel *panel,
                                       const uint8_t *data, int length,
                                       uint32_t outline_color,
                                       uint32_t fill_color) {
    if (!panel || !data || length <= 0 || length > 1024 * 1024) return;

    GError *error = NULL;
    RsvgHandle *handle = rsvg_handle_new_from_data(
        data, (gsize)length, &error);
    if (!handle) {
        if (error) g_error_free(error);
        return;
    }

    uint32_t outline_r = (outline_color >> 24) & 0xff;
    uint32_t outline_g = (outline_color >> 16) & 0xff;
    uint32_t outline_b = (outline_color >> 8) & 0xff;
    uint32_t outline_a = outline_color & 0xff;
    uint32_t fill_r = (fill_color >> 24) & 0xff;
    uint32_t fill_g = (fill_color >> 16) & 0xff;
    uint32_t fill_b = (fill_color >> 8) & 0xff;
    uint32_t fill_a = fill_color & 0xff;
#define SCALED_CHANNEL(channel, percent) \
    (((channel) * (percent) / 100) > 255 ? 255 : \
     ((channel) * (percent) / 100))
    char css[2048];
    int css_length = snprintf(
        css, sizeof(css),
        "#fill-start{stop-color:#%02x%02x%02x;stop-opacity:%.3f}"
        "#fill-middle{stop-color:#%02x%02x%02x;stop-opacity:%.3f}"
        "#fill-end{stop-color:#%02x%02x%02x;stop-opacity:%.3f}"
        "#outline-start{stop-color:#%02x%02x%02x;stop-opacity:%.3f}"
        "#outline-middle{stop-color:#%02x%02x%02x;stop-opacity:%.3f}"
        "#outline-end{stop-color:#%02x%02x%02x;stop-opacity:%.3f}"
        "#outline-glow{flood-color:#%02x%02x%02x;flood-opacity:%.3f}"
        "#inner-highlight{stroke:#%02x%02x%02x;stroke-opacity:%.3f}",
        SCALED_CHANNEL(fill_r, 125), SCALED_CHANNEL(fill_g, 125),
        SCALED_CHANNEL(fill_b, 125), fill_a / 255.0,
        fill_r, fill_g, fill_b, fill_a / 255.0,
        SCALED_CHANNEL(fill_r, 75), SCALED_CHANNEL(fill_g, 75),
        SCALED_CHANNEL(fill_b, 75), fill_a / 255.0,
        SCALED_CHANNEL(outline_r, 120), SCALED_CHANNEL(outline_g, 120),
        SCALED_CHANNEL(outline_b, 120), outline_a / 255.0,
        outline_r, outline_g, outline_b, outline_a / 255.0,
        SCALED_CHANNEL(outline_r, 110), SCALED_CHANNEL(outline_g, 110),
        SCALED_CHANNEL(outline_b, 110), outline_a / 255.0,
        outline_r, outline_g, outline_b, 0.38 * outline_a / 255.0,
        SCALED_CHANNEL(outline_r, 140), SCALED_CHANNEL(outline_g, 140),
        SCALED_CHANNEL(outline_b, 140), 0.72 * outline_a / 255.0);
#undef SCALED_CHANNEL
    if (css_length < 0 || css_length >= (int)sizeof(css) ||
        !rsvg_handle_set_stylesheet(
            handle, (const guint8 *)css, (gsize)css_length, &error)) {
        g_object_unref(handle);
        if (error) g_error_free(error);
        return;
    }

    int size = panel->launcher_size;
    cairo_surface_t *surface = cairo_image_surface_create(
        CAIRO_FORMAT_ARGB32, size, size);
    cairo_t *cr = cairo_create(surface);
    cairo_set_operator(cr, CAIRO_OPERATOR_SOURCE);
    cairo_set_source_rgba(cr, 0, 0, 0, 0);
    cairo_paint(cr);
    cairo_set_operator(cr, CAIRO_OPERATOR_OVER);
    RsvgRectangle viewport = {0, 0, size, size};
    if (!rsvg_handle_render_document(handle, cr, &viewport, &error)) {
        cairo_destroy(cr);
        cairo_surface_destroy(surface);
        g_object_unref(handle);
        if (error) g_error_free(error);
        return;
    }
    cairo_destroy(cr);
    g_object_unref(handle);

    if (panel->launcher_logo) cairo_surface_destroy(panel->launcher_logo);
    panel->launcher_logo = surface;
    panel_draw(panel);
}

int hatwm_panel_take_clock_click(HatWMPanel *panel) {
    if (!panel) return 0;
    int clicked = panel->pending_clock_click;
    panel->pending_clock_click = 0;
    return clicked;
}

int hatwm_panel_take_launcher_click(HatWMPanel *panel) {
    if (!panel) return 0;
    int clicked = panel->pending_launcher_click;
    panel->pending_launcher_click = 0;
    return clicked;
}

void hatwm_panel_set_volume(HatWMPanel *panel, int percent, int muted) {
    if (!panel) return;
    panel->volume_percent = clamp_int(percent, 0, 100);
    panel->volume_muted = muted != 0;
    panel_draw(panel);
}

int hatwm_panel_take_volume_change(HatWMPanel *panel) {
    if (!panel) return -1;
    int percent = panel->pending_volume_change;
    panel->pending_volume_change = -1;
    return percent;
}

void hatwm_panel_set_brightness(HatWMPanel *panel, int percent) {
    if (!panel) return;
    panel->brightness_percent = clamp_int(percent, 1, 100);
    panel_draw(panel);
}

int hatwm_panel_take_brightness_change(HatWMPanel *panel) {
    if (!panel) return -1;
    int percent = panel->pending_brightness_change;
    panel->pending_brightness_change = -1;
    return percent;
}

void hatwm_panel_set_microphone(HatWMPanel *panel, int percent, int muted) {
    if (!panel) return;
    panel->microphone_percent = clamp_int(percent, 0, 100);
    panel->microphone_muted = muted != 0;
    panel_draw(panel);
}

int hatwm_panel_take_microphone_change(HatWMPanel *panel) {
    if (!panel) return -1;
    int percent = panel->pending_microphone_change;
    panel->pending_microphone_change = -1;
    return percent;
}

void hatwm_panel_set_networks(HatWMPanel *panel,
                              const HatWMPanelNetwork *networks, int network_count) {
    if (!panel) return;
    if (panel->networks) {
        for (int i = 0; i < panel->network_count; i++)
            free(panel->networks[i].ssid);
        free(panel->networks);
    }
    panel->networks = NULL;
    panel->network_count = 0;
    if (network_count < 0) network_count = 0;
    if (network_count > 256) network_count = 256;
    if (networks && network_count > 0) {
        panel->networks = calloc((size_t)network_count, sizeof(*panel->networks));
        if (panel->networks) {
            panel->network_count = network_count;
            for (int i = 0; i < network_count; i++) {
                panel->networks[i].ssid = strdup(networks[i].ssid ? networks[i].ssid : "");
                panel->networks[i].signal = clamp_int(networks[i].signal, 0, 100);
                panel->networks[i].secured = networks[i].secured != 0;
                panel->networks[i].active = networks[i].active != 0;
                panel->networks[i].wired = networks[i].wired != 0;
            }
        }
    }
    int has_active = 0;
    panel->network_active_wired = 0;
    for (int i = 0; i < panel->network_count; i++) {
        if (!panel->networks[i].active) continue;
        if (!has_active ||
            (panel->networks[i].wired && !panel->network_active_wired)) {
            replace_string(&panel->network_label, panel->networks[i].ssid);
            panel->network_active_wired = panel->networks[i].wired;
        }
        has_active = 1;
    }
    if (!has_active) {
        replace_string(&panel->network_label, "Disconnected");
        panel->network_active_wired = 0;
    }
    panel_draw(panel);
}

int hatwm_panel_take_network_menu_request(HatWMPanel *panel) {
    if (!panel) return 0;
    int requested = panel->pending_network_menu;
    panel->pending_network_menu = 0;
    return requested;
}

int hatwm_panel_take_network_click(HatWMPanel *panel) {
    if (!panel) return 0;
    int clicked = panel->pending_network_click;
    panel->pending_network_click = 0;
    return clicked;
}

int hatwm_panel_take_network_disconnect(HatWMPanel *panel) {
    if (!panel) return 0;
    int requested = panel->pending_network_disconnect;
    panel->pending_network_disconnect = 0;
    return requested;
}

void hatwm_panel_set_bluetooth_devices(
    HatWMPanel *panel, const HatWMPanelBluetoothDevice *devices, int device_count) {
    if (!panel) return;
    int reopen_menu = panel->bluetooth_menu != NULL;
    if (reopen_menu) gtk_widget_destroy(panel->bluetooth_menu);
    if (panel->bluetooth_devices) {
        for (int i = 0; i < panel->bluetooth_device_count; i++) {
            free(panel->bluetooth_devices[i].address);
            free(panel->bluetooth_devices[i].name);
        }
        free(panel->bluetooth_devices);
    }
    panel->bluetooth_devices = NULL;
    panel->bluetooth_device_count = 0;
    if (device_count < 0) device_count = 0;
    if (device_count > 256) device_count = 256;
    if (devices && device_count > 0) {
        panel->bluetooth_devices =
            calloc((size_t)device_count, sizeof(*panel->bluetooth_devices));
        if (panel->bluetooth_devices) {
            panel->bluetooth_device_count = device_count;
            for (int i = 0; i < device_count; i++) {
                panel->bluetooth_devices[i].address =
                    strdup(devices[i].address ? devices[i].address : "");
                panel->bluetooth_devices[i].name =
                    strdup(devices[i].name ? devices[i].name : "");
                panel->bluetooth_devices[i].connected = devices[i].connected != 0;
            }
        }
    }
    int has_connected = 0;
    for (int i = 0; i < panel->bluetooth_device_count; i++) {
        if (!panel->bluetooth_devices[i].connected) continue;
        if (!has_connected)
            replace_string(&panel->bluetooth_label,
                           panel->bluetooth_devices[i].name);
        has_connected = 1;
    }
    if (!has_connected) replace_string(&panel->bluetooth_label, "Bluetooth");
    panel_draw(panel);
    if (reopen_menu) hatwm_panel_show_bluetooth_menu(panel);
}

int hatwm_panel_take_bluetooth_menu_request(HatWMPanel *panel) {
    if (!panel) return 0;
    int requested = panel->pending_bluetooth_menu;
    panel->pending_bluetooth_menu = 0;
    return requested;
}

int hatwm_panel_take_bluetooth_click(HatWMPanel *panel) {
    if (!panel) return 0;
    int clicked = panel->pending_bluetooth_click;
    panel->pending_bluetooth_click = 0;
    return clicked;
}

int hatwm_panel_take_bluetooth_disconnect(HatWMPanel *panel) {
    if (!panel) return 0;
    int requested = panel->pending_bluetooth_disconnect;
    panel->pending_bluetooth_disconnect = 0;
    return requested;
}

void hatwm_panel_set_battery(HatWMPanel *panel, int percent,
                             int charging, int full) {
    if (!panel) return;
    panel->battery_percent = clamp_int(percent, 0, 100);
    panel->battery_charging = charging != 0;
    panel->battery_full = full != 0;
    snprintf(panel->battery_label, sizeof(panel->battery_label),
             "%d%%", panel->battery_percent);

    if (panel->battery_dynamic_icon) {
        char icon_name[80];
        if (panel->battery_full || panel->battery_percent >= 100) {
            snprintf(icon_name, sizeof(icon_name),
                     "battery-level-100-charged-symbolic");
        } else {
            int level = ((panel->battery_percent + 5) / 10) * 10;
            level = clamp_int(level, 0, 90);
            snprintf(icon_name, sizeof(icon_name),
                     panel->battery_charging
                         ? "battery-level-%d-charging-symbolic"
                         : "battery-level-%d-symbolic",
                     level);
        }
        if (!panel->battery_icon_name ||
            strcmp(panel->battery_icon_name, icon_name) != 0) {
            cairo_surface_t *icon = load_system_icon(
                icon_name, panel->module_icon_size,
                panel->battery_icon_color, panel->battery_icon_tint);
            if (icon) {
                if (panel->battery_icon)
                    cairo_surface_destroy(panel->battery_icon);
                panel->battery_icon = icon;
                replace_string(&panel->battery_icon_name, icon_name);
            }
        }
    }
    panel_draw(panel);
}

void hatwm_panel_set_keyboard_layout(HatWMPanel *panel, const char *layout) {
    if (!panel) return;
    replace_string(&panel->keyboard_layout,
                   layout && layout[0] ? layout : "--");
    panel_draw(panel);
}

int hatwm_panel_take_keyboard_layout_click(HatWMPanel *panel) {
    if (!panel) return 0;
    int clicked = panel->pending_keyboard_layout_click;
    panel->pending_keyboard_layout_click = 0;
    return clicked;
}

void hatwm_panel_set_storage(HatWMPanel *panel, const char *text) {
    if (!panel) return;
    replace_string(&panel->storage_text, text && text[0] ? text : "-- / --");
    panel_draw(panel);
}

void hatwm_panel_set_ram(HatWMPanel *panel, const char *text) {
    if (!panel) return;
    replace_string(&panel->ram_text, text && text[0] ? text : "-- / --");
    panel_draw(panel);
}

void hatwm_panel_set_cpu(HatWMPanel *panel, const char *text) {
    if (!panel) return;
    replace_string(&panel->cpu_text, text && text[0] ? text : "0%");
    panel_draw(panel);
}

void hatwm_panel_set_gpu(HatWMPanel *panel, const char *text) {
    if (!panel) return;
    replace_string(&panel->gpu_text, text && text[0] ? text : "N/A");
    panel_draw(panel);
}

void hatwm_panel_set_netstat(HatWMPanel *panel, const char *text) {
    if (!panel) return;
    replace_string(&panel->netstat_text,
                   text && text[0] ? text : "↓ 0B/s ↑ 0B/s");
    panel_draw(panel);
}

void hatwm_panel_set_buttons(HatWMPanel *panel,
                             const HatWMPanelButton *buttons, int button_count) {
    if (!panel) return;
    if (panel->buttons) {
        for (int i = 0; i < panel->button_count; i++) {
            free(panel->buttons[i].text);
            if (panel->buttons[i].icon)
                cairo_surface_destroy(panel->buttons[i].icon);
        }
        free(panel->buttons);
    }
    panel->buttons = NULL;
    panel->button_count = 0;
    panel->pending_button_click = 0;
    if (!buttons || button_count <= 0) {
        panel_draw(panel);
        return;
    }
    if (button_count > 128) button_count = 128;
    panel->buttons = calloc((size_t)button_count, sizeof(*panel->buttons));
    if (!panel->buttons) return;
    panel->button_count = button_count;
    for (int i = 0; i < button_count; i++) {
        panel->buttons[i].group = buttons[i].group;
        panel->buttons[i].order = buttons[i].order;
        panel->buttons[i].text = strdup(buttons[i].text ? buttons[i].text : "");
        panel->buttons[i].icon = load_system_icon(
            buttons[i].icon, panel->module_icon_size,
            buttons[i].icon_color, buttons[i].icon_tint);
        panel->buttons[i].origin_x = -1;
    }
    panel_draw(panel);
}

void hatwm_panel_set_separators(
        HatWMPanel *panel, const HatWMPanelSeparator *separators,
        int separator_count) {
    if (!panel) return;
    free(panel->separators);
    panel->separators = NULL;
    panel->separator_count = 0;
    if (!separators || separator_count <= 0) {
        panel_draw(panel);
        return;
    }
    if (separator_count > 128) separator_count = 128;
    panel->separators =
        calloc((size_t)separator_count, sizeof(*panel->separators));
    if (!panel->separators) return;
    panel->separator_count = separator_count;
    for (int i = 0; i < separator_count; i++) {
        panel->separators[i].group = separators[i].group;
        panel->separators[i].width = separators[i].width;
        panel->separators[i].order = separators[i].order;
    }
    panel_draw(panel);
}

int hatwm_panel_take_button_click(HatWMPanel *panel) {
    if (!panel) return 0;
    int clicked = panel->pending_button_click;
    panel->pending_button_click = 0;
    return clicked;
}

void hatwm_panel_set_group_order(HatWMPanel *panel, int group,
                                 int launcher_order, int workspace_order,
                                 int tray_order,
                                 int clock_order, int volume_order,
                                 int brightness_order, int microphone_order,
                                 int network_order, int bluetooth_order,
                                 int battery_order, int keyboard_layout_order,
                                 int storage_order, int ram_order,
                                 int cpu_order, int gpu_order,
                                 int netstat_order,
                                 int button_order,
                                 int text_order) {
    if (!panel || group < HATWM_PANEL_GROUP_LEFT ||
        group > HATWM_PANEL_GROUP_RIGHT) return;
    panel->group_order[group][PANEL_COMPONENT_LAUNCHER] = launcher_order;
    panel->group_order[group][PANEL_COMPONENT_WORKSPACES] = workspace_order;
    panel->group_order[group][PANEL_COMPONENT_TRAY] = tray_order;
    panel->group_order[group][PANEL_COMPONENT_CLOCK] = clock_order;
    panel->group_order[group][PANEL_COMPONENT_VOLUME] = volume_order;
    panel->group_order[group][PANEL_COMPONENT_BRIGHTNESS] = brightness_order;
    panel->group_order[group][PANEL_COMPONENT_MICROPHONE] = microphone_order;
    panel->group_order[group][PANEL_COMPONENT_NETWORK] = network_order;
    panel->group_order[group][PANEL_COMPONENT_BLUETOOTH] = bluetooth_order;
    panel->group_order[group][PANEL_COMPONENT_BATTERY] = battery_order;
    panel->group_order[group][PANEL_COMPONENT_KEYBOARD_LAYOUT] =
        keyboard_layout_order;
    panel->group_order[group][PANEL_COMPONENT_STORAGE] = storage_order;
    panel->group_order[group][PANEL_COMPONENT_RAM] = ram_order;
    panel->group_order[group][PANEL_COMPONENT_CPU] = cpu_order;
    panel->group_order[group][PANEL_COMPONENT_GPU] = gpu_order;
    panel->group_order[group][PANEL_COMPONENT_NETSTAT] = netstat_order;
    panel->group_order[group][PANEL_COMPONENT_BUTTONS] = button_order;
    panel->group_order[group][PANEL_COMPONENT_TEXT] = text_order;
    panel_draw(panel);
}

void hatwm_panel_set_workspaces(HatWMPanel *panel,
                                const HatWMPanelWorkspace *workspaces, int workspace_count,
                                const HatWMPanelWindow *windows, int window_count,
                                const HatWMPanelOutput *output) {
    if (!panel) return;
    if (workspace_count < 0) workspace_count = 0;
    if (workspace_count > 128) workspace_count = 128;
    if (window_count < 0) window_count = 0;
    if (window_count > 4096) window_count = 4096;

    HatWMPanelWorkspace *workspace_copy = NULL;
    HatWMPanelWindow *window_copy = NULL;
    if (workspace_count > 0 && workspaces) {
        workspace_copy = calloc((size_t)workspace_count, sizeof(*workspace_copy));
        if (!workspace_copy) return;
        memcpy(workspace_copy, workspaces, (size_t)workspace_count * sizeof(*workspace_copy));
    }
    if (window_count > 0 && windows) {
        window_copy = calloc((size_t)window_count, sizeof(*window_copy));
        if (!window_copy) {
            free(workspace_copy);
            return;
        }
        memcpy(window_copy, windows, (size_t)window_count * sizeof(*window_copy));
    }

    free(panel->workspaces);
    free(panel->windows);
    panel->workspaces = workspace_copy;
    panel->workspace_count = workspace_count;
    panel->windows = window_copy;
    panel->window_count = window_count;
    if (output) panel->output = *output;
    panel_draw(panel);
}

void hatwm_panel_set_tray(HatWMPanel *panel, int count) {
    if (!panel) return;
    if (count < 0) count = 0;
    if (count > 128) count = 128;
    if (panel->tray_icons) {
        for (int i = 0; i < panel->tray_count; i++)
            if (panel->tray_icons[i]) cairo_surface_destroy(panel->tray_icons[i]);
        free(panel->tray_icons);
    }
    panel->tray_icons = count > 0 ? calloc((size_t)count, sizeof(*panel->tray_icons)) : NULL;
    panel->tray_count = count;
}

void hatwm_panel_set_tray_icon(HatWMPanel *panel, int index,
                               const uint32_t *pixels, int width, int height) {
    if (!panel || !panel->tray_icons || index < 0 || index >= panel->tray_count ||
        !pixels || width <= 0 || height <= 0 || width > 1024 || height > 1024) return;
    size_t size = (size_t)width * (size_t)height * sizeof(uint32_t);
    unsigned char *copy = malloc(size);
    if (!copy) return;
    memcpy(copy, pixels, size);
    cairo_surface_t *surface = cairo_image_surface_create_for_data(
        copy, CAIRO_FORMAT_ARGB32, width, height, width * (int)sizeof(uint32_t));
    if (cairo_surface_status(surface) != CAIRO_STATUS_SUCCESS) {
        cairo_surface_destroy(surface);
        free(copy);
        return;
    }
    static const cairo_user_data_key_t icon_data_key;
    if (cairo_surface_set_user_data(surface, &icon_data_key, copy, free) != CAIRO_STATUS_SUCCESS) {
        cairo_surface_destroy(surface);
        free(copy);
        return;
    }
    panel->tray_icons[index] = surface;
}

void hatwm_panel_set_tray_icon_file(HatWMPanel *panel, int index, const char *path) {
    if (!panel || !panel->tray_icons || index < 0 || index >= panel->tray_count ||
        !path || !path[0]) return;

    cairo_surface_t *surface = NULL;
    size_t length = strlen(path);
    if (length >= 4 && strcasecmp(path + length - 4, ".png") == 0) {
        surface = cairo_image_surface_create_from_png(path);
    } else {
        GError *error = NULL;
        RsvgHandle *handle = rsvg_handle_new_from_file(path, &error);
        if (!handle) {
            if (error) g_error_free(error);
            return;
        }
        const int size = panel->tray_item_size > 4 ? panel->tray_item_size - 4 : 16;
        surface = cairo_image_surface_create(CAIRO_FORMAT_ARGB32, size, size);
        cairo_t *icon_cr = cairo_create(surface);
        RsvgRectangle viewport = {0, 0, size, size};
        if (!rsvg_handle_render_document(handle, icon_cr, &viewport, &error)) {
            cairo_destroy(icon_cr);
            cairo_surface_destroy(surface);
            g_object_unref(handle);
            if (error) g_error_free(error);
            return;
        }
        cairo_destroy(icon_cr);
        g_object_unref(handle);
    }
    if (!surface || cairo_surface_status(surface) != CAIRO_STATUS_SUCCESS) {
        if (surface) cairo_surface_destroy(surface);
        return;
    }
    panel->tray_icons[index] = surface;
}

void hatwm_panel_redraw(HatWMPanel *panel) {
    if (panel) panel_draw(panel);
}

int hatwm_panel_take_workspace_click(HatWMPanel *panel) {
    if (!panel) return 0;
    int workspace = panel->pending_workspace_click;
    panel->pending_workspace_click = 0;
    return workspace;
}

int hatwm_panel_take_tray_click(HatWMPanel *panel) {
    if (!panel) return -1;
    int index = panel->pending_tray_click - 1;
    panel->pending_tray_click = 0;
    return index;
}

int hatwm_panel_take_tray_button(HatWMPanel *panel) {
    if (!panel) return 0;
    int button = panel->pending_tray_button;
    panel->pending_tray_button = 0;
    return button;
}

typedef struct PendingMenu {
    GtkWidget *menu;
    GtkWidget *anchor;
    HatWMPanel *network_panel;
    HatWMPanel *bluetooth_panel;
    int panel_at_bottom;
    int align_right;
    int align_center;
    int attempts;
    guint popup_source;
    guint destroy_source;
} PendingMenu;
