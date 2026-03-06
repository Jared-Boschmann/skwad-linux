/*
 * vte_impl.h — C shim for VTE terminal integration.
 *
 * To be implemented during the terminal integration phase.
 * Requires: vte-2.91, gtk+-3.0
 */
#pragma once

#include <stdint.h>

/* Opaque handle representing a single VTE terminal instance. */
typedef struct SkwadVTE SkwadVTE;

/* Create a new VTE terminal, spawn command, wire up callbacks.
 * go_ptr is passed back to Go callback exports. */
SkwadVTE* skwad_vte_new(const char* command, char* const* env, uintptr_t go_ptr);

/* Feed text to the child process stdin. */
void skwad_vte_feed_child(SkwadVTE* vte, const char* text, int len);

/* Set terminal character dimensions. */
void skwad_vte_set_size(SkwadVTE* vte, int cols, int rows);

/* Grab keyboard focus. */
void skwad_vte_grab_focus(SkwadVTE* vte);

/* Terminate the child process and destroy the widget. */
void skwad_vte_kill(SkwadVTE* vte);

/* Return the native GtkWidget* for embedding (as void*). */
void* skwad_vte_widget(SkwadVTE* vte);

/* Go callback declarations — implemented in vte.go via //export */
extern void goOnVTEOutput(uintptr_t ptr);
extern void goOnVTETitleChange(uintptr_t ptr, const char* title);
extern void goOnVTEKeyPress(uintptr_t ptr, unsigned int keyval);
