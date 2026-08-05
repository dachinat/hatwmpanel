static void calendar_window_destroyed(GtkWidget *widget, gpointer data) {
    HatWMPanel *panel = data;
    if (panel && panel->calendar_window == widget) {
        panel->calendar_window = NULL;
        panel->calendar_left_margin = -1;
    }
}

static void center_calendar_window(GtkWidget *widget, GtkAllocation *allocation,
                                   gpointer data) {
    HatWMPanel *panel = data;
    if (!panel || panel->calendar_window != widget || allocation->width <= 0)
        return;
    int clock_center = panel->clock_origin_x + panel->clock_width / 2;
    int left_margin = clock_center - allocation->width / 2;
    if (left_margin < 0) left_margin = 0;
    if (left_margin + allocation->width > panel->width)
        left_margin = panel->width - allocation->width;
    if (left_margin < 0) left_margin = 0;
    if (left_margin == panel->calendar_left_margin) return;
    panel->calendar_left_margin = left_margin;
    gtk_layer_set_margin(GTK_WINDOW(widget), GTK_LAYER_SHELL_EDGE_LEFT, left_margin);
}

static void select_today_in_current_month(GtkCalendar *calendar, gpointer data) {
    (void)data;
    guint year = 0, month = 0, day = 0;
    gtk_calendar_get_date(calendar, &year, &month, &day);

    GDateTime *today = g_date_time_new_now_local();
    if (!today) return;
    if ((int)year == g_date_time_get_year(today) &&
        (int)month == g_date_time_get_month(today) - 1) {
        gtk_calendar_select_day(calendar, g_date_time_get_day_of_month(today));
    } else {
        gtk_calendar_select_day(calendar, 0);
    }
    g_date_time_unref(today);
}

int hatwm_panel_show_calendar(HatWMPanel *panel) {
    if (!panel || panel->clock_origin_x < 0) return 0;
    if (panel->calendar_window) {
        gtk_widget_destroy(panel->calendar_window);
        return 1;
    }
    if (!gtk_initialized) {
        if (!gtk_init_check(NULL, NULL)) return 0;
        gtk_initialized = 1;
    }

    GtkWidget *window = gtk_window_new(GTK_WINDOW_TOPLEVEL);
    panel->calendar_window = window;
    panel->calendar_left_margin = -1;
    gtk_widget_set_name(window, "hatwm-calendar-window");
    gtk_window_set_decorated(GTK_WINDOW(window), FALSE);
    gtk_window_set_resizable(GTK_WINDOW(window), FALSE);
    gtk_window_set_skip_taskbar_hint(GTK_WINDOW(window), TRUE);
    gtk_layer_init_for_window(GTK_WINDOW(window));
    gtk_layer_set_namespace(GTK_WINDOW(window), "hatwmpanel-calendar");
    gtk_layer_set_layer(GTK_WINDOW(window), GTK_LAYER_SHELL_LAYER_OVERLAY);
    gtk_layer_set_keyboard_mode(GTK_WINDOW(window), GTK_LAYER_SHELL_KEYBOARD_MODE_ON_DEMAND);

    if (panel->panel_at_bottom) {
        gtk_layer_set_anchor(GTK_WINDOW(window), GTK_LAYER_SHELL_EDGE_BOTTOM, TRUE);
    } else {
        gtk_layer_set_anchor(GTK_WINDOW(window), GTK_LAYER_SHELL_EDGE_TOP, TRUE);
    }

    GtkWidget *calendar = gtk_calendar_new();
    gtk_widget_set_halign(calendar, GTK_ALIGN_START);
    gtk_widget_set_valign(calendar, GTK_ALIGN_START);
    gtk_calendar_set_display_options(GTK_CALENDAR(calendar),
        GTK_CALENDAR_SHOW_HEADING | GTK_CALENDAR_SHOW_DAY_NAMES);
    GDateTime *today = g_date_time_new_now_local();
    if (today) {
        gtk_calendar_select_month(GTK_CALENDAR(calendar),
                                  g_date_time_get_month(today) - 1,
                                  g_date_time_get_year(today));
        g_date_time_unref(today);
    }
    g_signal_connect(calendar, "month-changed",
                     G_CALLBACK(select_today_in_current_month), NULL);
    select_today_in_current_month(GTK_CALENDAR(calendar), NULL);
    gtk_container_add(GTK_CONTAINER(window), calendar);

    int clock_center = panel->clock_origin_x + panel->clock_width / 2;
    gtk_layer_set_anchor(GTK_WINDOW(window), GTK_LAYER_SHELL_EDGE_LEFT, TRUE);
    gtk_layer_set_margin(GTK_WINDOW(window), GTK_LAYER_SHELL_EDGE_LEFT, clock_center);

    g_signal_connect(window, "destroy",
                     G_CALLBACK(calendar_window_destroyed), panel);
    g_signal_connect(window, "size-allocate",
                     G_CALLBACK(center_calendar_window), panel);
    gtk_widget_show_all(window);
    gtk_window_present(GTK_WINDOW(window));
    gtk_widget_grab_focus(calendar);
    return 1;
}
