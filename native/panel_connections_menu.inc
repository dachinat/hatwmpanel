static void destroy_network_menu_choice(gpointer data) {
    NetworkMenuChoice *choice = data;
    if (!choice) return;
    free(choice->ssid);
    free(choice);
}

static void network_menu_item_activated(GtkMenuItem *item, gpointer data) {
    (void)item;
    NetworkMenuChoice *choice = data;
    if (!choice || !choice->panel) return;
    for (int i = 0; i < choice->panel->network_count; i++) {
        const char *ssid = choice->panel->networks[i].ssid;
        if (ssid && strcmp(ssid, choice->ssid) == 0 &&
            choice->panel->networks[i].wired == choice->wired) {
            choice->panel->pending_network_click = i + 1;
            return;
        }
    }
}

static void network_disconnect_activated(GtkMenuItem *item, gpointer data) {
    (void)item;
    HatWMPanel *panel = data;
    if (panel) panel->pending_network_disconnect = 1;
}

int hatwm_panel_show_network_menu(HatWMPanel *panel) {
    if (!panel || panel->network_origin_x < 0) return 0;
    if (panel->network_menu) {
        gtk_widget_destroy(panel->network_menu);
        return 1;
    }
    if (!gtk_initialized) {
        if (!gtk_init_check(NULL, NULL)) return 0;
        gtk_initialized = 1;
    }

    GtkWidget *menu = gtk_menu_new();
    if (!menu) return 0;
    int has_active_network = 0;
    for (int i = 0; i < panel->network_count; i++)
        if (panel->networks[i].active) has_active_network = 1;

    if (panel->network_count == 0) {
        GtkWidget *empty = gtk_menu_item_new_with_label("No networks found");
        gtk_widget_set_sensitive(empty, FALSE);
        gtk_menu_shell_append(GTK_MENU_SHELL(menu), empty);
    } else {
        for (int i = 0; i < panel->network_count; i++) {
            PanelNetwork *network = &panel->networks[i];
            gchar *name_text = g_strdup_printf(
                "%s%s",
                network->active ? "✓  " : "",
                network->ssid && network->ssid[0] ? network->ssid : "(hidden)");
            GtkWidget *item = gtk_menu_item_new_with_label(name_text);
            g_free(name_text);

            NetworkMenuChoice *choice = calloc(1, sizeof(*choice));
            if (!choice) {
                gtk_widget_destroy(item);
                continue;
            }
            choice->panel = panel;
            choice->ssid = strdup(network->ssid ? network->ssid : "");
            choice->wired = network->wired;
            if (!choice->ssid) {
                free(choice);
                gtk_widget_destroy(item);
                continue;
            }
            g_object_set_data_full(
                G_OBJECT(item), "hatwmpanel-network-choice", choice,
                destroy_network_menu_choice);
            g_signal_connect(item, "activate",
                             G_CALLBACK(network_menu_item_activated), choice);
            gtk_menu_shell_append(GTK_MENU_SHELL(menu), item);
        }
    }
    if (has_active_network) {
        GtkWidget *separator = gtk_separator_menu_item_new();
        gtk_menu_shell_append(GTK_MENU_SHELL(menu), separator);

        GtkWidget *disconnect = gtk_menu_item_new_with_label("Disconnect");
        g_signal_connect(disconnect, "activate",
                         G_CALLBACK(network_disconnect_activated), panel);
        gtk_menu_shell_append(GTK_MENU_SHELL(menu), disconnect);
    }

    PendingMenu *pending = calloc(1, sizeof(*pending));
    if (!pending) {
        gtk_widget_destroy(menu);
        return 0;
    }
    pending->menu = g_object_ref_sink(menu);
    pending->network_panel = panel;
    pending->panel_at_bottom = panel->panel_at_bottom;
    pending->align_center = 1;
    int network_center = panel->network_origin_x + panel->network_width / 2;
    pending->align_right = 0;

    pending->anchor = gtk_window_new(GTK_WINDOW_TOPLEVEL);
    gtk_window_set_decorated(GTK_WINDOW(pending->anchor), FALSE);
    gtk_window_set_skip_taskbar_hint(GTK_WINDOW(pending->anchor), TRUE);
    gtk_widget_set_opacity(pending->anchor, 0.0);
    gtk_widget_set_size_request(pending->anchor, 1, 1);
    gtk_layer_init_for_window(GTK_WINDOW(pending->anchor));
    gtk_layer_set_namespace(GTK_WINDOW(pending->anchor),
                            "hatwmpanel-network-menu-anchor");
    gtk_layer_set_layer(GTK_WINDOW(pending->anchor),
                        GTK_LAYER_SHELL_LAYER_OVERLAY);
    gtk_layer_set_keyboard_mode(GTK_WINDOW(pending->anchor),
                                GTK_LAYER_SHELL_KEYBOARD_MODE_NONE);
    gtk_layer_set_anchor(GTK_WINDOW(pending->anchor),
                         GTK_LAYER_SHELL_EDGE_LEFT, TRUE);
    gtk_layer_set_margin(GTK_WINDOW(pending->anchor),
                         GTK_LAYER_SHELL_EDGE_LEFT, network_center);
    if (panel->panel_at_bottom) {
        gtk_layer_set_anchor(GTK_WINDOW(pending->anchor),
                             GTK_LAYER_SHELL_EDGE_BOTTOM, TRUE);
        gtk_layer_set_margin(GTK_WINDOW(pending->anchor),
                             GTK_LAYER_SHELL_EDGE_BOTTOM, panel->height);
    } else {
        gtk_layer_set_anchor(GTK_WINDOW(pending->anchor),
                             GTK_LAYER_SHELL_EDGE_TOP, TRUE);
        gtk_layer_set_margin(GTK_WINDOW(pending->anchor),
                             GTK_LAYER_SHELL_EDGE_TOP, panel->height);
    }

    panel->network_menu = menu;
    gtk_widget_show(pending->anchor);
    gtk_menu_attach_to_widget(GTK_MENU(menu), pending->anchor, NULL);
    g_signal_connect(menu, "deactivate",
                     G_CALLBACK(pending_menu_deactivated), pending);
    g_signal_connect(menu, "destroy", G_CALLBACK(destroy_pending_menu), pending);
    active_tray_menus++;
    pending->popup_source = g_timeout_add(20, popup_menu_when_ready, pending);
    return 1;
}

