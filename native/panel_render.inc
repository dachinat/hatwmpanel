static int create_memfd(size_t size) {
#ifdef SYS_memfd_create
    int fd = (int)syscall(SYS_memfd_create, "hatwmpanel-buffer", 0x0001U);
    if (fd < 0) return -1;
    if (ftruncate(fd, (off_t)size) < 0) {
        close(fd);
        return -1;
    }
    return fd;
#else
    (void)size;
    errno = ENOSYS;
    return -1;
#endif
}

static void buffer_release(void *data, struct wl_buffer *wl_buffer) {
    PanelBuffer *buffer = data;
    if (!buffer) return;
    wl_buffer_destroy(wl_buffer);
    if (buffer->data && buffer->data != MAP_FAILED) munmap(buffer->data, buffer->size);
    free(buffer);
}

static const struct wl_buffer_listener buffer_listener = {
    .release = buffer_release,
};

static void set_source_rgba_u32(cairo_t *cr, uint32_t rgba) {
    double r = ((rgba >> 24) & 0xff) / 255.0;
    double g = ((rgba >> 16) & 0xff) / 255.0;
    double b = ((rgba >> 8) & 0xff) / 255.0;
    double a = (rgba & 0xff) / 255.0;
    cairo_set_source_rgba(cr, r, g, b, a);
}

static void color_components(
        uint32_t rgba, double *red, double *green,
        double *blue, double *alpha) {
    *red = ((rgba >> 24) & 0xff) / 255.0;
    *green = ((rgba >> 16) & 0xff) / 255.0;
    *blue = ((rgba >> 8) & 0xff) / 255.0;
    *alpha = (rgba & 0xff) / 255.0;
}

static void draw_panel_shadow(cairo_t *cr, HatWMPanel *panel) {
    int size = panel->shadow_size;
    if (size <= 0) return;
    if (size > panel->height) size = panel->height;

    double red, green, blue, alpha;
    color_components(
        panel->shadow_color, &red, &green, &blue, &alpha);
    double start_y = panel->panel_at_bottom ? 0 : panel->height - size;
    double end_y = panel->panel_at_bottom ? size : panel->height;
    cairo_pattern_t *gradient = cairo_pattern_create_linear(
        0, start_y, 0, end_y);
    if (panel->panel_at_bottom) {
        cairo_pattern_add_color_stop_rgba(
            gradient, 0, red, green, blue, alpha);
        cairo_pattern_add_color_stop_rgba(
            gradient, 1, red, green, blue, 0);
    } else {
        cairo_pattern_add_color_stop_rgba(
            gradient, 0, red, green, blue, 0);
        cairo_pattern_add_color_stop_rgba(
            gradient, 1, red, green, blue, alpha);
    }
    cairo_set_source(cr, gradient);
    cairo_rectangle(cr, 0, start_y, panel->width, size);
    cairo_fill(cr);
    cairo_pattern_destroy(gradient);
}

static void draw_panel_border(cairo_t *cr, HatWMPanel *panel) {
    int width = panel->border_width;
    if (width <= 0 || width * 2 >= panel->width ||
        width * 2 >= panel->height) return;
    set_source_rgba_u32(cr, panel->border_color);
    cairo_set_line_width(cr, width);
    double inset = width / 2.0;
    cairo_rectangle(
        cr, inset, inset,
        panel->width - width, panel->height - width);
    cairo_stroke(cr);
}

static GdkPixbuf *load_icon_pixbuf(const char *name, int size, int symbolic) {
    if (!gtk_initialized || !name || !name[0] || size <= 0) return NULL;
    GtkIconTheme *theme = gtk_icon_theme_get_default();
    if (!theme) return NULL;
    GError *error = NULL;
    GdkPixbuf *pixbuf = gtk_icon_theme_load_icon(
        theme, name, size,
        symbolic ? GTK_ICON_LOOKUP_FORCE_SYMBOLIC : GTK_ICON_LOOKUP_FORCE_SIZE,
        &error);

    // Some distributions install standalone icons outside indexed theme
    // directories (for example /usr/share/icons/archlinux.svg). GTK's theme
    // lookup cannot find those by name, so retain system-directory fallbacks.
    if (!pixbuf) {
        if (error) g_clear_error(&error);
        const char *directories[] = {
            "/usr/share/icons", "/usr/share/pixmaps", NULL};
        const char *extensions[] = {"svg", "png", NULL};
        for (int directory = 0; directories[directory] && !pixbuf;
             directory++) {
            for (int extension = 0; extensions[extension] && !pixbuf;
                 extension++) {
                char *path = g_strdup_printf(
                    "%s/%s.%s", directories[directory], name,
                    extensions[extension]);
                if (g_file_test(path, G_FILE_TEST_IS_REGULAR)) {
                    pixbuf = gdk_pixbuf_new_from_file_at_scale(
                        path, size, size, TRUE, &error);
                    if (!pixbuf && error) g_clear_error(&error);
                }
                g_free(path);
            }
        }
    }
    if (!pixbuf) {
        fprintf(stderr, "hatwmpanel: could not load system icon %s\n", name);
        return NULL;
    }
    return pixbuf;
}

