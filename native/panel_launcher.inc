static gint compare_launcher_apps(gconstpointer left, gconstpointer right) {
    const char *left_name =
        g_app_info_get_display_name(G_APP_INFO((gpointer)left));
    const char *right_name =
        g_app_info_get_display_name(G_APP_INFO((gpointer)right));
    return g_utf8_collate(left_name ? left_name : "",
                          right_name ? right_name : "");
}

static void destroy_launcher_result(GtkWidget *widget, gpointer data) {
    (void)data;
    gtk_widget_destroy(widget);
}

static int launcher_app_matches(GAppInfo *app, const char *query_folded) {
    if (!app || !query_folded || !query_folded[0] ||
        !g_app_info_should_show(app))
        return 0;
    const char *fields[] = {
        g_app_info_get_display_name(app),
        g_app_info_get_name(app),
        g_app_info_get_executable(app),
        g_app_info_get_description(app),
        g_app_info_get_id(app),
    };
    for (size_t i = 0; i < sizeof(fields) / sizeof(fields[0]); i++) {
        if (!fields[i] || !fields[i][0]) continue;
        char *folded = g_utf8_casefold(fields[i], -1);
        int matches = folded && strstr(folded, query_folded) != NULL;
        g_free(folded);
        if (matches) return 1;
    }
    return 0;
}

static void launch_launcher_app(HatWMPanel *panel, GAppInfo *app) {
    if (!app) return;

    GError *error = NULL;
    if (!g_app_info_launch(app, NULL, NULL, &error)) {
        fprintf(stderr, "hatwmpanel: could not launch %s: %s\n",
                g_app_info_get_display_name(app),
                error ? error->message : "unknown error");
    }
    if (error) g_error_free(error);
    if (panel->launcher_window)
        gtk_widget_destroy(panel->launcher_window);
}

static void launcher_result_activated(GtkListBox *list, GtkListBoxRow *row,
                                      gpointer data) {
    (void)list;
    HatWMPanel *panel = data;
    if (!panel || !row) return;
    GAppInfo *app = g_object_get_data(
        G_OBJECT(row), "hatwmpanel-launcher-app");
    launch_launcher_app(panel, app);
}

static GtkWidget *launcher_result_row(GAppInfo *app) {
    GtkWidget *row = gtk_list_box_row_new();
    GtkWidget *box = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 10);
    GIcon *icon = g_app_info_get_icon(app);
    GtkWidget *image = icon
        ? gtk_image_new_from_gicon(icon, GTK_ICON_SIZE_DIALOG)
        : gtk_image_new_from_icon_name(
            "application-x-executable", GTK_ICON_SIZE_DIALOG);
    gtk_image_set_pixel_size(GTK_IMAGE(image), 28);
    GtkWidget *label = gtk_label_new(g_app_info_get_display_name(app));
    gtk_label_set_xalign(GTK_LABEL(label), 0.0f);
    gtk_label_set_ellipsize(GTK_LABEL(label), PANGO_ELLIPSIZE_END);
    gtk_widget_set_hexpand(label, TRUE);
    gtk_box_pack_start(GTK_BOX(box), image, FALSE, FALSE, 0);
    gtk_box_pack_start(GTK_BOX(box), label, TRUE, TRUE, 0);
    gtk_container_add(GTK_CONTAINER(row), box);
    g_object_set_data_full(
        G_OBJECT(row), "hatwmpanel-launcher-app",
        g_object_ref(app), g_object_unref);
    return row;
}

