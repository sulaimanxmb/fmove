#ifndef FFMPEG_H
#define FFMPEG_H

#include <libavformat/avformat.h>
#include <stdint.h>

typedef struct ConcatState ConcatState;

typedef struct {
    int64_t duration;
    char* creation_time;
    int width;
    int height;
    double fps;
    int has_video;
} VideoMetadata;

void init_ffmpeg(void);
VideoMetadata* get_metadata(const char* filepath);
ConcatState* init_output(const char* first_input, const char* output_filename);
int append_file(ConcatState *state, const char* filepath);
void finalize_output(ConcatState *state);

#endif
