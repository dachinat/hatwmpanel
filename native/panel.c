/*
 * Single cgo translation unit assembled from focused implementation files.
 *
 * Keeping these as .inc files lets private static helpers and HatWMPanel's
 * internal representation remain private without turning panel.c into one
 * monolithic source file.
 */
#include "panel_internal.inc"
#include "panel_style.inc"
#include "panel_input.inc"
#include "panel_render.inc"
#include "panel_lifecycle.inc"
#include "panel_state.inc"
#include "panel_tray_menu.inc"
#include "panel_connections_menu.inc"
#include "panel_calendar.inc"
#include "panel_launcher.inc"
#include "panel_dispatch.inc"
