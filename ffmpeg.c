#include "ffmpeg.h"
#include <libavutil/dict.h>
#include <libavutil/log.h>
#include <stdlib.h>
#include <string.h>

void init_ffmpeg(void) {
    av_log_set_level(AV_LOG_QUIET);
}

VideoMetadata* get_metadata(const char* filepath) {
    AVFormatContext *fmt_ctx = NULL;
    if (avformat_open_input(&fmt_ctx, filepath, NULL, NULL) < 0) {
        return NULL;
    }
    if (avformat_find_stream_info(fmt_ctx, NULL) < 0) {
        avformat_close_input(&fmt_ctx);
        return NULL;
    }

    VideoMetadata *meta = calloc(1, sizeof(VideoMetadata));
    meta->duration = fmt_ctx->duration;

    AVDictionaryEntry *tag = av_dict_get(fmt_ctx->metadata, "creation_time", NULL, AV_DICT_IGNORE_SUFFIX);
    if (tag) {
        meta->creation_time = strdup(tag->value);
    }

    for (int i = 0; i < fmt_ctx->nb_streams; i++) {
        AVStream *st = fmt_ctx->streams[i];
        if (st->codecpar->codec_type == AVMEDIA_TYPE_VIDEO) {
            meta->has_video = 1;
            meta->width = st->codecpar->width;
            meta->height = st->codecpar->height;
            if (st->avg_frame_rate.den > 0) {
                meta->fps = av_q2d(st->avg_frame_rate);
            } else if (st->r_frame_rate.den > 0) {
                meta->fps = av_q2d(st->r_frame_rate);
            }
            break;
        }
    }

    avformat_close_input(&fmt_ctx);
    return meta;
}

struct ConcatState {
    AVFormatContext *ofmt_ctx;
    int64_t pts_offset[10];
    int64_t dts_offset[10];
    int64_t last_pts[10];
    int64_t last_dts[10];
    int stream_mapping[10];
    int stream_count;
};

ConcatState* init_output(const char* first_input, const char* output_filename) {
    ConcatState *state = calloc(1, sizeof(ConcatState));
    
    AVFormatContext *ifmt_ctx = NULL;
    int ret = avformat_open_input(&ifmt_ctx, first_input, NULL, NULL);
    if (ret < 0) {
        char errbuf[256];
        av_strerror(ret, errbuf, sizeof(errbuf));
        printf("DEBUG: avformat_open_input failed for %s: %s\n", first_input, errbuf);
        free(state);
        return NULL;
    }
    ret = avformat_find_stream_info(ifmt_ctx, NULL);
    if (ret < 0) {
        char errbuf[256];
        av_strerror(ret, errbuf, sizeof(errbuf));
        printf("DEBUG: avformat_find_stream_info failed: %s\n", errbuf);
        avformat_close_input(&ifmt_ctx);
        free(state);
        return NULL;
    }

    ret = avformat_alloc_output_context2(&state->ofmt_ctx, NULL, NULL, output_filename);
    if (!state->ofmt_ctx) {
        printf("DEBUG: avformat_alloc_output_context2 failed to deduce output format for %s\n", output_filename);
        avformat_close_input(&ifmt_ctx);
        free(state);
        return NULL;
    }

    state->stream_count = ifmt_ctx->nb_streams;
    if (state->stream_count > 10) state->stream_count = 10;

    for (int i = 0; i < state->stream_count; i++) {
        state->stream_mapping[i] = -1; // Default to skipped
        AVStream *in_stream = ifmt_ctx->streams[i];
        
        // Skip streams that are not video or audio, or have unknown codecs (like iPhone 'apac' Spatial Audio)
        if ((in_stream->codecpar->codec_type != AVMEDIA_TYPE_VIDEO && 
             in_stream->codecpar->codec_type != AVMEDIA_TYPE_AUDIO) ||
             in_stream->codecpar->codec_id == AV_CODEC_ID_NONE) {
            continue;
        }

        AVStream *out_stream = avformat_new_stream(state->ofmt_ctx, NULL);
        if (!out_stream) {
            printf("DEBUG: avformat_new_stream failed\n");
            return NULL;
        }

        ret = avcodec_parameters_copy(out_stream->codecpar, in_stream->codecpar);
        if (ret < 0) {
            char errbuf[256];
            av_strerror(ret, errbuf, sizeof(errbuf));
            printf("DEBUG: avcodec_parameters_copy failed: %s\n", errbuf);
            return NULL;
        }
        
        // We explicitly keep the original codec_tag (e.g. 'hvc1' for iPhone HEVC).
        // If we zero it out, FFmpeg defaults to 'hev1' which Apple devices refuse to play.
        state->stream_mapping[i] = out_stream->index;
    }

    if (!(state->ofmt_ctx->oformat->flags & AVFMT_NOFILE)) {
        ret = avio_open(&state->ofmt_ctx->pb, output_filename, AVIO_FLAG_WRITE);
        if (ret < 0) {
            char errbuf[256];
            av_strerror(ret, errbuf, sizeof(errbuf));
            printf("DEBUG: avio_open failed for %s: %s\n", output_filename, errbuf);
            return NULL;
        }
    }

    AVDictionary *opt = NULL;
    av_dict_set(&opt, "strict", "unofficial", 0);
    
    ret = avformat_write_header(state->ofmt_ctx, &opt);
    av_dict_free(&opt);
    
    if (ret < 0) {
        char errbuf[256];
        av_strerror(ret, errbuf, sizeof(errbuf));
        printf("DEBUG: avformat_write_header failed: %s\n", errbuf);
        return NULL;
    }

    avformat_close_input(&ifmt_ctx);
    return state;
}