static void launcher_search_changed(GtkEditable *editable, gpointer data) {
    HatWMPanel *panel = data;
    if (!panel || !panel->launcher_results) return;
    GtkWidget *list = panel->launcher_list;
    if (!list) return;
    gtk_container_foreach(
        GTK_CONTAINER(list), destroy_launcher_result, NULL);

    const char *text = gtk_entry_get_text(GTK_ENTRY(editable));
    if (panel->launcher_hint) {
        if (text && text[0])
            gtk_widget_hide(panel->launcher_hint);
        else
            gtk_widget_show(panel->launcher_hint);
    }
    char *trimmed = g_strdup(text ? text : "");
    g_strstrip(trimmed);
    if (!trimmed[0]) {
        gtk_widget_hide(panel->launcher_results);
        g_free(trimmed);
        return;
    }

    char *query_folded = g_utf8_casefold(trimmed, -1);
    int result_count = 0;
    for (GList *item = panel->launcher_apps;
         item && result_count < 8; item = item->next) {
        GAppInfo *app = G_APP_INFO(item->data);
        if (!launcher_app_matches(app, query_folded)) continue;
        gtk_list_box_insert(
            GTK_LIST_BOX(list), launcher_result_row(app), -1);
        result_count++;
    }
    if (result_count == 0) {
        GtkWidget *row = gtk_list_box_row_new();
        GtkWidget *label = gtk_label_new("No applications found");
        gtk_label_set_xalign(GTK_LABEL(label), 0.0f);
        gtk_widget_set_sensitive(row, FALSE);
        gtk_container_add(GTK_CONTAINER(row), label);
        gtk_list_box_insert(GTK_LIST_BOX(list), row, -1);
    } else {
        GtkListBoxRow *first =
            gtk_list_box_get_row_at_index(GTK_LIST_BOX(list), 0);
        gtk_list_box_select_row(GTK_LIST_BOX(list), first);
    }
    g_free(query_folded);
    g_free(trimmed);
    gtk_widget_show_all(panel->launcher_results);
}

static void launcher_command_activated(GtkEntry *entry, gpointer data) {
    HatWMPanel *panel = data;
    if (!panel) return;
    if (panel->launcher_list) {
        GtkListBoxRow *selected = gtk_list_box_get_selected_row(
            GTK_LIST_BOX(panel->launcher_list));
        GAppInfo *app = selected
            ? g_object_get_data(
                G_OBJECT(selected), "hatwmpanel-launcher-app")
            : NULL;
        if (app) {
            launch_launcher_app(panel, app);
            return;
        }
    }
    const char *text = gtk_entry_get_text(entry);
    char *command = g_strdup(text ? text : "");
    g_strstrip(command);
    if (!command[0]) {
        g_free(command);
        return;
    }

    char *arguments[] = {"/bin/sh", "-c", command, NULL};
    GError *error = NULL;
    if (!g_spawn_async(NULL, arguments, NULL, G_SPAWN_SEARCH_PATH,
                       NULL, NULL, NULL, &error)) {
        fprintf(stderr, "hatwmpanel: could not run command: %s\n",
                error ? error->message : "unknown error");
    }
    if (error) g_error_free(error);
    g_free(command);
    if (panel->launcher_window)
        gtk_widget_destroy(panel->launcher_window);
}

static gboolean launcher_key_pressed(GtkWidget *widget, GdkEventKey *event,
                                     gpointer data) {
    (void)widget;
    HatWMPanel *panel = data;
    if (panel && panel->launcher_list && event &&
        (event->keyval == GDK_KEY_Down || event->keyval == GDK_KEY_Up)) {
        GtkListBox *list = GTK_LIST_BOX(panel->launcher_list);
        GtkListBoxRow *selected = gtk_list_box_get_selected_row(list);
        int index = selected ? gtk_list_box_row_get_index(selected) : -1;
        index += event->keyval == GDK_KEY_Down ? 1 : -1;
        if (index < 0) index = 0;
        GtkListBoxRow *next = gtk_list_box_get_row_at_index(list, index);
        if (!next && selected)
            next = selected;
        if (next && gtk_widget_get_sensitive(GTK_WIDGET(next))) {
            gtk_list_box_select_row(list, next);
            GtkAdjustment *adjustment =
                gtk_scrolled_window_get_vadjustment(
                    GTK_SCROLLED_WINDOW(panel->launcher_results));
            GtkAllocation allocation;
            gtk_widget_get_allocation(GTK_WIDGET(next), &allocation);
            double value = gtk_adjustment_get_value(adjustment);
            double page = gtk_adjustment_get_page_size(adjustment);
            if (allocation.y < value)
                gtk_adjustment_set_value(adjustment, allocation.y);
            else if (allocation.y + allocation.height > value + page)
                gtk_adjustment_set_value(
                    adjustment, allocation.y + allocation.height - page);
        }
        return TRUE;
    }
    if (event && event->keyval == GDK_KEY_Escape) {
        if (panel && panel->launcher_window)
            gtk_widget_destroy(panel->launcher_window);
        return TRUE;
    }
    return FALSE;
}

