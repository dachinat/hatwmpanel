static void destroy_pending_menu(GtkWidget *menu, gpointer data) {
    PendingMenu *pending = data;
    if (pending->popup_source) g_source_remove(pending->popup_source);
    if (pending->destroy_source) g_source_remove(pending->destroy_source);
    if (pending->network_panel &&
        pending->network_panel->network_menu == menu)
        pending->network_panel->network_menu = NULL;
    if (pending->bluetooth_panel &&
        pending->bluetooth_panel->bluetooth_menu == menu)
        pending->bluetooth_panel->bluetooth_menu = NULL;
    if (pending->anchor) gtk_widget_destroy(pending->anchor);
    g_object_unref(pending->menu);
    if (active_tray_menus > 0) active_tray_menus--;
    free(pending);
}

static gboolean destroy_deactivated_menu(gpointer data) {
    PendingMenu *pending = data;
    pending->destroy_source = 0;
    gtk_widget_destroy(pending->menu);
    return G_SOURCE_REMOVE;
}

static void pending_menu_deactivated(GtkWidget *menu, gpointer data) {
    (void)menu;
    PendingMenu *pending = data;
    /*
     * GtkMenuItem's default activate handler can deactivate the menu before
     * application activate handlers run. Destroying it synchronously would
     * also free the network choice before its SSID is recorded.
     */
    if (!pending->destroy_source)
        pending->destroy_source = g_idle_add(destroy_deactivated_menu, pending);
}

/*
 * GTK normally chooses a submenu side from the menu window's global
 * coordinates. Wayland deliberately does not expose those coordinates, and
 * nested popups parented through gtk-layer-shell can consequently keep opening
 * to the right even at the output's right edge. Set the menu-item direction
 * explicitly, but restore LTR on item contents so labels and icons are not
 * mirrored.
 */
static void set_submenu_direction(GtkWidget *menu, gboolean open_left) {
    if (!GTK_IS_MENU(menu)) return;
    gtk_widget_set_direction(menu, open_left ? GTK_TEXT_DIR_RTL : GTK_TEXT_DIR_LTR);

    GList *items = gtk_container_get_children(GTK_CONTAINER(menu));
    for (GList *node = items; node; node = node->next) {
        GtkWidget *item = GTK_WIDGET(node->data);
        if (!GTK_IS_MENU_ITEM(item)) continue;
        gtk_widget_set_direction(item, open_left ? GTK_TEXT_DIR_RTL : GTK_TEXT_DIR_LTR);

        GtkWidget *content = gtk_bin_get_child(GTK_BIN(item));
        if (content) gtk_widget_set_direction(content, GTK_TEXT_DIR_LTR);

        GtkWidget *submenu = gtk_menu_item_get_submenu(GTK_MENU_ITEM(item));
        if (submenu) set_submenu_direction(submenu, open_left);
    }
    g_list_free(items);
}

static gboolean popup_menu_when_ready(gpointer data) {
    PendingMenu *pending = data;
    GList *children = gtk_container_get_children(GTK_CONTAINER(pending->menu));
    gboolean ready = children != NULL;
    g_list_free(children);
    if (!ready && ++pending->attempts < 20) return G_SOURCE_CONTINUE;
    pending->popup_source = 0;
    set_submenu_direction(pending->menu, pending->align_right);
    gtk_widget_show_all(pending->menu);

    /*
     * The pointer press arrives on the panel's separate wayland-client
     * connection, so GTK has no current GdkEvent. Associate the popup with its
     * mapped anchor window to keep GTK's popup/submenu machinery from warning
     * about a missing trigger event.
     */
    GdkEvent *trigger = gdk_event_new(GDK_BUTTON_PRESS);
    GdkWindow *anchor_window = gtk_widget_get_window(pending->anchor);
    if (!anchor_window) {
        gtk_widget_realize(pending->anchor);
        anchor_window = gtk_widget_get_window(pending->anchor);
    }
    if (!anchor_window) {
        gtk_widget_destroy(pending->menu);
        return G_SOURCE_REMOVE;
    }
    trigger->button.window = g_object_ref(anchor_window);
    trigger->button.send_event = TRUE;
    trigger->button.time = GDK_CURRENT_TIME;
    trigger->button.x = 0;
    trigger->button.y = 0;
    trigger->button.button = 3;
    GdkDisplay *display = gtk_widget_get_display(pending->anchor);
    GdkSeat *seat = gdk_display_get_default_seat(display);
    if (seat)
        gdk_event_set_device(trigger, gdk_seat_get_pointer(seat));

    if (pending->align_center && pending->panel_at_bottom) {
        gtk_menu_popup_at_widget(
            GTK_MENU(pending->menu), pending->anchor,
            GDK_GRAVITY_NORTH, GDK_GRAVITY_SOUTH, trigger);
    } else if (pending->align_center) {
        gtk_menu_popup_at_widget(
            GTK_MENU(pending->menu), pending->anchor,
            GDK_GRAVITY_SOUTH, GDK_GRAVITY_NORTH, trigger);
    } else if (pending->panel_at_bottom) {
        gtk_menu_popup_at_widget(
            GTK_MENU(pending->menu), pending->anchor,
            pending->align_right ? GDK_GRAVITY_NORTH_EAST : GDK_GRAVITY_NORTH_WEST,
            pending->align_right ? GDK_GRAVITY_SOUTH_EAST : GDK_GRAVITY_SOUTH_WEST,
            trigger);
    } else {
        gtk_menu_popup_at_widget(
            GTK_MENU(pending->menu), pending->anchor,
            pending->align_right ? GDK_GRAVITY_SOUTH_EAST : GDK_GRAVITY_SOUTH_WEST,
            pending->align_right ? GDK_GRAVITY_NORTH_EAST : GDK_GRAVITY_NORTH_WEST,
            trigger);
    }
    gdk_event_free(trigger);
    return G_SOURCE_REMOVE;
}

