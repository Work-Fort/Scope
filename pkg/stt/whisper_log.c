#include <whisper.h>

// No-op log callback to suppress whisper.cpp's stderr output.
// Called via whisper_log_set() from transcriber.go init().
void whisper_noop_log(enum ggml_log_level level, const char *text, void *user_data) {
    (void)level;
    (void)text;
    (void)user_data;
}