static void launcher_window_destroyed(GtkWidget *widget, gpointer data) {
    HatWMPanel *panel = data;
    if (!panel || panel->launcher_window != widget) return;
    panel->launcher_window = NULL;
    panel->launcher_entry = NULL;
    panel->launcher_hint = NULL;
    panel->launcher_results = NULL;
    panel->launcher_list = NULL;
    panel->launcher_left_margin = -1;
    g_list_free_full(panel->launcher_apps, g_object_unref);
    panel->launcher_apps = NULL;
    if (active_tray_menus > 0) active_tray_menus--;
}

static void position_launcher_window(GtkWidget *widget,
                                     GtkAllocation *allocation,
                                     gpointer data) {
    HatWMPanel *panel = data;
    if (!panel || panel->launcher_window != widget ||
        allocation->width <= 0)
        return;
    int left_margin = panel->launcher_origin_x;
    if (left_margin < 0) left_margin = 0;
    if (left_margin + allocation->width > panel->width)
        left_margin = panel->width - allocation->width;
    if (left_margin < 0) left_margin = 0;
    if (left_margin == panel->launcher_left_margin) return;
    panel->launcher_left_margin = left_margin;
    gtk_layer_set_margin(
        GTK_WINDOW(widget), GTK_LAYER_SHELL_EDGE_LEFT, left_margin);
}