static cairo_surface_t *load_system_icon(
        const char *name, int size, uint32_t color, int tint) {
    GdkPixbuf *pixbuf = load_icon_pixbuf(name, size, tint != 0);
    if (!pixbuf) return NULL;
    cairo_surface_t *surface = cairo_image_surface_create(
        CAIRO_FORMAT_ARGB32, size, size);
    cairo_t *cr = cairo_create(surface);
    cairo_set_operator(cr, CAIRO_OPERATOR_SOURCE);
    cairo_set_source_rgba(cr, 0, 0, 0, 0);
    cairo_paint(cr);
    cairo_set_operator(cr, CAIRO_OPERATOR_OVER);
    int width = gdk_pixbuf_get_width(pixbuf);
    int height = gdk_pixbuf_get_height(pixbuf);
    gdk_cairo_set_source_pixbuf(cr, pixbuf, (size - width) / 2.0,
                               (size - height) / 2.0);
    cairo_paint(cr);
    if (tint) {
        cairo_set_operator(cr, CAIRO_OPERATOR_IN);
        set_source_rgba_u32(cr, color);
        cairo_paint(cr);
    }
    cairo_destroy(cr);
    g_object_unref(pixbuf);
    return surface;
}

static void rounded_rectangle(cairo_t *cr, double x, double y,
                              double width, double height, double radius) {
    if (radius < 0) radius = 0;
    if (radius > width / 2) radius = width / 2;
    if (radius > height / 2) radius = height / 2;
    cairo_new_sub_path(cr);
    cairo_arc(cr, x + width - radius, y + radius, radius, -M_PI_2, 0);
    cairo_arc(cr, x + width - radius, y + height - radius, radius, 0, M_PI_2);
    cairo_arc(cr, x + radius, y + height - radius, radius, M_PI_2, M_PI);
    cairo_arc(cr, x + radius, y + radius, radius, M_PI, 3 * M_PI_2);
    cairo_close_path(cr);
}

static void draw_text(cairo_t *cr, const char *text, const char *font, uint32_t color,
                      double x, double y, int max_width, int align_center) {
    if (!text || text[0] == '\0' || max_width <= 0) return;

    cairo_save(cr);
    cairo_rectangle(cr, x, 0, max_width, 10000);
    cairo_clip(cr);

    PangoLayout *layout = pango_cairo_create_layout(cr);
    PangoFontDescription *desc = pango_font_description_from_string(font && font[0] ? font : "Sans 11");
    pango_layout_set_font_description(layout, desc);
    pango_layout_set_text(layout, text, -1);
    pango_layout_set_single_paragraph_mode(layout, TRUE);
    pango_layout_set_ellipsize(layout, PANGO_ELLIPSIZE_END);
    pango_layout_set_width(layout, max_width * PANGO_SCALE);
    if (align_center) pango_layout_set_alignment(layout, PANGO_ALIGN_CENTER);

    set_source_rgba_u32(cr, color);
    cairo_move_to(cr, x, y);
    pango_cairo_show_layout(cr, layout);

    pango_font_description_free(desc);
    g_object_unref(layout);
    cairo_restore(cr);
}

static int text_height(cairo_t *cr, const char *font) {
    PangoLayout *layout = pango_cairo_create_layout(cr);
    PangoFontDescription *desc = pango_font_description_from_string(font && font[0] ? font : "Sans 11");
    pango_layout_set_font_description(layout, desc);
    pango_layout_set_text(layout, "Ag", -1);
    int w = 0, h = 0;
    pango_layout_get_pixel_size(layout, &w, &h);
    pango_font_description_free(desc);
    g_object_unref(layout);
    return h > 0 ? h : 14;
}

static int text_width(cairo_t *cr, const char *text, const char *font) {
    if (!text || text[0] == '\0') return 0;
    PangoLayout *layout = pango_cairo_create_layout(cr);
    PangoFontDescription *desc = pango_font_description_from_string(font && font[0] ? font : "Sans 11");
    pango_layout_set_font_description(layout, desc);
    pango_layout_set_text(layout, text, -1);
    int w = 0, h = 0;
    pango_layout_get_pixel_size(layout, &w, &h);
    pango_font_description_free(desc);
    g_object_unref(layout);
    return w;
}

static int clamp_int(int value, int minimum, int maximum) {
    if (value < minimum) return minimum;
    if (value > maximum) return maximum;
    return value;
}

static int workspace_block_width(HatWMPanel *panel) {
    if (!panel || panel->workspace_group == HATWM_PANEL_GROUP_NONE || panel->workspace_count <= 0) return 0;
    return panel->workspace_count * panel->workspace_draw_width +
           (panel->workspace_count - 1) * panel->workspace_gap;
}