extern void progressCallback(int64_t current_time_us);

int append_file(ConcatState *state, const char* filepath) {
    AVFormatContext *ifmt_ctx = NULL;
    if (avformat_open_input(&ifmt_ctx, filepath, NULL, NULL) < 0) return -1;
    if (avformat_find_stream_info(ifmt_ctx, NULL) < 0) {
        avformat_close_input(&ifmt_ctx);
        return -1;
    }

    AVPacket *pkt = av_packet_alloc();
    if (!pkt) return -1;

    int64_t max_pts[10] = {0};
    int64_t max_dts[10] = {0};

    while (1) {
        int ret = av_read_frame(ifmt_ctx, pkt);
        if (ret < 0) break;

        if (pkt->stream_index >= state->stream_count) {
            av_packet_unref(pkt);
            continue;
        }
        
        int stream_index = state->stream_mapping[pkt->stream_index];
        if (stream_index < 0) {
            av_packet_unref(pkt);
            continue;
        }

        AVStream *in_stream  = ifmt_ctx->streams[pkt->stream_index];
        AVStream *out_stream = state->ofmt_ctx->streams[stream_index];

        pkt->pts = av_rescale_q_rnd(pkt->pts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
        pkt->dts = av_rescale_q_rnd(pkt->dts, in_stream->time_base, out_stream->time_base, AV_ROUND_NEAR_INF|AV_ROUND_PASS_MINMAX);
        pkt->duration = av_rescale_q(pkt->duration, in_stream->time_base, out_stream->time_base);
        
        pkt->pts += state->pts_offset[stream_index];
        pkt->dts += state->dts_offset[stream_index];
        pkt->pos = -1;

        if (pkt->pts > max_pts[stream_index]) max_pts[stream_index] = pkt->pts;
        if (pkt->dts > max_dts[stream_index]) max_dts[stream_index] = pkt->dts;

        if (stream_index == 0) {
            int64_t current_time_us = av_rescale_q(pkt->pts, out_stream->time_base, AV_TIME_BASE_Q);
            progressCallback(current_time_us);
        }

        ret = av_interleaved_write_frame(state->ofmt_ctx, pkt);
        av_packet_unref(pkt);
    }

    av_packet_free(&pkt);

    for (int i = 0; i < state->stream_count; i++) {
        state->pts_offset[i] = max_pts[i] + 1;
        state->dts_offset[i] = max_dts[i] + 1;
    }

    avformat_close_input(&ifmt_ctx);
    return 0;
}

void finalize_output(ConcatState *state) {
    if (!state || !state->ofmt_ctx) return;
    av_write_trailer(state->ofmt_ctx);
    if (!(state->ofmt_ctx->oformat->flags & AVFMT_NOFILE)) {
        avio_closep(&state->ofmt_ctx->pb);
    }
    avformat_free_context(state->ofmt_ctx);
    free(state);
}
