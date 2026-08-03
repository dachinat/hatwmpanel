static uint32_t min_u32(uint32_t a, uint32_t b) {
    return a < b ? a : b;
}

static void cancel_slider_drag(HatWMPanel *panel) {
    if (!panel) return;
    panel->pointer_left_pressed = 0;
    panel->volume_dragging = 0;
    panel->brightness_dragging = 0;
    panel->microphone_dragging = 0;
}

static void pointer_enter(void *data, struct wl_pointer *pointer, uint32_t serial,
                          struct wl_surface *surface, wl_fixed_t sx, wl_fixed_t sy) {
    (void)pointer;
    (void)serial;
    HatWMPanel *panel = data;
    if (!panel || surface != panel->surface) return;
    panel->pointer_x = wl_fixed_to_double(sx);
    panel->pointer_y = wl_fixed_to_double(sy);
}

static void pointer_leave(void *data, struct wl_pointer *pointer, uint32_t serial,
                          struct wl_surface *surface) {
    (void)pointer;
    (void)serial;
    (void)surface;
    HatWMPanel *panel = data;
    if (!panel) return;
    cancel_slider_drag(panel);
    panel->pointer_x = -1;
    panel->pointer_y = -1;
}

static void pointer_motion(void *data, struct wl_pointer *pointer, uint32_t time,
                           wl_fixed_t sx, wl_fixed_t sy) {
    (void)pointer;
    (void)time;
    HatWMPanel *panel = data;
    if (!panel) return;
    panel->pointer_x = wl_fixed_to_double(sx);
    panel->pointer_y = wl_fixed_to_double(sy);
    if (panel->pointer_left_pressed &&
        panel->volume_dragging && panel->volume_slider_width > 0) {
        int percent = (int)round(
            (panel->pointer_x - panel->volume_slider_origin_x) * 100.0 /
            panel->volume_slider_width);
        panel->pending_volume_change = clamp_int(percent, 0, 100);
        panel->volume_percent = panel->pending_volume_change;
        panel_draw(panel);
    }
    if (panel->pointer_left_pressed &&
        panel->brightness_dragging && panel->brightness_slider_width > 0) {
        int percent = (int)round(
            (panel->pointer_x - panel->brightness_slider_origin_x) * 100.0 /
            panel->brightness_slider_width);
        panel->pending_brightness_change = clamp_int(percent, 1, 100);
        panel->brightness_percent = panel->pending_brightness_change;
        panel_draw(panel);
    }
    if (panel->pointer_left_pressed &&
        panel->microphone_dragging && panel->microphone_slider_width > 0) {
        int percent = (int)round(
            (panel->pointer_x - panel->microphone_slider_origin_x) * 100.0 /
            panel->microphone_slider_width);
        panel->pending_microphone_change = clamp_int(percent, 0, 100);
        panel->microphone_percent = panel->pending_microphone_change;
        panel_draw(panel);
    }
}