static void draw_mini_window(cairo_t *cr, HatWMPanel *panel,
                             const HatWMPanelWindow *window,
                             int inner_x, int inner_y, int inner_w, int inner_h) {
    if (!window || !window->mapped || window->width <= 0 || window->height <= 0) return;

    int output_x = panel->output.x;
    int output_y = panel->output.y;
    int output_w = panel->output.width;
    int output_h = panel->output.height;
    if (output_w <= 0 || output_h <= 0) {
        output_x = panel->output.usable_x;
        output_y = panel->output.usable_y;
        output_w = panel->output.usable_width;
        output_h = panel->output.usable_height;
    }
    if (output_w <= 0 || output_h <= 0 || inner_w <= 0 || inner_h <= 0) return;

    double nx = (double)(window->x - output_x) / (double)output_w;
    double ny = (double)(window->y - output_y) / (double)output_h;
    double nw = (double)window->width / (double)output_w;
    double nh = (double)window->height / (double)output_h;

    int x = inner_x + (int)floor(nx * inner_w);
    int y = inner_y + (int)floor(ny * inner_h);
    int w = (int)ceil(nw * inner_w);
    int h = (int)ceil(nh * inner_h);
    if (w < 2) w = 2;
    if (h < 2) h = 2;

    int right = clamp_int(x + w, inner_x, inner_x + inner_w);
    int bottom = clamp_int(y + h, inner_y, inner_y + inner_h);
    x = clamp_int(x, inner_x, inner_x + inner_w - 1);
    y = clamp_int(y, inner_y, inner_y + inner_h - 1);
    w = right - x;
    h = bottom - y;
    if (w < 1 || h < 1) return;

    set_source_rgba_u32(cr, window->focused ? panel->workspace_focused_window_color
                                            : panel->workspace_window_color);
    cairo_rectangle(cr, x, y, w, h);
    cairo_fill(cr);
}

static void draw_workspace_block(cairo_t *cr, HatWMPanel *panel, int origin_x) {
    if (!panel || panel->workspace_count <= 0) return;

    int width = panel->workspace_draw_width;
    int height = panel->workspace_draw_height;
    int origin_y = (panel->height - height) / 2;
    panel->workspace_origin_x = origin_x;
    panel->workspace_origin_y = origin_y;

    for (int i = 0; i < panel->workspace_count; i++) {
        HatWMPanelWorkspace *workspace = &panel->workspaces[i];
        int x = origin_x + i * (width + panel->workspace_gap);
        int y = origin_y;

        uint32_t workspace_color = workspace->urgent
            ? panel->workspace_urgent_color
            : (workspace->active ? panel->workspace_active_color
                                 : panel->workspace_color);
        set_source_rgba_u32(cr, workspace_color);
        cairo_rectangle(cr, x, y, width, height);
        cairo_fill(cr);

        set_source_rgba_u32(cr, panel->workspace_border_color);
        cairo_set_line_width(cr, workspace->active ? 2.0 : 1.0);
        cairo_rectangle(cr, x + 0.5, y + 0.5, width - 1.0, height - 1.0);
        cairo_stroke(cr);

        int inset = panel->workspace_inset + (workspace->active ? 1 : 0);
        int inner_x = x + inset;
        int inner_y = y + inset;
        int inner_w = width - 2 * inset;
        int inner_h = height - 2 * inset;

        // Draw ordinary windows first and the focused one last so it remains visible.
        for (int pass = 0; pass < 2; pass++) {
            for (int j = 0; j < panel->window_count; j++) {
                HatWMPanelWindow *window = &panel->windows[j];
                if (window->workspace != workspace->number) continue;
                if ((pass == 0 && window->focused) || (pass == 1 && !window->focused)) continue;
                draw_mini_window(cr, panel, window, inner_x, inner_y, inner_w, inner_h);
            }
        }
    }
}

static int tray_block_width(HatWMPanel *panel) {
    return panel->tray_count * panel->tray_item_size;
}

static void draw_tray_block(cairo_t *cr, HatWMPanel *panel, int origin_x) {
    panel->tray_origin_x = origin_x;
    panel->tray_origin_y = (panel->height - panel->tray_item_size) / 2;
    for (int i = 0; i < panel->tray_count; i++) {
        int x = origin_x + i * panel->tray_item_size;
        if (panel->tray_icons && panel->tray_icons[i]) {
            cairo_surface_t *icon = panel->tray_icons[i];
            int width = cairo_image_surface_get_width(icon);
            int height = cairo_image_surface_get_height(icon);
            double scale = (double)(panel->tray_item_size - 4) /
                           (double)(width > height ? width : height);
            double draw_width = width * scale;
            double draw_height = height * scale;
            cairo_save(cr);
            cairo_translate(cr, x + (panel->tray_item_size - draw_width) / 2.0,
                            panel->tray_origin_y + (panel->tray_item_size - draw_height) / 2.0);
            cairo_scale(cr, scale, scale);
            cairo_set_source_surface(cr, icon, 0, 0);
            cairo_pattern_set_filter(cairo_get_source(cr), CAIRO_FILTER_BILINEAR);
            cairo_paint(cr);
            cairo_restore(cr);
            continue;
        }
        /* Never replace a missing tray icon with text or a geometric glyph.
         * The tray host supplies a themed fallback icon when possible. */
    }
}

static void draw_clock_block(cairo_t *cr, HatWMPanel *panel, int origin_x,
                             int text_y, int text_width) {
    if (!panel || panel->clock_width <= 0) return;
    panel->clock_origin_x = origin_x;

    int background_y = 2;
    int background_height = panel->height - 4;
    if (background_height < 1) {
        background_y = 0;
        background_height = panel->height;
    }
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->clock_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->clock_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->clock_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size + 5;
    }
    draw_text(cr, panel->clock, panel->font, panel->foreground_color,
              content_x, text_y, text_width, 0);
}

