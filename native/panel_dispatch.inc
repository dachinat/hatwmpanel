int hatwm_panel_dispatch(HatWMPanel *panel, int timeout_ms) {
    if (!panel || !panel->display || panel->closed) return 0;
    if (gtk_initialized)
        while (gtk_events_pending()) gtk_main_iteration_do(FALSE);
    /* GTK uses its own Wayland connection. While a menu is visible, do not
     * leave that connection and D-Bus menu updates waiting behind the panel's
     * normal 200 ms idle poll. */
    if (active_tray_menus > 0 && timeout_ms > 8) timeout_ms = 8;

    if (wl_display_dispatch_pending(panel->display) < 0) {
        set_error("wl_display_dispatch_pending failed");
        return -1;
    }
    if (wl_display_flush(panel->display) < 0 && errno != EAGAIN) {
        set_errno_error("wl_display_flush failed");
        return -1;
    }

    struct pollfd pfd = {
        .fd = wl_display_get_fd(panel->display),
        .events = POLLIN,
        .revents = 0,
    };
    int rc = poll(&pfd, 1, timeout_ms);
    if (rc < 0) {
        if (errno == EINTR) return 0;
        set_errno_error("poll failed");
        return -1;
    }
    if (rc == 0) return 0;
    if (pfd.revents & (POLLERR | POLLHUP | POLLNVAL)) {
        set_error("Wayland connection closed");
        return -1;
    }
    if (pfd.revents & POLLIN) {
        if (wl_display_dispatch(panel->display) < 0) {
            set_error("wl_display_dispatch failed");
            return -1;
        }
    }
    return 0;
}

int hatwm_panel_closed(HatWMPanel *panel) {
    return !panel || panel->closed;
}