int hatwm_panel_show_tray_menu(HatWMPanel *panel, const char *service, const char *path) {
    if (!panel || !service || !service[0] || !path || !path[0]) return 0;
    if (!gtk_initialized) {
        if (!gtk_init_check(NULL, NULL)) return 0;
        gtk_initialized = 1;
    }
    GtkWidget *menu = GTK_WIDGET(dbusmenu_gtkmenu_new((gchar *)service, (gchar *)path));
    if (!menu) return 0;
    PendingMenu *pending = calloc(1, sizeof(*pending));
    if (!pending) {
        gtk_widget_destroy(menu);
        return 0;
    }
    pending->menu = g_object_ref_sink(menu);
    pending->panel_at_bottom = panel->panel_at_bottom;
    pending->align_right = panel->pointer_x > panel->width / 2;

    /*
     * GtkMenu needs a mapped GdkWindow to create an xdg_popup on Wayland.
     * The panel itself uses wayland-client directly, so give GTK a transparent
     * one-pixel layer-shell anchor at the tray icon instead of calling
     * gtk_menu_popup_at_pointer() without a GDK event/window (which is a no-op
     * on Wayland).
     */
    pending->anchor = gtk_window_new(GTK_WINDOW_TOPLEVEL);
    gtk_window_set_decorated(GTK_WINDOW(pending->anchor), FALSE);
    gtk_window_set_skip_taskbar_hint(GTK_WINDOW(pending->anchor), TRUE);
    gtk_widget_set_opacity(pending->anchor, 0.0);
    gtk_widget_set_size_request(pending->anchor, 1, 1);
    gtk_layer_init_for_window(GTK_WINDOW(pending->anchor));
    gtk_layer_set_namespace(GTK_WINDOW(pending->anchor), "hatwmpanel-menu-anchor");
    gtk_layer_set_layer(GTK_WINDOW(pending->anchor), GTK_LAYER_SHELL_LAYER_OVERLAY);
    gtk_layer_set_keyboard_mode(GTK_WINDOW(pending->anchor),
                                GTK_LAYER_SHELL_KEYBOARD_MODE_NONE);
    gtk_layer_set_anchor(GTK_WINDOW(pending->anchor), GTK_LAYER_SHELL_EDGE_LEFT, TRUE);
    gtk_layer_set_margin(GTK_WINDOW(pending->anchor), GTK_LAYER_SHELL_EDGE_LEFT,
                         (int)panel->pointer_x);
    if (panel->panel_at_bottom) {
        gtk_layer_set_anchor(GTK_WINDOW(pending->anchor), GTK_LAYER_SHELL_EDGE_BOTTOM, TRUE);
    } else {
        gtk_layer_set_anchor(GTK_WINDOW(pending->anchor), GTK_LAYER_SHELL_EDGE_TOP, TRUE);
    }
    gtk_widget_show(pending->anchor);
    gtk_menu_attach_to_widget(GTK_MENU(menu), pending->anchor, NULL);
    g_signal_connect(menu, "deactivate",
                     G_CALLBACK(pending_menu_deactivated), pending);
    g_signal_connect(menu, "destroy", G_CALLBACK(destroy_pending_menu), pending);
    active_tray_menus++;
    pending->popup_source = g_timeout_add(50, popup_menu_when_ready, pending);
    return 1;
}

typedef struct NetworkMenuChoice {
    HatWMPanel *panel;
    char *ssid;
    int wired;
} NetworkMenuChoice;