static void draw_volume_block(cairo_t *cr, HatWMPanel *panel, int origin_x) {
    panel->volume_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->volume_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->volume_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->volume_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size + 6;
    }
    panel->volume_slider_origin_x = content_x;
    int track_y = panel->height / 2 - 2;
    set_source_rgba_u32(cr, panel->workspace_border_color);
    rounded_rectangle(cr, content_x, track_y, panel->volume_slider_width, 4, 2);
    cairo_fill(cr);

    int shown_percent = panel->volume_muted ? 0 : panel->volume_percent;
    double fill_width = panel->volume_slider_width * shown_percent / 100.0;
    if (fill_width > 0) {
        set_source_rgba_u32(cr, panel->volume_color);
        rounded_rectangle(cr, content_x, track_y, fill_width, 4, 2);
        cairo_fill(cr);
    }
    double knob_x = content_x + panel->volume_slider_width * shown_percent / 100.0;
    set_source_rgba_u32(cr, panel->volume_color);
    cairo_arc(cr, knob_x, panel->height / 2.0, 5, 0, 2 * M_PI);
    cairo_fill(cr);
}

static void draw_brightness_block(cairo_t *cr, HatWMPanel *panel, int origin_x) {
    panel->brightness_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->brightness_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->brightness_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->brightness_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size + 6;
    }
    panel->brightness_slider_origin_x = content_x;
    int track_y = panel->height / 2 - 2;
    set_source_rgba_u32(cr, panel->workspace_border_color);
    rounded_rectangle(cr, content_x, track_y,
                      panel->brightness_slider_width, 4, 2);
    cairo_fill(cr);

    double fill_width =
        panel->brightness_slider_width * panel->brightness_percent / 100.0;
    set_source_rgba_u32(cr, panel->brightness_color);
    rounded_rectangle(cr, content_x, track_y, fill_width, 4, 2);
    cairo_fill(cr);
    double knob_x = content_x +
        panel->brightness_slider_width * panel->brightness_percent / 100.0;
    cairo_arc(cr, knob_x, panel->height / 2.0, 5, 0, 2 * M_PI);
    cairo_fill(cr);
}

static void draw_microphone_block(cairo_t *cr, HatWMPanel *panel, int origin_x) {
    panel->microphone_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->microphone_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->microphone_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->microphone_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size + 6;
    }
    panel->microphone_slider_origin_x = content_x;
    int track_y = panel->height / 2 - 2;
    set_source_rgba_u32(cr, panel->workspace_border_color);
    rounded_rectangle(cr, content_x, track_y,
                      panel->microphone_slider_width, 4, 2);
    cairo_fill(cr);

    int shown_percent = panel->microphone_muted ? 0 : panel->microphone_percent;
    double fill_width =
        panel->microphone_slider_width * shown_percent / 100.0;
    if (fill_width > 0) {
        set_source_rgba_u32(cr, panel->microphone_color);
        rounded_rectangle(cr, content_x, track_y, fill_width, 4, 2);
        cairo_fill(cr);
    }
    double knob_x = content_x +
        panel->microphone_slider_width * shown_percent / 100.0;
    set_source_rgba_u32(cr, panel->microphone_color);
    cairo_arc(cr, knob_x, panel->height / 2.0, 5, 0, 2 * M_PI);
    cairo_fill(cr);
}

static void draw_network_block(cairo_t *cr, HatWMPanel *panel, int origin_x,
                               int text_y) {
    panel->network_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->network_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    cairo_surface_t *network_icon =
        panel->network_active_wired && panel->network_wired_icon
        ? panel->network_wired_icon : panel->network_icon;
    if (network_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, network_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size;
        if (panel->network_show_text) content_x += 5;
    }
    if (panel->network_show_text) {
        int text_w = text_width(cr, panel->network_label, panel->font);
        draw_text(cr, panel->network_label, panel->font,
                  panel->foreground_color, content_x, text_y, text_w, 0);
    }
}

static void draw_bluetooth_block(cairo_t *cr, HatWMPanel *panel, int origin_x,
                                 int text_y) {
    panel->bluetooth_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->bluetooth_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->bluetooth_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->bluetooth_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size;
        if (panel->bluetooth_show_text) content_x += 5;
    }
    if (panel->bluetooth_show_text) {
        int text_w = text_width(cr, panel->bluetooth_label, panel->font);
        draw_text(cr, panel->bluetooth_label, panel->font,
                  panel->foreground_color, content_x, text_y, text_w, 0);
    }
}

static void draw_battery_block(cairo_t *cr, HatWMPanel *panel, int origin_x,
                               int text_y) {
    panel->battery_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->battery_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->battery_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->battery_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size;
        if (panel->battery_show_text) content_x += 5;
    }
    if (panel->battery_show_text) {
        int text_w = text_width(cr, panel->battery_label, panel->font);
        draw_text(cr, panel->battery_label, panel->font,
                  panel->foreground_color, content_x, text_y, text_w, 0);
    }
}

static void draw_keyboard_layout_block(cairo_t *cr, HatWMPanel *panel,
                                       int origin_x, int text_y) {
    panel->keyboard_layout_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->keyboard_layout_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->keyboard_layout_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(
            cr, panel->keyboard_layout_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size + 5;
    }
    int text_w = text_width(cr, panel->keyboard_layout, panel->font);
    draw_text(cr, panel->keyboard_layout, panel->font,
              panel->foreground_color, content_x, text_y, text_w, 0);
}