static void pointer_button(void *data, struct wl_pointer *pointer, uint32_t serial,
                           uint32_t time, uint32_t button, uint32_t state) {
    (void)pointer;
    (void)serial;
    (void)time;
    HatWMPanel *panel = data;
    const uint32_t button_left = 0x110;   // BTN_LEFT
    const uint32_t button_right = 0x111;  // BTN_RIGHT
    const uint32_t button_middle = 0x112; // BTN_MIDDLE
    if (!panel) return;
    if (state == WL_POINTER_BUTTON_STATE_RELEASED) {
        if (button == button_left) {
            cancel_slider_drag(panel);
        }
        return;
    }
    if (state != WL_POINTER_BUTTON_STATE_PRESSED) return;
    if (button != button_left && button != button_right && button != button_middle) return;
    if (button == button_left) {
        cancel_slider_drag(panel);
        panel->pointer_left_pressed = 1;
    }
    if (button == button_left && panel->volume_origin_x >= 0 &&
        panel->pointer_x >= panel->volume_slider_origin_x &&
        panel->pointer_x <= panel->volume_slider_origin_x +
                            panel->volume_slider_width) {
        int percent = (int)round(
            (panel->pointer_x - panel->volume_slider_origin_x) * 100.0 /
            panel->volume_slider_width);
        panel->pending_volume_change = clamp_int(percent, 0, 100);
        panel->volume_percent = panel->pending_volume_change;
        panel->volume_dragging = 1;
        panel_draw(panel);
        return;
    }
    if (button == button_left && panel->brightness_origin_x >= 0 &&
        panel->pointer_x >= panel->brightness_slider_origin_x &&
        panel->pointer_x <= panel->brightness_slider_origin_x +
                            panel->brightness_slider_width) {
        int percent = (int)round(
            (panel->pointer_x - panel->brightness_slider_origin_x) * 100.0 /
            panel->brightness_slider_width);
        panel->pending_brightness_change = clamp_int(percent, 1, 100);
        panel->brightness_percent = panel->pending_brightness_change;
        panel->brightness_dragging = 1;
        panel_draw(panel);
        return;
    }
    if (button == button_left && panel->microphone_origin_x >= 0 &&
        panel->pointer_x >= panel->microphone_slider_origin_x &&
        panel->pointer_x <= panel->microphone_slider_origin_x +
                            panel->microphone_slider_width) {
        int percent = (int)round(
            (panel->pointer_x - panel->microphone_slider_origin_x) * 100.0 /
            panel->microphone_slider_width);
        panel->pending_microphone_change = clamp_int(percent, 0, 100);
        panel->microphone_percent = panel->pending_microphone_change;
        panel->microphone_dragging = 1;
        panel_draw(panel);
        return;
    }
    if (button == button_left && panel->network_origin_x >= 0 &&
        panel->pointer_x >= panel->network_origin_x &&
        panel->pointer_x < panel->network_origin_x + panel->network_width) {
        panel->pending_network_menu = 1;
        return;
    }
    if (button == button_left && panel->bluetooth_origin_x >= 0 &&
        panel->pointer_x >= panel->bluetooth_origin_x &&
        panel->pointer_x < panel->bluetooth_origin_x + panel->bluetooth_width) {
        panel->pending_bluetooth_menu = 1;
        return;
    }
    if (button == button_left && panel->keyboard_layout_origin_x >= 0 &&
        panel->pointer_x >= panel->keyboard_layout_origin_x &&
        panel->pointer_x <
            panel->keyboard_layout_origin_x + panel->keyboard_layout_width) {
        panel->pending_keyboard_layout_click = 1;
        return;
    }
    if (panel->tray_origin_x >= 0 && panel->tray_count > 0 &&
        panel->pointer_x >= panel->tray_origin_x &&
        panel->pointer_x < panel->tray_origin_x +
                           panel->tray_count * panel->tray_item_size &&
        panel->pointer_y >= panel->tray_origin_y &&
        panel->pointer_y < panel->tray_origin_y + panel->tray_item_size) {
        int index = (int)(panel->pointer_x - panel->tray_origin_x) / panel->tray_item_size;
        if (index >= 0 && index < panel->tray_count) {
            panel->pending_tray_click = index + 1;
            panel->pending_tray_button =
                button == button_right ? 3 : (button == button_middle ? 2 : 1);
            return;
        }
    }
    if (button == button_left && panel->clock_origin_x >= 0 &&
        panel->pointer_x >= panel->clock_origin_x &&
        panel->pointer_x < panel->clock_origin_x + panel->clock_width) {
        panel->pending_clock_click = 1;
        return;
    }
    if (button == button_left && panel->launcher_origin_x >= 0 &&
        panel->pointer_x >= panel->launcher_origin_x &&
        panel->pointer_x <
            panel->launcher_origin_x + panel->launcher_width) {
        panel->pending_launcher_click = 1;
        return;
    }
    if (button == button_left) {
        for (int i = 0; i < panel->button_count; i++) {
            PanelButton *module = &panel->buttons[i];
            if (module->origin_x >= 0 &&
                panel->pointer_x >= module->origin_x &&
                panel->pointer_x < module->origin_x + module->width) {
                panel->pending_button_click = i + 1;
                return;
            }
        }
    }
    if (panel->workspace_origin_x < 0 || panel->workspace_count <= 0) return;
    if (panel->pointer_y < panel->workspace_origin_y ||
        panel->pointer_y >= panel->workspace_origin_y + panel->workspace_draw_height) return;

    for (int i = 0; i < panel->workspace_count; i++) {
        int x = panel->workspace_origin_x + i * (panel->workspace_draw_width + panel->workspace_gap);
        if (panel->pointer_x >= x && panel->pointer_x < x + panel->workspace_draw_width) {
            panel->pending_workspace_click = panel->workspaces[i].number;
            return;
        }
    }
}

