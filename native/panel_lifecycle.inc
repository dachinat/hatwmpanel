static uint32_t parse_layer(const char *layer) {
    if (!layer) return ZWLR_LAYER_SHELL_V1_LAYER_TOP;
    if (strcmp(layer, "background") == 0) return ZWLR_LAYER_SHELL_V1_LAYER_BACKGROUND;
    if (strcmp(layer, "bottom") == 0) return ZWLR_LAYER_SHELL_V1_LAYER_BOTTOM;
    if (strcmp(layer, "overlay") == 0) return ZWLR_LAYER_SHELL_V1_LAYER_OVERLAY;
    return ZWLR_LAYER_SHELL_V1_LAYER_TOP;
}

HatWMPanel *hatwm_panel_create(const HatWMPanelConfig *config) {
    if (!config) {
        set_error("missing panel config");
        return NULL;
    }
    configure_menu_style(config);

    HatWMPanel *panel = calloc(1, sizeof(*panel));
    if (!panel) {
        set_error("out of memory");
        return NULL;
    }
    for (int group = 0; group < 4; group++)
        for (int kind = 0; kind < PANEL_COMPONENT_COUNT; kind++)
            panel->group_order[group][kind] = kind;

    panel->requested_height = config->height > 0 ? config->height : 32;
    panel->padding = config->padding >= 0 ? config->padding : 10;
    panel->background_color = config->background_color;
    panel->foreground_color = config->foreground_color;
    panel->panel_opacity = config->panel_opacity;
    if (panel->panel_opacity < 0) panel->panel_opacity = 0;
    if (panel->panel_opacity > 1) panel->panel_opacity = 1;
    panel->border_width =
        config->border_width >= 0 ? config->border_width : 0;
    panel->border_color = config->border_color;
    panel->shadow_size =
        config->shadow_size >= 0 ? config->shadow_size : 0;
    panel->shadow_color = config->shadow_color;
    panel->font = strdup(config->font ? config->font : "Sans 11");
    panel->separator = strdup(config->separator ? config->separator : "  |  ");
    panel->left = strdup("");
    panel->center = strdup("");
    panel->right = strdup("");
    panel->clock = strdup("");
    panel->clock_group = config->clock_group;
    panel->module_background_color = config->module_background_color;
    panel->module_icon_color = config->module_icon_color;
    panel->module_padding = config->module_padding >= 0 ? config->module_padding : 7;
    panel->module_radius = config->module_radius >= 0 ? config->module_radius : 6;
    panel->module_icon_size = config->module_icon_size > 0 ? config->module_icon_size : 16;
    panel->launcher_group = config->launcher_group;
    panel->launcher_origin_x = -1;
    // Use most of the panel height for detailed launcher artwork.
    panel->launcher_size = panel->requested_height - 4;
    if (panel->launcher_size < 16) panel->launcher_size = 16;
    if (panel->launcher_size > panel->requested_height)
        panel->launcher_size = panel->requested_height;
    panel->launcher_logo = load_system_icon(
        config->launcher_icon, panel->launcher_size,
        config->launcher_icon_color, config->launcher_icon_tint);
    panel->clock_icon = load_system_icon(
        config->clock_icon, panel->module_icon_size, config->clock_icon_color,
        config->clock_icon_tint);
    panel->volume_group = config->volume_group;
    panel->volume_slider_width =
        config->volume_slider_width > 0 ? config->volume_slider_width : 80;
    panel->volume_step = config->volume_step > 0 ? config->volume_step : 5;
    panel->volume_color = config->volume_icon_color;
    panel->volume_percent = 0;
    panel->pending_volume_change = -1;
    panel->volume_origin_x = -1;
    panel->volume_icon = load_system_icon(
        config->volume_icon, panel->module_icon_size, config->volume_icon_color,
        config->volume_icon_tint);
    panel->brightness_group = config->brightness_group;
    panel->brightness_slider_width =
        config->brightness_slider_width > 0 ? config->brightness_slider_width : 80;
    panel->brightness_step = config->brightness_step > 0 ? config->brightness_step : 5;
    panel->brightness_percent = 1;
    panel->pending_brightness_change = -1;
    panel->brightness_origin_x = -1;
    panel->brightness_color = config->brightness_icon_color;
    panel->brightness_icon = load_system_icon(
        config->brightness_icon, panel->module_icon_size,
        config->brightness_icon_color, config->brightness_icon_tint);
    panel->microphone_group = config->microphone_group;
    panel->microphone_slider_width =
        config->microphone_slider_width > 0 ? config->microphone_slider_width : 80;
    panel->microphone_step = config->microphone_step > 0 ? config->microphone_step : 5;
    panel->microphone_percent = 0;
    panel->pending_microphone_change = -1;
    panel->microphone_origin_x = -1;
    panel->microphone_color = config->microphone_icon_color;
    panel->microphone_icon = load_system_icon(
        config->microphone_icon, panel->module_icon_size,
        config->microphone_icon_color, config->microphone_icon_tint);
    panel->network_group = config->network_group;
    panel->network_show_text = config->network_show_text;
    panel->network_origin_x = -1;
    panel->network_label = strdup("Disconnected");
    panel->network_icon = load_system_icon(
        config->network_icon, panel->module_icon_size,
        config->network_icon_color, config->network_icon_tint);
    panel->network_wired_icon = load_system_icon(
        config->network_wired_icon, panel->module_icon_size,
        config->network_icon_color, config->network_wired_icon_tint);
    panel->bluetooth_group = config->bluetooth_group;
    panel->bluetooth_show_text = config->bluetooth_show_text;
    panel->bluetooth_origin_x = -1;
    panel->bluetooth_label = strdup("Bluetooth");
    panel->bluetooth_icon = load_system_icon(
        config->bluetooth_icon, panel->module_icon_size,
        config->bluetooth_icon_color, config->bluetooth_icon_tint);
    panel->battery_group = config->battery_group;
    panel->battery_show_text = config->battery_show_text;
    panel->battery_dynamic_icon = config->battery_dynamic_icon;
    panel->battery_icon_color = config->battery_icon_color;
    panel->battery_icon_tint = config->battery_icon_tint;
    panel->battery_origin_x = -1;
    panel->battery_percent = 0;
    snprintf(panel->battery_label, sizeof(panel->battery_label), "0%%");
    panel->battery_icon_name = strdup(config->battery_icon ? config->battery_icon : "");
    panel->battery_icon = load_system_icon(
        config->battery_icon, panel->module_icon_size,
        config->battery_icon_color, config->battery_icon_tint);
    panel->keyboard_layout_group = config->keyboard_layout_group;
    panel->keyboard_layout_origin_x = -1;
    panel->keyboard_layout = strdup("--");
    panel->keyboard_layout_icon = load_system_icon(
        config->keyboard_layout_icon, panel->module_icon_size,
        config->keyboard_layout_icon_color,
        config->keyboard_layout_icon_tint);
    panel->storage_group = config->storage_group;
    panel->storage_origin_x = -1;
    panel->storage_text = strdup("-- / --");
    panel->storage_icon = load_system_icon(
        config->storage_icon, panel->module_icon_size,
        config->storage_icon_color, config->storage_icon_tint);
    panel->ram_group = config->ram_group;
    panel->ram_origin_x = -1;
    panel->ram_text = strdup("-- / --");
    panel->ram_icon = load_system_icon(
        config->ram_icon, panel->module_icon_size,
        config->ram_icon_color, config->ram_icon_tint);
    panel->cpu_group = config->cpu_group;
    panel->cpu_origin_x = -1;
    panel->cpu_text = strdup("0%");
    panel->cpu_icon = load_system_icon(
        config->cpu_icon, panel->module_icon_size,
        config->cpu_icon_color, config->cpu_icon_tint);
    panel->gpu_group = config->gpu_group;
    panel->gpu_origin_x = -1;
    panel->gpu_text = strdup("N/A");
    panel->gpu_icon = load_system_icon(
        config->gpu_icon, panel->module_icon_size,
        config->gpu_icon_color, config->gpu_icon_tint);
    panel->netstat_group = config->netstat_group;
    panel->netstat_origin_x = -1;
    panel->netstat_text = strdup("↓ 0B/s ↑ 0B/s");
    panel->netstat_icon = load_system_icon(
        config->netstat_icon, panel->module_icon_size,
        config->netstat_icon_color, config->netstat_icon_tint);
    panel->clock_origin_x = -1;
    panel->workspace_group = config->workspace_group;
    panel->tray_group = config->tray_group;
    panel->tray_item_size = panel->requested_height > 24 ? 24 : panel->requested_height;
    if (panel->tray_item_size < 16) panel->tray_item_size = 16;
    panel->workspace_width = config->workspace_width > 0 ? config->workspace_width : 24;
    panel->workspace_height = config->workspace_height > 0 ? config->workspace_height : 22;
    panel->workspace_draw_width = panel->workspace_width;
    panel->workspace_draw_height = panel->workspace_height;
    panel->workspace_gap = config->workspace_gap >= 0 ? config->workspace_gap : 5;
    panel->workspace_inset = config->workspace_inset > 0 ? config->workspace_inset : 2;
    panel->workspace_color = config->workspace_color;
    panel->workspace_active_color = config->workspace_active_color;
    panel->workspace_urgent_color = config->workspace_urgent_color;
    panel->workspace_border_color = config->workspace_border_color;
    panel->workspace_window_color = config->workspace_window_color;
    panel->workspace_focused_window_color = config->workspace_focused_window_color;
    panel->workspace_origin_x = -1;
    panel->workspace_origin_y = -1;
    panel->pointer_x = -1;
    panel->pointer_y = -1;
    panel->tray_origin_x = -1;
    panel->panel_at_bottom =
        config->position && strcmp(config->position, "bottom") == 0;

    panel->display = wl_display_connect(NULL);
    if (!panel->display) {
        set_error("could not connect to Wayland display; is WAYLAND_DISPLAY set?");
        hatwm_panel_destroy(panel);
        return NULL;
    }

    panel->registry = wl_display_get_registry(panel->display);
    wl_registry_add_listener(panel->registry, &registry_listener, panel);
    if (wl_display_roundtrip(panel->display) < 0) {
        set_error("Wayland registry roundtrip failed");
        hatwm_panel_destroy(panel);
        return NULL;
    }

    if (!panel->compositor || !panel->shm || !panel->layer_shell) {
        set_error("compositor is missing wl_compositor, wl_shm, or wlr-layer-shell-v1");
        hatwm_panel_destroy(panel);
        return NULL;
    }

    panel->surface = wl_compositor_create_surface(panel->compositor);
    if (!panel->surface) {
        set_error("failed to create wl_surface");
        hatwm_panel_destroy(panel);
        return NULL;
    }

    panel->layer_surface = zwlr_layer_shell_v1_get_layer_surface(
        panel->layer_shell, panel->surface, NULL, parse_layer(config->layer), "hatwmpanel");
    if (!panel->layer_surface) {
        set_error("failed to create layer surface");
        hatwm_panel_destroy(panel);
        return NULL;
    }
    zwlr_layer_surface_v1_add_listener(panel->layer_surface, &layer_surface_listener, panel);

    uint32_t anchor = ZWLR_LAYER_SURFACE_V1_ANCHOR_LEFT | ZWLR_LAYER_SURFACE_V1_ANCHOR_RIGHT;
    if (panel->panel_at_bottom) {
        anchor |= ZWLR_LAYER_SURFACE_V1_ANCHOR_BOTTOM;
    } else {
        anchor |= ZWLR_LAYER_SURFACE_V1_ANCHOR_TOP;
    }

    zwlr_layer_surface_v1_set_size(panel->layer_surface, 0, (uint32_t)panel->requested_height);
    zwlr_layer_surface_v1_set_anchor(panel->layer_surface, anchor);
    zwlr_layer_surface_v1_set_margin(panel->layer_surface,
                                     config->margin_top, config->margin_right,
                                     config->margin_bottom, config->margin_left);
    zwlr_layer_surface_v1_set_keyboard_interactivity(panel->layer_surface,
                                                      ZWLR_LAYER_SURFACE_V1_KEYBOARD_INTERACTIVITY_NONE);
    zwlr_layer_surface_v1_set_exclusive_zone(panel->layer_surface,
                                              config->exclusive_zone ? panel->requested_height : 0);

    // Layer-shell requires an initial commit without a buffer before configure.
    wl_surface_commit(panel->surface);
    wl_display_flush(panel->display);
    return panel;
}