typedef struct BluetoothMenuChoice {
    HatWMPanel *panel;
    char *address;
} BluetoothMenuChoice;

static void destroy_bluetooth_menu_choice(gpointer data) {
    BluetoothMenuChoice *choice = data;
    if (!choice) return;
    free(choice->address);
    free(choice);
}

static void bluetooth_menu_item_activated(GtkMenuItem *item, gpointer data) {
    (void)item;
    BluetoothMenuChoice *choice = data;
    if (!choice || !choice->panel) return;
    for (int i = 0; i < choice->panel->bluetooth_device_count; i++) {
        const char *address = choice->panel->bluetooth_devices[i].address;
        if (address && strcmp(address, choice->address) == 0) {
            choice->panel->pending_bluetooth_click = i + 1;
            return;
        }
    }
}

static void bluetooth_disconnect_activated(GtkMenuItem *item, gpointer data) {
    (void)item;
    HatWMPanel *panel = data;
    if (panel) panel->pending_bluetooth_disconnect = 1;
}

int hatwm_panel_show_bluetooth_menu(HatWMPanel *panel) {
    if (!panel || panel->bluetooth_origin_x < 0) return 0;
    if (panel->bluetooth_menu) {
        gtk_widget_destroy(panel->bluetooth_menu);
        return 1;
    }
    if (!gtk_initialized) {
        if (!gtk_init_check(NULL, NULL)) return 0;
        gtk_initialized = 1;
    }

    GtkWidget *menu = gtk_menu_new();
    if (!menu) return 0;
    int has_connected_device = 0;
    for (int i = 0; i < panel->bluetooth_device_count; i++)
        if (panel->bluetooth_devices[i].connected) has_connected_device = 1;

    if (panel->bluetooth_device_count == 0) {
        GtkWidget *empty = gtk_menu_item_new_with_label("No Bluetooth devices found");
        gtk_widget_set_sensitive(empty, FALSE);
        gtk_menu_shell_append(GTK_MENU_SHELL(menu), empty);
    } else {
        for (int i = 0; i < panel->bluetooth_device_count; i++) {
            PanelBluetoothDevice *device = &panel->bluetooth_devices[i];
            gchar *name_text = g_strdup_printf(
                "%s%s",
                device->connected ? "✓  " : "",
                device->name && device->name[0] ? device->name : device->address);
            GtkWidget *item = gtk_menu_item_new_with_label(name_text);
            g_free(name_text);

            BluetoothMenuChoice *choice = calloc(1, sizeof(*choice));
            if (!choice) {
                gtk_widget_destroy(item);
                continue;
            }
            choice->panel = panel;
            choice->address = strdup(device->address ? device->address : "");
            if (!choice->address) {
                free(choice);
                gtk_widget_destroy(item);
                continue;
            }
            g_object_set_data_full(
                G_OBJECT(item), "hatwmpanel-bluetooth-choice", choice,
                destroy_bluetooth_menu_choice);
            g_signal_connect(item, "activate",
                             G_CALLBACK(bluetooth_menu_item_activated), choice);
            gtk_menu_shell_append(GTK_MENU_SHELL(menu), item);
        }
    }
    if (has_connected_device) {
        GtkWidget *separator = gtk_separator_menu_item_new();
        gtk_menu_shell_append(GTK_MENU_SHELL(menu), separator);
        GtkWidget *disconnect = gtk_menu_item_new_with_label("Disconnect");
        g_signal_connect(disconnect, "activate",
                         G_CALLBACK(bluetooth_disconnect_activated), panel);
        gtk_menu_shell_append(GTK_MENU_SHELL(menu), disconnect);
    }

    PendingMenu *pending = calloc(1, sizeof(*pending));
    if (!pending) {
        gtk_widget_destroy(menu);
        return 0;
    }
    pending->menu = g_object_ref_sink(menu);
    pending->bluetooth_panel = panel;
    pending->panel_at_bottom = panel->panel_at_bottom;
    pending->align_center = 1;
    int module_center =
        panel->bluetooth_origin_x + panel->bluetooth_width / 2;

    pending->anchor = gtk_window_new(GTK_WINDOW_TOPLEVEL);
    gtk_window_set_decorated(GTK_WINDOW(pending->anchor), FALSE);
    gtk_window_set_skip_taskbar_hint(GTK_WINDOW(pending->anchor), TRUE);
    gtk_widget_set_opacity(pending->anchor, 0.0);
    gtk_widget_set_size_request(pending->anchor, 1, 1);
    gtk_layer_init_for_window(GTK_WINDOW(pending->anchor));
    gtk_layer_set_namespace(GTK_WINDOW(pending->anchor),
                            "hatwmpanel-bluetooth-menu-anchor");
    gtk_layer_set_layer(GTK_WINDOW(pending->anchor),
                        GTK_LAYER_SHELL_LAYER_OVERLAY);
    gtk_layer_set_keyboard_mode(GTK_WINDOW(pending->anchor),
                                GTK_LAYER_SHELL_KEYBOARD_MODE_NONE);
    gtk_layer_set_anchor(GTK_WINDOW(pending->anchor),
                         GTK_LAYER_SHELL_EDGE_LEFT, TRUE);
    gtk_layer_set_margin(GTK_WINDOW(pending->anchor),
                         GTK_LAYER_SHELL_EDGE_LEFT, module_center);
    if (panel->panel_at_bottom) {
        gtk_layer_set_anchor(GTK_WINDOW(pending->anchor),
                             GTK_LAYER_SHELL_EDGE_BOTTOM, TRUE);
        gtk_layer_set_margin(GTK_WINDOW(pending->anchor),
                             GTK_LAYER_SHELL_EDGE_BOTTOM, panel->height);
    } else {
        gtk_layer_set_anchor(GTK_WINDOW(pending->anchor),
                             GTK_LAYER_SHELL_EDGE_TOP, TRUE);
        gtk_layer_set_margin(GTK_WINDOW(pending->anchor),
                             GTK_LAYER_SHELL_EDGE_TOP, panel->height);
    }

    panel->bluetooth_menu = menu;
    gtk_widget_show(pending->anchor);
    gtk_menu_attach_to_widget(GTK_MENU(menu), pending->anchor, NULL);
    g_signal_connect(menu, "deactivate",
                     G_CALLBACK(pending_menu_deactivated), pending);
    g_signal_connect(menu, "destroy", G_CALLBACK(destroy_pending_menu), pending);
    active_tray_menus++;
    pending->popup_source = g_timeout_add(20, popup_menu_when_ready, pending);
    return 1;
}