static void pointer_axis(void *data, struct wl_pointer *pointer, uint32_t time,
                         uint32_t axis, wl_fixed_t value) {
    (void)pointer;
    (void)time;
    (void)axis;
    HatWMPanel *panel = data;
    if (!panel) return;
    int direction = wl_fixed_to_double(value) < 0 ? 1 : -1;
    if (panel->volume_origin_x >= 0 &&
        panel->pointer_x >= panel->volume_origin_x &&
        panel->pointer_x < panel->volume_origin_x + panel->volume_width) {
        panel->pending_volume_change = clamp_int(
            panel->volume_percent + direction * panel->volume_step, 0, 100);
        panel->volume_percent = panel->pending_volume_change;
        panel_draw(panel);
        return;
    }
    if (panel->brightness_origin_x >= 0 &&
        panel->pointer_x >= panel->brightness_origin_x &&
        panel->pointer_x < panel->brightness_origin_x + panel->brightness_width) {
        panel->pending_brightness_change = clamp_int(
            panel->brightness_percent + direction * panel->brightness_step, 1, 100);
        panel->brightness_percent = panel->pending_brightness_change;
        panel_draw(panel);
        return;
    }
    if (panel->microphone_origin_x >= 0 &&
        panel->pointer_x >= panel->microphone_origin_x &&
        panel->pointer_x < panel->microphone_origin_x + panel->microphone_width) {
        panel->pending_microphone_change = clamp_int(
            panel->microphone_percent + direction * panel->microphone_step, 0, 100);
        panel->microphone_percent = panel->pending_microphone_change;
        panel_draw(panel);
    }
}

static void pointer_frame(void *data, struct wl_pointer *pointer) {
    (void)data;
    (void)pointer;
}

static void pointer_axis_source(void *data, struct wl_pointer *pointer, uint32_t axis_source) {
    (void)data;
    (void)pointer;
    (void)axis_source;
}

static void pointer_axis_stop(void *data, struct wl_pointer *pointer,
                              uint32_t time, uint32_t axis) {
    (void)data;
    (void)pointer;
    (void)time;
    (void)axis;
}

static void pointer_axis_discrete(void *data, struct wl_pointer *pointer,
                                  uint32_t axis, int32_t discrete) {
    (void)data;
    (void)pointer;
    (void)axis;
    (void)discrete;
}

static const struct wl_pointer_listener pointer_listener = {
    .enter = pointer_enter,
    .leave = pointer_leave,
    .motion = pointer_motion,
    .button = pointer_button,
    .axis = pointer_axis,
    .frame = pointer_frame,
    .axis_source = pointer_axis_source,
    .axis_stop = pointer_axis_stop,
    .axis_discrete = pointer_axis_discrete,
};

static void seat_capabilities(void *data, struct wl_seat *seat, uint32_t capabilities) {
    HatWMPanel *panel = data;
    if (!panel) return;
    int has_pointer = (capabilities & WL_SEAT_CAPABILITY_POINTER) != 0;
    if (has_pointer && !panel->pointer) {
        panel->pointer = wl_seat_get_pointer(seat);
        wl_pointer_add_listener(panel->pointer, &pointer_listener, panel);
    } else if (!has_pointer && panel->pointer) {
        cancel_slider_drag(panel);
        wl_pointer_destroy(panel->pointer);
        panel->pointer = NULL;
    }
}

static void seat_name(void *data, struct wl_seat *seat, const char *name) {
    (void)data;
    (void)seat;
    (void)name;
}

static const struct wl_seat_listener seat_listener = {
    .capabilities = seat_capabilities,
    .name = seat_name,
};

static void registry_global(void *data, struct wl_registry *registry, uint32_t name,
                            const char *interface, uint32_t version) {
    HatWMPanel *panel = data;
    if (strcmp(interface, wl_compositor_interface.name) == 0) {
        panel->compositor = wl_registry_bind(registry, name, &wl_compositor_interface, min_u32(version, 4));
    } else if (strcmp(interface, wl_shm_interface.name) == 0) {
        panel->shm = wl_registry_bind(registry, name, &wl_shm_interface, 1);
    } else if (strcmp(interface, wl_seat_interface.name) == 0) {
        panel->seat = wl_registry_bind(registry, name, &wl_seat_interface, min_u32(version, 5));
        wl_seat_add_listener(panel->seat, &seat_listener, panel);
    } else if (strcmp(interface, zwlr_layer_shell_v1_interface.name) == 0) {
        panel->layer_shell = wl_registry_bind(registry, name, &zwlr_layer_shell_v1_interface, min_u32(version, 4));
    }
}

static void registry_global_remove(void *data, struct wl_registry *registry, uint32_t name) {
    (void)data;
    (void)registry;
    (void)name;
}

static const struct wl_registry_listener registry_listener = {
    .global = registry_global,
    .global_remove = registry_global_remove,
};