static void draw_storage_block(cairo_t *cr, HatWMPanel *panel, int origin_x,
                               int text_y) {
    panel->storage_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->storage_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->storage_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->storage_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size + 5;
    }
    int text_w = text_width(cr, panel->storage_text, panel->font);
    draw_text(cr, panel->storage_text, panel->font,
              panel->foreground_color, content_x, text_y, text_w, 0);
}

static void draw_ram_block(cairo_t *cr, HatWMPanel *panel, int origin_x,
                           int text_y) {
    panel->ram_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->ram_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->ram_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->ram_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size + 5;
    }
    int text_w = text_width(cr, panel->ram_text, panel->font);
    draw_text(cr, panel->ram_text, panel->font,
              panel->foreground_color, content_x, text_y, text_w, 0);
}

static void draw_cpu_block(cairo_t *cr, HatWMPanel *panel, int origin_x,
                           int text_y) {
    panel->cpu_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->cpu_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->cpu_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->cpu_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size + 5;
    }
    int text_w = text_width(cr, panel->cpu_text, panel->font);
    draw_text(cr, panel->cpu_text, panel->font,
              panel->foreground_color, content_x, text_y, text_w, 0);
}

static void draw_gpu_block(cairo_t *cr, HatWMPanel *panel, int origin_x,
                           int text_y) {
    panel->gpu_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->gpu_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->gpu_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->gpu_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size + 5;
    }
    int text_w = text_width(cr, panel->gpu_text, panel->font);
    draw_text(cr, panel->gpu_text, panel->font,
              panel->foreground_color, content_x, text_y, text_w, 0);
}

static void draw_netstat_block(cairo_t *cr, HatWMPanel *panel, int origin_x,
                               int text_y) {
    panel->netstat_origin_x = origin_x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, origin_x, background_y, panel->netstat_width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = origin_x + panel->module_padding;
    if (panel->netstat_icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, panel->netstat_icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size + 5;
    }
    int text_w = text_width(cr, panel->netstat_text, panel->font);
    draw_text(cr, panel->netstat_text, panel->font,
              panel->foreground_color, content_x, text_y, text_w, 0);
}

static int panel_button_width(cairo_t *cr, HatWMPanel *panel, PanelButton *button) {
    int text_w = text_width(cr, button->text, panel->font);
    int content_w = text_w;
    if (button->icon) {
        content_w += panel->module_icon_size;
        if (text_w > 0) content_w += 5;
    }
    return content_w > 0 ? content_w + panel->module_padding * 2 : 0;
}

static void draw_panel_button(cairo_t *cr, HatWMPanel *panel, int index,
                              int x, int text_y) {
    PanelButton *button = &panel->buttons[index];
    button->origin_x = x;
    int background_y = 2;
    int background_height = panel->height - 4;
    set_source_rgba_u32(cr, panel->module_background_color);
    rounded_rectangle(cr, x, background_y, button->width,
                      background_height, panel->module_radius);
    cairo_fill(cr);

    int content_x = x + panel->module_padding;
    if (button->icon) {
        int icon_y = (panel->height - panel->module_icon_size) / 2;
        cairo_set_source_surface(cr, button->icon, content_x, icon_y);
        cairo_paint(cr);
        content_x += panel->module_icon_size;
        if (button->text && button->text[0]) content_x += 5;
    }
    if (button->text && button->text[0]) {
        int text_w = text_width(cr, button->text, panel->font);
        draw_text(cr, button->text, panel->font, panel->foreground_color,
                  content_x, text_y, text_w, 0);
    }
}

static void draw_launcher_block(cairo_t *cr, HatWMPanel *panel, int origin_x) {
    panel->launcher_origin_x = origin_x;
    if (!panel->launcher_logo || panel->launcher_width <= 0) return;

    int source_width =
        cairo_image_surface_get_width(panel->launcher_logo);
    int source_height =
        cairo_image_surface_get_height(panel->launcher_logo);
    if (source_width <= 0 || source_height <= 0) return;

    double scale_x = (double)panel->launcher_size / source_width;
    double scale_y = (double)panel->launcher_size / source_height;
    double scale = scale_x < scale_y ? scale_x : scale_y;
    double width = source_width * scale;
    double height = source_height * scale;
    double x = origin_x + (panel->launcher_width - width) / 2.0;
    double y = (panel->height - height) / 2.0;

    cairo_save(cr);
    cairo_translate(cr, x, y);
    cairo_scale(cr, scale, scale);
    cairo_set_source_surface(cr, panel->launcher_logo, 0, 0);
    // Launcher artwork can contain much more detail than symbolic module
    // icons. Use Cairo's highest-quality downsampling to avoid jagged edges
    // and lost highlights at panel size.
    cairo_pattern_set_filter(cairo_get_source(cr), CAIRO_FILTER_BEST);
    cairo_paint(cr);
    cairo_restore(cr);
}

typedef struct LayoutComponent {
    int kind;
    int width;
    int order;
    int button_index;
} LayoutComponent;

static int component_gap(const LayoutComponent *left,
                         const LayoutComponent *right, int normal_gap) {
    if (!left || !right ||
        left->kind == PANEL_COMPONENT_SEPARATORS ||
        right->kind == PANEL_COMPONENT_SEPARATORS)
        return 0;
    return normal_gap;
}

