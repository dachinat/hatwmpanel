static void format_gtk_css_color(char output[32], uint32_t rgba) {
    unsigned int red = (rgba >> 24) & 0xff;
    unsigned int green = (rgba >> 16) & 0xff;
    unsigned int blue = (rgba >> 8) & 0xff;
    double alpha = (rgba & 0xff) / 255.0;
    snprintf(output, 32, "rgba(%u,%u,%u,%.3f)", red, green, blue, alpha);
}

static void quote_gtk_css_string(char *output, size_t size, const char *value) {
    if (!output || size < 3) return;
    size_t written = 0;
    output[written++] = '"';
    for (const char *p = value ? value : "Sans"; *p && written + 2 < size; p++) {
        if (*p == '\\' || *p == '"') output[written++] = '\\';
        output[written++] = *p;
    }
    output[written++] = '"';
    output[written] = '\0';
}

static void configure_menu_style(const HatWMPanelConfig *config) {
    if (!gtk_initialized) {
        if (!gtk_init_check(NULL, NULL)) return;
        gtk_initialized = 1;
    }
    PangoFontDescription *font = pango_font_description_from_string(
        config->menu_font ? config->menu_font : "Sans 11");
    char font_family[256];
    quote_gtk_css_string(font_family, sizeof(font_family),
                         pango_font_description_get_family(font));
    int font_size = pango_font_description_get_size(font);
    double font_points = font_size > 0 ? (double)font_size / PANGO_SCALE : 11.0;
    int font_weight = pango_font_description_get_weight(font);
    const char *font_style = "normal";
    if (pango_font_description_get_style(font) == PANGO_STYLE_ITALIC)
        font_style = "italic";
    else if (pango_font_description_get_style(font) == PANGO_STYLE_OBLIQUE)
        font_style = "oblique";
    pango_font_description_free(font);

    char background[32], foreground[32], selected_background[32];
    char selected_foreground[32], border[32];
    format_gtk_css_color(background, config->menu_background_color);
    format_gtk_css_color(foreground, config->menu_foreground_color);
    format_gtk_css_color(selected_background, config->menu_selected_background_color);
    format_gtk_css_color(selected_foreground, config->menu_selected_foreground_color);
    format_gtk_css_color(border, config->menu_border_color);

    char css[8192];
    snprintf(css, sizeof(css),
        "menu, .menu {"
        " background-color: %s; color: %s;"
        " border: %dpx solid %s; border-radius: %dpx;"
        " font-family: %s; font-size: %.3fpt; font-weight: %d; font-style: %s;"
        " padding: %dpx;"
        "}"
        "menuitem, .menuitem { color: %s; padding: %dpx %dpx; border-radius: %dpx; }"
        "menuitem:hover, menuitem:focus, .menuitem:hover, .menuitem:focus {"
        " background-color: %s; color: %s;"
        "}"
        "window#hatwm-calendar-window {"
        " background-color: %s; color: %s;"
        " border: %dpx solid %s; border-radius: %dpx; padding: 0;"
        "}"
        "window#hatwm-calendar-window calendar {"
        " background-color: transparent; border: 0; border-radius: 0;"
        " box-shadow: none; margin: 0; padding: 0; outline: 0;"
        " font-family: %s; font-size: %.3fpt; font-weight: %d; font-style: %s;"
        "}"
        "window#hatwm-calendar-window calendar:hover,"
        "window#hatwm-calendar-window calendar.view:hover {"
        " background-color: transparent; box-shadow: none;"
        "}"
        "window#hatwm-calendar-window calendar > grid > label.today,"
        "window#hatwm-calendar-window calendar > grid > label.today:selected,"
        "window#hatwm-calendar-window calendar > grid > label.day-number:checked {"
        " background-color: %s; color: %s; border-radius: 4px;"
        "}"
        "window#hatwm-calendar-window calendar:selected,"
        "window#hatwm-calendar-window calendar:backdrop:selected {"
        " background-color: %s; color: %s; border-radius: 4px;"
        "}"
        "window#hatwm-launcher-window {"
        " background-color: %s; color: %s;"
        " border: %dpx solid %s; border-radius: %dpx;"
        " font-family: %s; font-size: %.3fpt; font-weight: %d; font-style: %s;"
        "}"
        "#hatwm-launcher-box { padding: %dpx; }"
        "#hatwm-launcher-entry {"
        " background-color: %s; color: %s; border: 1px solid %s;"
        " border-radius: %dpx; padding: %dpx;"
        "}"
        "#hatwm-launcher-results, #hatwm-launcher-results row {"
        " background-color: transparent; color: %s;"
        "}"
        "#hatwm-launcher-results row {"
        " padding: %dpx; border-radius: %dpx;"
        "}"
        "#hatwm-launcher-results row:hover,"
        "#hatwm-launcher-results row:selected {"
        " background-color: %s; color: %s;"
        "}"
        "menu separator { background-color: %s; min-height: 1px; margin: 3px %dpx; }",
        background, foreground,
        config->menu_border_width, border, config->menu_radius,
        font_family, font_points, font_weight, font_style,
        config->menu_padding,
        foreground, config->menu_padding,
        config->menu_padding * 2, config->menu_radius > 2 ? config->menu_radius - 2 : 0,
        selected_background, selected_foreground,
        background, foreground,
        config->menu_border_width, border, config->menu_radius,
        font_family, font_points, font_weight, font_style,
        selected_background, selected_foreground,
        selected_background, selected_foreground,
        background, foreground,
        config->menu_border_width, border, config->menu_radius,
        font_family, font_points, font_weight, font_style,
        config->menu_padding,
        background, foreground, border,
        config->menu_radius > 2 ? config->menu_radius - 2 : 0,
        config->menu_padding,
        foreground,
        config->menu_padding,
        config->menu_radius > 2 ? config->menu_radius - 2 : 0,
        selected_background, selected_foreground,
        border, config->menu_padding);
    GtkCssProvider *provider = gtk_css_provider_new();
    GError *error = NULL;
    if (gtk_css_provider_load_from_data(provider, css, -1, &error)) {
        GdkScreen *screen = gdk_screen_get_default();
        if (tray_menu_style) {
            gtk_style_context_remove_provider_for_screen(
                screen, GTK_STYLE_PROVIDER(tray_menu_style));
            g_object_unref(tray_menu_style);
        }
        tray_menu_style = provider;
        gtk_style_context_add_provider_for_screen(
            screen, GTK_STYLE_PROVIDER(tray_menu_style),
            GTK_STYLE_PROVIDER_PRIORITY_APPLICATION);
    } else {
        g_object_unref(provider);
    }
    if (error) {
        fprintf(stderr, "hatwmpanel: invalid tray menu CSS: %s\n", error->message);
        g_error_free(error);
    }
}

static void set_error(const char *message) {
    if (!message) message = "unknown error";
    snprintf(last_error, sizeof(last_error), "%s", message);
}

static void set_errno_error(const char *prefix) {
    snprintf(last_error, sizeof(last_error), "%s: %s", prefix, strerror(errno));
}

const char *hatwm_panel_last_error(void) {
    return last_error;
}