int hatwm_panel_show_launcher(HatWMPanel *panel) {
    if (!panel || panel->launcher_origin_x < 0) return 0;
    if (panel->launcher_window) {
        gtk_widget_destroy(panel->launcher_window);
        return 1;
    }
    if (!gtk_initialized) {
        if (!gtk_init_check(NULL, NULL)) return 0;
        gtk_initialized = 1;
    }

    GtkWidget *window = gtk_window_new(GTK_WINDOW_TOPLEVEL);
    GtkWidget *box = gtk_box_new(GTK_ORIENTATION_VERTICAL, 6);
    GtkWidget *entry_overlay = gtk_overlay_new();
    GtkWidget *entry = gtk_search_entry_new();
    GtkWidget *hint =
        gtk_label_new("Search applications or run a command");
    GtkWidget *scrolled = gtk_scrolled_window_new(NULL, NULL);
    GtkWidget *list = gtk_list_box_new();
    if (!window || !box || !entry_overlay || !entry || !hint ||
        !scrolled || !list) {
        if (window) gtk_widget_destroy(window);
        return 0;
    }

    panel->launcher_window = window;
    panel->launcher_entry = entry;
    panel->launcher_hint = hint;
    panel->launcher_results = scrolled;
    panel->launcher_list = list;
    panel->launcher_left_margin = -1;
    panel->launcher_apps =
        g_list_sort(g_app_info_get_all(), compare_launcher_apps);
    active_tray_menus++;

    gtk_widget_set_name(window, "hatwm-launcher-window");
    gtk_widget_set_name(box, "hatwm-launcher-box");
    gtk_widget_set_name(entry, "hatwm-launcher-entry");
    gtk_widget_set_name(hint, "hatwm-launcher-hint");
    gtk_widget_set_name(list, "hatwm-launcher-results");
    gtk_widget_set_halign(hint, GTK_ALIGN_START);
    gtk_widget_set_valign(hint, GTK_ALIGN_CENTER);
    gtk_widget_set_margin_start(hint, 34);
    gtk_widget_set_opacity(hint, 0.58);
    gtk_window_set_decorated(GTK_WINDOW(window), FALSE);
    gtk_window_set_resizable(GTK_WINDOW(window), FALSE);
    gtk_window_set_skip_taskbar_hint(GTK_WINDOW(window), TRUE);
    gtk_widget_set_size_request(box, 380, -1);
    gtk_scrolled_window_set_policy(
        GTK_SCROLLED_WINDOW(scrolled),
        GTK_POLICY_NEVER, GTK_POLICY_AUTOMATIC);
    gtk_scrolled_window_set_propagate_natural_height(
        GTK_SCROLLED_WINDOW(scrolled), TRUE);
    gtk_scrolled_window_set_max_content_height(
        GTK_SCROLLED_WINDOW(scrolled), 320);
    gtk_list_box_set_selection_mode(
        GTK_LIST_BOX(list), GTK_SELECTION_SINGLE);
    gtk_container_add(GTK_CONTAINER(entry_overlay), entry);
    gtk_overlay_add_overlay(GTK_OVERLAY(entry_overlay), hint);
    gtk_overlay_set_overlay_pass_through(
        GTK_OVERLAY(entry_overlay), hint, TRUE);
    gtk_container_add(GTK_CONTAINER(scrolled), list);
    gtk_box_pack_start(GTK_BOX(box), entry_overlay, FALSE, FALSE, 0);
    gtk_box_pack_start(GTK_BOX(box), scrolled, TRUE, TRUE, 0);
    gtk_container_add(GTK_CONTAINER(window), box);

    gtk_layer_init_for_window(GTK_WINDOW(window));
    gtk_layer_set_namespace(GTK_WINDOW(window), "hatwmpanel-launcher");
    gtk_layer_set_layer(
        GTK_WINDOW(window), GTK_LAYER_SHELL_LAYER_OVERLAY);
    gtk_layer_set_keyboard_mode(
        GTK_WINDOW(window), GTK_LAYER_SHELL_KEYBOARD_MODE_EXCLUSIVE);
    gtk_layer_set_anchor(
        GTK_WINDOW(window), GTK_LAYER_SHELL_EDGE_LEFT, TRUE);
    gtk_layer_set_margin(
        GTK_WINDOW(window), GTK_LAYER_SHELL_EDGE_LEFT,
        panel->launcher_origin_x);
    if (panel->panel_at_bottom) {
        gtk_layer_set_anchor(
            GTK_WINDOW(window), GTK_LAYER_SHELL_EDGE_BOTTOM, TRUE);
        gtk_layer_set_margin(
            GTK_WINDOW(window), GTK_LAYER_SHELL_EDGE_BOTTOM, panel->height);
    } else {
        gtk_layer_set_anchor(
            GTK_WINDOW(window), GTK_LAYER_SHELL_EDGE_TOP, TRUE);
        gtk_layer_set_margin(
            GTK_WINDOW(window), GTK_LAYER_SHELL_EDGE_TOP, panel->height);
    }

    g_signal_connect(entry, "changed",
                     G_CALLBACK(launcher_search_changed), panel);
    g_signal_connect(entry, "activate",
                     G_CALLBACK(launcher_command_activated), panel);
    g_signal_connect(entry, "key-press-event",
                     G_CALLBACK(launcher_key_pressed), panel);
    g_signal_connect(list, "row-activated",
                     G_CALLBACK(launcher_result_activated), panel);
    g_signal_connect(window, "key-press-event",
                     G_CALLBACK(launcher_key_pressed), panel);
    g_signal_connect(window, "destroy",
                     G_CALLBACK(launcher_window_destroyed), panel);
    g_signal_connect(window, "size-allocate",
                     G_CALLBACK(position_launcher_window), panel);

    gtk_widget_show_all(window);
    gtk_widget_hide(scrolled);
    gtk_window_present(GTK_WINDOW(window));
    gtk_widget_grab_focus(entry);
    return 1;
}