static void draw_panel_group(cairo_t *cr, HatWMPanel *panel, int group,
                             int area_start, int area_end, int alignment,
                             int text_y, int module_gap, int workspace_width,
                             int tray_width, int clock_width,
    int clock_text_width) {
    const char *text = group == HATWM_PANEL_GROUP_LEFT ? panel->left :
                       group == HATWM_PANEL_GROUP_CENTER ? panel->center :
                       panel->right;
    int text_w = text_width(cr, text, panel->font);
    LayoutComponent components[PANEL_COMPONENT_COUNT + 256];
    int count = 0;
#define ADD_COMPONENT(component_kind, component_width) do { \
    if ((component_width) > 0) { \
        components[count++] = (LayoutComponent){ \
            (component_kind), (component_width), \
            panel->group_order[group][component_kind], -1}; \
    } \
} while (0)
    if (panel->launcher_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_LAUNCHER, panel->launcher_width);
    if (panel->workspace_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_WORKSPACES, workspace_width);
    if (panel->tray_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_TRAY, tray_width);
    if (panel->clock_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_CLOCK, clock_width);
    if (panel->volume_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_VOLUME, panel->volume_width);
    if (panel->brightness_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_BRIGHTNESS, panel->brightness_width);
    if (panel->microphone_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_MICROPHONE, panel->microphone_width);
    if (panel->network_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_NETWORK, panel->network_width);
    if (panel->bluetooth_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_BLUETOOTH, panel->bluetooth_width);
    if (panel->battery_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_BATTERY, panel->battery_width);
    if (panel->keyboard_layout_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_KEYBOARD_LAYOUT,
                      panel->keyboard_layout_width);
    if (panel->storage_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_STORAGE, panel->storage_width);
    if (panel->ram_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_RAM, panel->ram_width);
    if (panel->cpu_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_CPU, panel->cpu_width);
    if (panel->gpu_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_GPU, panel->gpu_width);
    if (panel->netstat_group == group)
        ADD_COMPONENT(PANEL_COMPONENT_NETSTAT, panel->netstat_width);
    for (int i = 0; i < panel->separator_count; i++) {
        PanelSeparator *separator = &panel->separators[i];
        if (separator->group != group || separator->width <= 0) continue;
        components[count++] = (LayoutComponent){
            PANEL_COMPONENT_SEPARATORS, separator->width,
            separator->order, -1};
    }
    for (int i = 0; i < panel->button_count; i++) {
        PanelButton *button = &panel->buttons[i];
        if (button->group != group) continue;
        button->width = panel_button_width(cr, panel, button);
        if (button->width <= 0) continue;
        components[count++] = (LayoutComponent){
            PANEL_COMPONENT_BUTTONS, button->width, button->order, i};
    }
    ADD_COMPONENT(PANEL_COMPONENT_TEXT, text_w);
#undef ADD_COMPONENT

    for (int i = 1; i < count; i++) {
        LayoutComponent current = components[i];
        int j = i - 1;
        while (j >= 0 && components[j].order > current.order) {
            components[j + 1] = components[j];
            j--;
        }
        components[j + 1] = current;
    }

    int total = 0;
    for (int i = 0; i < count; i++) {
        total += components[i].width;
        if (i > 0)
            total += component_gap(
                &components[i - 1], &components[i], module_gap);
    }
    int available = area_end - area_start;
    int x = area_start;
    if (alignment == 1)
        x = area_start + (available - total) / 2;
    else if (alignment == 2)
        x = area_end - total;
    /*
     * Keep the right group anchored to the monitor's right padding and let it
     * grow left into otherwise unused center space. Clamping it to area_start
     * made an oversized right group begin at the two-thirds boundary and
     * extend beyond the monitor instead.
     */
    int minimum_x = alignment == 2 ? panel->padding : area_start;
    if (x < minimum_x) x = minimum_x;

    for (int i = 0; i < count; i++) {
        LayoutComponent *component = &components[i];
        switch (component->kind) {
        case PANEL_COMPONENT_LAUNCHER:
            draw_launcher_block(cr, panel, x);
            break;
        case PANEL_COMPONENT_WORKSPACES:
            draw_workspace_block(cr, panel, x);
            break;
        case PANEL_COMPONENT_TRAY:
            draw_tray_block(cr, panel, x);
            break;
        case PANEL_COMPONENT_CLOCK:
            draw_clock_block(cr, panel, x, text_y, clock_text_width);
            break;
        case PANEL_COMPONENT_VOLUME:
            draw_volume_block(cr, panel, x);
            break;
        case PANEL_COMPONENT_BRIGHTNESS:
            draw_brightness_block(cr, panel, x);
            break;
        case PANEL_COMPONENT_MICROPHONE:
            draw_microphone_block(cr, panel, x);
            break;
        case PANEL_COMPONENT_NETWORK:
            draw_network_block(cr, panel, x, text_y);
            break;
        case PANEL_COMPONENT_BLUETOOTH:
            draw_bluetooth_block(cr, panel, x, text_y);
            break;
        case PANEL_COMPONENT_BATTERY:
            draw_battery_block(cr, panel, x, text_y);
            break;
        case PANEL_COMPONENT_KEYBOARD_LAYOUT:
            draw_keyboard_layout_block(cr, panel, x, text_y);
            break;
        case PANEL_COMPONENT_STORAGE:
            draw_storage_block(cr, panel, x, text_y);
            break;
        case PANEL_COMPONENT_RAM:
            draw_ram_block(cr, panel, x, text_y);
            break;
        case PANEL_COMPONENT_CPU:
            draw_cpu_block(cr, panel, x, text_y);
            break;
        case PANEL_COMPONENT_GPU:
            draw_gpu_block(cr, panel, x, text_y);
            break;
        case PANEL_COMPONENT_NETSTAT:
            draw_netstat_block(cr, panel, x, text_y);
            break;
        case PANEL_COMPONENT_SEPARATORS:
            break;
        case PANEL_COMPONENT_BUTTONS:
            draw_panel_button(cr, panel, component->button_index, x, text_y);
            break;
        case PANEL_COMPONENT_TEXT:
            draw_text(cr, text, panel->font, panel->foreground_color,
                      x, text_y, component->width, 0);
            break;
        }
        x += component->width;
        if (i + 1 < count)
            x += component_gap(component, &components[i + 1], module_gap);
    }
}

