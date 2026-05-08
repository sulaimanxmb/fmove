#ifndef FFMPEG_H
#define FFMPEG_H

#include <libavformat/avformat.h>
#include <stdint.h>

typedef struct ConcatState ConcatState;

void init_ffmpeg(void);
char* get_creation_time(const char* filepath);
int64_t get_duration(const char* filepath);
ConcatState* init_output(const char* first_input, const char* output_filename);
int append_file(ConcatState *state, const char* filepath);
void finalize_output(ConcatState *state);

#endif