void hatwm_panel_destroy(HatWMPanel *panel) {
    if (!panel) return;
    if (panel->launcher_window)
        gtk_widget_destroy(panel->launcher_window);
    if (panel->calendar_window) gtk_widget_destroy(panel->calendar_window);
    if (panel->network_menu) gtk_widget_destroy(panel->network_menu);
    if (panel->bluetooth_menu) gtk_widget_destroy(panel->bluetooth_menu);
    if (panel->pointer) wl_pointer_destroy(panel->pointer);
    if (panel->seat) wl_seat_destroy(panel->seat);
    if (panel->layer_surface) zwlr_layer_surface_v1_destroy(panel->layer_surface);
    if (panel->surface) wl_surface_destroy(panel->surface);
    if (panel->layer_shell) zwlr_layer_shell_v1_destroy(panel->layer_shell);
    if (panel->shm) wl_shm_destroy(panel->shm);
    if (panel->compositor) wl_compositor_destroy(panel->compositor);
    if (panel->registry) wl_registry_destroy(panel->registry);
    if (panel->display) wl_display_disconnect(panel->display);
    free(panel->font);
    free(panel->separator);
    free(panel->left);
    free(panel->center);
    free(panel->right);
    free(panel->clock);
    if (panel->launcher_logo)
        cairo_surface_destroy(panel->launcher_logo);
    if (panel->clock_icon) cairo_surface_destroy(panel->clock_icon);
    if (panel->volume_icon) cairo_surface_destroy(panel->volume_icon);
    if (panel->brightness_icon) cairo_surface_destroy(panel->brightness_icon);
    if (panel->microphone_icon) cairo_surface_destroy(panel->microphone_icon);
    if (panel->network_icon) cairo_surface_destroy(panel->network_icon);
    if (panel->network_wired_icon)
        cairo_surface_destroy(panel->network_wired_icon);
    free(panel->network_label);
    if (panel->networks) {
        for (int i = 0; i < panel->network_count; i++)
            free(panel->networks[i].ssid);
        free(panel->networks);
    }
    if (panel->bluetooth_icon) cairo_surface_destroy(panel->bluetooth_icon);
    free(panel->bluetooth_label);
    if (panel->bluetooth_devices) {
        for (int i = 0; i < panel->bluetooth_device_count; i++) {
            free(panel->bluetooth_devices[i].address);
            free(panel->bluetooth_devices[i].name);
        }
        free(panel->bluetooth_devices);
    }
    if (panel->battery_icon) cairo_surface_destroy(panel->battery_icon);
    free(panel->battery_icon_name);
    if (panel->keyboard_layout_icon)
        cairo_surface_destroy(panel->keyboard_layout_icon);
    free(panel->keyboard_layout);
    if (panel->storage_icon) cairo_surface_destroy(panel->storage_icon);
    free(panel->storage_text);
    if (panel->ram_icon) cairo_surface_destroy(panel->ram_icon);
    free(panel->ram_text);
    if (panel->cpu_icon) cairo_surface_destroy(panel->cpu_icon);
    free(panel->cpu_text);
    if (panel->gpu_icon) cairo_surface_destroy(panel->gpu_icon);
    free(panel->gpu_text);
    if (panel->netstat_icon) cairo_surface_destroy(panel->netstat_icon);
    free(panel->netstat_text);
    if (panel->buttons) {
        for (int i = 0; i < panel->button_count; i++) {
            free(panel->buttons[i].text);
            if (panel->buttons[i].icon)
                cairo_surface_destroy(panel->buttons[i].icon);
        }
        free(panel->buttons);
    }
    free(panel->separators);
    if (panel->tray_icons) {
        for (int i = 0; i < panel->tray_count; i++)
            if (panel->tray_icons[i]) cairo_surface_destroy(panel->tray_icons[i]);
        free(panel->tray_icons);
    }
    free(panel->workspaces);
    free(panel->windows);
    free(panel);
}