static int panel_draw(HatWMPanel *panel) {
    if (!panel || !panel->configured || panel->closed || panel->width <= 0 || panel->height <= 0) return 0;

    int stride = cairo_format_stride_for_width(CAIRO_FORMAT_ARGB32, panel->width);
    size_t size = (size_t)stride * (size_t)panel->height;
    int fd = create_memfd(size);
    if (fd < 0) {
        set_errno_error("memfd_create failed");
        return -1;
    }

    void *data = mmap(NULL, size, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    if (data == MAP_FAILED) {
        set_errno_error("mmap failed");
        close(fd);
        return -1;
    }

    struct wl_shm_pool *pool = wl_shm_create_pool(panel->shm, fd, (int)size);
    struct wl_buffer *wl_buffer = wl_shm_pool_create_buffer(pool, 0, panel->width, panel->height,
                                                            stride, WL_SHM_FORMAT_ARGB8888);
    wl_shm_pool_destroy(pool);
    close(fd);

    if (!wl_buffer) {
        munmap(data, size);
        set_error("wl_shm_pool_create_buffer failed");
        return -1;
    }

    PanelBuffer *buffer = calloc(1, sizeof(*buffer));
    if (!buffer) {
        wl_buffer_destroy(wl_buffer);
        munmap(data, size);
        set_error("out of memory");
        return -1;
    }
    buffer->buffer = wl_buffer;
    buffer->data = data;
    buffer->size = size;
    wl_buffer_add_listener(wl_buffer, &buffer_listener, buffer);

    cairo_surface_t *surface = cairo_image_surface_create_for_data((unsigned char *)data, CAIRO_FORMAT_ARGB32,
                                                                   panel->width, panel->height, stride);
    cairo_t *cr = cairo_create(surface);

    cairo_set_source_rgba(cr, 0, 0, 0, 0);
    cairo_set_operator(cr, CAIRO_OPERATOR_SOURCE);
    cairo_paint(cr);
    cairo_set_operator(cr, CAIRO_OPERATOR_OVER);
    cairo_push_group(cr);

    set_source_rgba_u32(cr, panel->background_color);
    cairo_set_operator(cr, CAIRO_OPERATOR_SOURCE);
    cairo_paint(cr);
    cairo_set_operator(cr, CAIRO_OPERATOR_OVER);

    panel->workspace_origin_x = -1;
    panel->workspace_origin_y = -1;
    panel->tray_origin_x = -1;
    panel->clock_origin_x = -1;
    panel->volume_origin_x = -1;
    panel->brightness_origin_x = -1;
    panel->microphone_origin_x = -1;
    panel->network_origin_x = -1;
    panel->bluetooth_origin_x = -1;
    panel->battery_origin_x = -1;
    panel->keyboard_layout_origin_x = -1;
    panel->storage_origin_x = -1;
    panel->ram_origin_x = -1;
    panel->cpu_origin_x = -1;
    panel->gpu_origin_x = -1;
    panel->netstat_origin_x = -1;
    panel->launcher_origin_x = -1;
    for (int i = 0; i < panel->button_count; i++)
        panel->buttons[i].origin_x = -1;
    panel->launcher_width =
        panel->launcher_group != HATWM_PANEL_GROUP_NONE &&
        panel->launcher_logo ? panel->launcher_size : 0;
    int block_width = workspace_block_width(panel);
    int tray_width = tray_block_width(panel);
    int module_gap = panel->padding / 2;
    if (module_gap < 4) module_gap = 4;

    int th = text_height(cr, panel->font);
    double text_y = (panel->height - th) / 2.0;
    int clock_text_width = text_width(cr, panel->clock, panel->font);
    int clock_width = 0;
    if (clock_text_width > 0) {
        clock_width = clock_text_width + panel->module_padding * 2;
        if (panel->clock_icon)
            clock_width += panel->module_icon_size + 5;
    }
    panel->clock_width = clock_width;
    panel->volume_width = panel->module_padding * 2 + panel->volume_slider_width;
    if (panel->volume_icon)
        panel->volume_width += panel->module_icon_size + 6;
    panel->brightness_width =
        panel->module_padding * 2 + panel->brightness_slider_width;
    if (panel->brightness_icon)
        panel->brightness_width += panel->module_icon_size + 6;
    panel->microphone_width =
        panel->module_padding * 2 + panel->microphone_slider_width;
    if (panel->microphone_icon)
        panel->microphone_width += panel->module_icon_size + 6;
    int network_text_width = panel->network_show_text
        ? text_width(cr, panel->network_label, panel->font) : 0;
    panel->network_width = panel->module_padding * 2 + network_text_width;
    cairo_surface_t *network_icon =
        panel->network_active_wired && panel->network_wired_icon
        ? panel->network_wired_icon : panel->network_icon;
    if (network_icon) {
        panel->network_width += panel->module_icon_size;
        if (network_text_width > 0) panel->network_width += 5;
    }
    int bluetooth_text_width = panel->bluetooth_show_text
        ? text_width(cr, panel->bluetooth_label, panel->font) : 0;
    panel->bluetooth_width = panel->module_padding * 2 + bluetooth_text_width;
    if (panel->bluetooth_icon) {
        panel->bluetooth_width += panel->module_icon_size;
        if (bluetooth_text_width > 0) panel->bluetooth_width += 5;
    }
    int battery_text_width = panel->battery_show_text
        ? text_width(cr, panel->battery_label, panel->font) : 0;
    panel->battery_width = panel->module_padding * 2 + battery_text_width;
    if (panel->battery_icon) {
        panel->battery_width += panel->module_icon_size;
        if (battery_text_width > 0) panel->battery_width += 5;
    }
    int keyboard_layout_text_width =
        text_width(cr, panel->keyboard_layout, panel->font);
    panel->keyboard_layout_width =
        panel->module_padding * 2 + keyboard_layout_text_width;
    if (panel->keyboard_layout_icon)
        panel->keyboard_layout_width += panel->module_icon_size + 5;
    int storage_text_width =
        text_width(cr, panel->storage_text, panel->font);
    panel->storage_width = panel->module_padding * 2 + storage_text_width;
    if (panel->storage_icon)
        panel->storage_width += panel->module_icon_size + 5;
    int ram_text_width = text_width(cr, panel->ram_text, panel->font);
    panel->ram_width = panel->module_padding * 2 + ram_text_width;
    if (panel->ram_icon)
        panel->ram_width += panel->module_icon_size + 5;
    int cpu_text_width = text_width(cr, panel->cpu_text, panel->font);
    panel->cpu_width = panel->module_padding * 2 + cpu_text_width;
    if (panel->cpu_icon)
        panel->cpu_width += panel->module_icon_size + 5;
    int gpu_text_width = text_width(cr, panel->gpu_text, panel->font);
    panel->gpu_width = panel->module_padding * 2 + gpu_text_width;
    if (panel->gpu_icon)
        panel->gpu_width += panel->module_icon_size + 5;
    int netstat_text_width =
        text_width(cr, panel->netstat_text, panel->font);
    panel->netstat_width =
        panel->module_padding * 2 + netstat_text_width;
    if (panel->netstat_icon)
        panel->netstat_width += panel->module_icon_size + 5;
    int third = panel->width / 3;
    int padding = panel->padding;
    draw_panel_group(cr, panel, HATWM_PANEL_GROUP_LEFT,
                     padding, third, 0, text_y, module_gap,
                     block_width, tray_width, clock_width, clock_text_width);
    draw_panel_group(cr, panel, HATWM_PANEL_GROUP_CENTER,
                     third, third * 2, 1, text_y, module_gap,
                     block_width, tray_width, clock_width, clock_text_width);
    draw_panel_group(cr, panel, HATWM_PANEL_GROUP_RIGHT,
                     third * 2, panel->width - padding, 2, text_y, module_gap,
                     block_width, tray_width, clock_width, clock_text_width);

    draw_panel_shadow(cr, panel);
    draw_panel_border(cr, panel);
    cairo_pop_group_to_source(cr);
    cairo_set_operator(cr, CAIRO_OPERATOR_OVER);
    cairo_paint_with_alpha(cr, panel->panel_opacity);

    cairo_destroy(cr);
    cairo_surface_flush(surface);
    cairo_surface_destroy(surface);

    wl_surface_attach(panel->surface, wl_buffer, 0, 0);
    wl_surface_damage(panel->surface, 0, 0, panel->width, panel->height);
    wl_surface_commit(panel->surface);
    wl_display_flush(panel->display);
    return 0;
}

static void layer_configure(void *data, struct zwlr_layer_surface_v1 *surface,
                            uint32_t serial, uint32_t width, uint32_t height) {
    HatWMPanel *panel = data;
    zwlr_layer_surface_v1_ack_configure(surface, serial);
    panel->width = width > 0 ? (int)width : 1;
    panel->height = height > 0 ? (int)height : panel->requested_height;
    panel->workspace_draw_height = panel->workspace_height;
    if (panel->workspace_draw_height > panel->height - 2) panel->workspace_draw_height = panel->height - 2;
    if (panel->workspace_draw_height < 4) panel->workspace_draw_height = 4;
    panel->workspace_draw_width = panel->workspace_width;
    if (panel->workspace_draw_width < 4) panel->workspace_draw_width = 4;
    panel->configured = 1;
    panel_draw(panel);
}

static void layer_closed(void *data, struct zwlr_layer_surface_v1 *surface) {
    (void)surface;
    HatWMPanel *panel = data;
    panel->closed = 1;
}

static const struct zwlr_layer_surface_v1_listener layer_surface_listener = {
    .configure = layer_configure,
    .closed = layer_closed,
};
